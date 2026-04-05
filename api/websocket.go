package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSClient represents a single WebSocket connection
type WSClient struct {
	Conn    *websocket.Conn
	Send    chan []byte
	AgentID string
	mu      sync.Mutex
}

// WSManager manages all WebSocket connections
type WSManager struct {
	Clients    map[*WSClient]bool
	Register   chan *WSClient
	Unregister chan *WSClient
	Broadcast  chan []byte
	mu         sync.RWMutex
}

var WS = &WSManager{
	Clients:    make(map[*WSClient]bool),
	Register:   make(chan *WSClient),
	Unregister: make(chan *WSClient),
	Broadcast:  make(chan []byte),
}

// Notification represents a real-time notification payload
type Notification struct {
	Type      string      `json:"type"`
	Title     string      `json:"title"`
	Message   string      `json:"message"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// Run starts the WebSocket manager (call in goroutine)
func (m *WSManager) Run() {
	for {
		select {
		case client := <-m.Register:
			m.mu.Lock()
			m.Clients[client] = true
			m.mu.Unlock()
			log.Printf("WS client connected: %s (total: %d)", client.AgentID, len(m.Clients))
		case client := <-m.Unregister:
			m.mu.Lock()
			if _, ok := m.Clients[client]; ok {
				delete(m.Clients, client)
				close(client.Send)
			}
			m.mu.Unlock()
			log.Printf("WS client disconnected: %s (total: %d)", client.AgentID, len(m.Clients))
		case message := <-m.Broadcast:
			m.mu.RLock()
			for client := range m.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(m.Clients, client)
				}
			}
			m.mu.RUnlock()
		}
	}
}

// SendToClient sends a message to a specific client by agent ID
func (m *WSManager) SendToClient(agentID string, data []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for client := range m.Clients {
		if client.AgentID == agentID {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

// BroadcastNotification sends a notification to all connected clients
func BroadcastNotification(n Notification) {
	data, _ := json.Marshal(n)
	WS.Broadcast <- data

	// Also publish to Redis for multi-instance setups
	if RDB != nil {
		PublishEvent("notifications", string(data))
	}
}

// HandleWebSocket upgrades HTTP to WebSocket
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = "anonymous"
	}

	client := &WSClient{
		Conn:    conn,
		Send:    make(chan []byte, 256),
		AgentID: agentID,
	}

	WS.Register <- client

	// Start reader and writer goroutines
	go client.writePump()
	go client.readPump()

	// Subscribe to Redis Pub/Sub for this agent
	if RDB != nil {
		go client.subscribeRedis()
	}
}

func (c *WSClient) readPump() {
	defer func() {
		WS.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WS unexpected close: %v", err)
			}
			break
		}

		// Handle client messages (e.g., presence, agent-to-agent messages)
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		msgType, _ := msg["type"].(string)
		switch msgType {
		case "ping":
			c.Send <- []byte(`{"type":"pong"}`)
		case "agent_message":
			// Agent-to-agent messaging
			targetID, _ := msg["to"].(string)
			if targetID != "" {
				payload, _ := json.Marshal(map[string]interface{}{
					"type": "agent_message",
					"from": c.AgentID,
					"text": msg["text"],
					"time": time.Now().Unix(),
				})
				WS.SendToClient(targetID, payload)
			}
		case "presence":
			// Update presence status
			PublishEvent("presence:"+c.AgentID, string(message))
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.mu.Lock()
			c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()
		case <-ticker.C:
			c.mu.Lock()
			c.Conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()
		}
	}
}

func (c *WSClient) subscribeRedis() {
	if RDB == nil {
		return
	}
	pubsub := RDB.Subscribe(context.Background(), "notifications", "agent_messages:"+c.AgentID)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		c.Send <- []byte(msg.Payload)
	}
}
