package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket feed types
const (
	FeedTypeActions    = "actions"
	FeedTypeFinancial  = "financial"
	FeedTypeChain      = "chain"
	FeedTypePresence   = "presence"
	FeedTypeAll        = "all"
)

// WSFeedClient represents a WebSocket client connected to a feed
type WSFeedClient struct {
	conn       *websocket.Conn
	send       chan map[string]interface{}
	developerID string
	feedType   string
	agentFilter []string
	mu         sync.Mutex
}

// WSFeedManager manages multiple feed channels
type WSFeedManager struct {
	clients    map[string]map[*WSFeedClient]bool // feedType -> clients
	mu         sync.RWMutex
	broadcast  chan FeedEvent
	register   chan WSFeedRegistration
	unregister chan WSFeedRegistration
}

// FeedEvent represents an event to broadcast
type FeedEvent struct {
	FeedType string                 `json:"feed_type"`
	Event    string                 `json:"event"`
	Data     map[string]interface{} `json:"data"`
	Timestamp int64                `json:"timestamp"`
}

// WSFeedRegistration handles client registration/unregistration
type WSFeedRegistration struct {
	client *WSFeedClient
	feedType string
}

var FeedManager *WSFeedManager

func InitFeedManager() {
	FeedManager = &WSFeedManager{
		clients:    make(map[string]map[*WSFeedClient]bool),
		broadcast:  make(chan FeedEvent, 256),
		register:   make(chan WSFeedRegistration),
		unregister: make(chan WSFeedRegistration),
	}

	go FeedManager.run()
}

func (fm *WSFeedManager) run() {
	for {
		select {
		case registration := <-fm.register:
			fm.mu.Lock()
			if _, exists := fm.clients[registration.feedType]; !exists {
				fm.clients[registration.feedType] = make(map[*WSFeedClient]bool)
			}
			fm.clients[registration.feedType][registration.client] = true
			fm.mu.Unlock()

		case registration := <-fm.unregister:
			fm.mu.Lock()
			if clients, exists := fm.clients[registration.feedType]; exists {
				delete(clients, registration.client)
				close(registration.client.send)
			}
			fm.mu.Unlock()

		case event := <-fm.broadcast:
			fm.mu.RLock()
			// Broadcast to specific feed type
			if clients, exists := fm.clients[event.FeedType]; exists {
				for client := range clients {
					select {
					case client.send <- map[string]interface{}{
						"type":      event.Event,
						"feed_type": event.FeedType,
						"data":      event.Data,
						"timestamp": event.Timestamp,
					}:
					default:
						// Client buffer full, skip
						close(client.send)
						delete(clients, client)
					}
				}
			}

			// Also broadcast to "all" feed subscribers
			if clients, exists := fm.clients[FeedTypeAll]; exists {
				for client := range clients {
					// Apply agent filter if set
					if len(client.agentFilter) > 0 {
						agentID, ok := event.Data["agent_id"].(string)
						if !ok || !containsAgent(client.agentFilter, agentID) {
							continue
						}
					}

					select {
					case client.send <- map[string]interface{}{
						"type":      event.Event,
						"feed_type": event.FeedType,
						"data":      event.Data,
						"timestamp": event.Timestamp,
					}:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
			fm.mu.RUnlock()
		}
	}
}

// Broadcast sends an event to all subscribers of a feed type
func BroadcastEvent(feedType string, event string, data map[string]interface{}) {
	if FeedManager == nil {
		return
	}

	FeedManager.broadcast <- FeedEvent{
		FeedType:  feedType,
		Event:     event,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

func containsAgent(filter []string, agentID string) bool {
	for _, id := range filter {
		if id == agentID {
			return true
		}
	}
	return false
}

// LiveFeedWebSocketHandler handles WebSocket connections for live feeds
// GET /api/v1/ws/feeds/:feedType
func (s *Server) LiveFeedWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for external API
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Extract API key from query param or subprotocol
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"api_key query parameter is required"}`))
		conn.Close()
		return
	}

	// Validate API key
	validatedKey, err := s.ValidateAPIKey(apiKey)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Invalid API key"}`))
		conn.Close()
		return
	}

	// Extract feed type from path
	feedType := extractFeedType(r.URL.Path)

	// Extract optional agent filter
	agentFilter := r.URL.Query().Get("agent_filter") // Comma-separated
	var agents []string
	if agentFilter != "" {
		for _, id := range strings.Split(agentFilter, ",") {
			if id != "" {
				agents = append(agents, id)
			}
		}
	}

	// Create client
	client := &WSFeedClient{
		conn:        conn,
		send:        make(chan map[string]interface{}, 256),
		developerID: validatedKey.DeveloperID,
		feedType:    feedType,
		agentFilter: agents,
	}

	// Register client
	FeedManager.register <- WSFeedRegistration{
		client:   client,
		feedType: feedType,
	}

	// Start read/write loops
	go client.writePump()
	go client.readPump(s, FeedManager)
}

func (fc *WSFeedClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		fc.conn.Close()
	}()

	for {
		select {
		case message, ok := <-fc.send:
			fc.mu.Lock()
			fc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				fc.conn.WriteMessage(websocket.CloseMessage, []byte{})
				fc.mu.Unlock()
				return
			}

			err := fc.conn.WriteJSON(message)
			fc.mu.Unlock()

			if err != nil {
				return
			}

		case <-ticker.C:
			fc.mu.Lock()
			fc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := fc.conn.WriteMessage(websocket.PingMessage, nil)
			fc.mu.Unlock()

			if err != nil {
				return
			}
		}
	}
}

func (fc *WSFeedClient) readPump(s *Server, fm *WSFeedManager) {
	defer func() {
		fm.unregister <- WSFeedRegistration{
			client:   fc,
			feedType: fc.feedType,
		}
		fc.conn.Close()
	}()

	fc.conn.SetReadLimit(512)
	fc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	fc.conn.SetPongHandler(func(string) error {
		fc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := fc.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				s.log.Printf("WebSocket unexpected close: %v", err)
			}
			break
		}

		// Handle control messages
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		msgType, _ := msg["type"].(string)
		switch msgType {
		case "ping":
			fc.mu.Lock()
			fc.conn.WriteJSON(map[string]string{"type": "pong"})
			fc.mu.Unlock()
		case "subscribe":
			// Update agent filter
			if filter, ok := msg["agent_filter"].([]interface{}); ok {
				fc.agentFilter = []string{}
				for _, f := range filter {
					if id, ok := f.(string); ok {
						fc.agentFilter = append(fc.agentFilter, id)
					}
				}
			}
		}
	}
}

func extractFeedType(path string) string {
	parts := strings.Split(path, "/")
	filtered := []string{}
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) >= 4 {
		return filtered[3] // /api/v1/ws/feeds/actions -> actions
	}
	return FeedTypeAll
}
