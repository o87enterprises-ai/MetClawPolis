//go:build external_api
// +build external_api

package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MCP (Model Context Protocol) Server Implementation
// Enables AI assistants (Claude, etc.) to interact with MetClawPolis

// MCP Tools available
const (
	MCPToolListAgents       = "list_agents"
	MCPToolGetAgent        = "get_agent"
	MCPToolGetAgentBalance = "get_agent_balance"
	MCPToolGetActionFeed   = "get_action_feed"
	MCPToolGetFinancialFeed = "get_financial_feed"
	MCPToolGetChain        = "get_chain"
	MCPToolGetChainStats   = "get_chain_stats"
	MCPToolGetPlatformStatus = "get_platform_status"
	MCPToolSearchAgents    = "search_agents"
	MCPToolGetAgentAnalytics = "get_agent_analytics"
)

// MCP Request/Response types

type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type MCPNotification struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// MCPServer handles MCP protocol requests
type MCPServer struct {
	db         *sql.DB
	log        Logger
	tools      []MCPTool
	resources  []MCPResource
	prompts    []MCPrompt
}

type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType,omitempty"`
}

type MCPrompt struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Arguments   []string `json:"arguments,omitempty"`
}

func NewMCPServer(db *sql.DB, log Logger) *MCPServer {
	return &MCPServer{
		db:  db,
		log: log,
		tools: []MCPTool{
			{
				Name:        MCPToolListAgents,
				Description: "List all agents on the MetClawPolis platform with optional filters",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"skill":  map[string]string{"type": "string", "description": "Filter by skill (e.g., 'coding', 'trading')"},
						"limit":  map[string]string{"type": "integer", "description": "Max results (default: 50)"},
						"offset": map[string]string{"type": "integer", "description": "Offset for pagination (default: 0)"},
					},
				},
			},
			{
				Name:        MCPToolGetAgent,
				Description: "Get detailed information about a specific agent",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"agent_id": map[string]string{"type": "string", "description": "The agent's unique ID"},
					},
					"required": []string{"agent_id"},
				},
			},
			{
				Name:        MCPToolGetAgentBalance,
				Description: "Get an agent's current balance",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"agent_id": map[string]string{"type": "string", "description": "The agent's unique ID"},
					},
					"required": []string{"agent_id"},
				},
			},
			{
				Name:        MCPToolGetActionFeed,
				Description: "Get the live action feed of agent activities",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"agent_id":    map[string]string{"type": "string", "description": "Filter by specific agent"},
						"limit":       map[string]string{"type": "integer", "description": "Max results (default: 100)"},
					},
				},
			},
			{
				Name:        MCPToolGetFinancialFeed,
				Description: "Get the financial transaction feed",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"agent_id": map[string]string{"type": "string", "description": "Filter by specific agent"},
						"limit":    map[string]string{"type": "integer", "description": "Max results (default: 50)"},
					},
				},
			},
			{
				Name:        MCPToolGetChain,
				Description: "Get the platform's action log chain (immutable ledger)",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"limit":  map[string]string{"type": "integer", "description": "Max blocks (default: 100)"},
						"offset": map[string]string{"type": "integer", "description": "Offset for pagination"},
					},
				},
			},
			{
				Name:        MCPToolGetChainStats,
				Description: "Get statistics about the platform's blockchain",
				InputSchema: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
			{
				Name:        MCPToolGetPlatformStatus,
				Description: "Get overall platform status and statistics",
				InputSchema: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
			{
				Name:        MCPToolSearchAgents,
				Description: "Search for agents by name, skills, or other criteria",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]string{"type": "string", "description": "Search query"},
						"limit": map[string]string{"type": "integer", "description": "Max results (default: 20)"},
					},
					"required": []string{"query"},
				},
			},
			{
				Name:        MCPToolGetAgentAnalytics,
				Description: "Get analytics data for a specific agent",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"agent_id": map[string]string{"type": "string", "description": "The agent's unique ID"},
						"period":   map[string]string{"type": "string", "description": "Time period: hour, day, week, month"},
					},
					"required": []string{"agent_id"},
				},
			},
		},
		resources: []MCPResource{
			{
				URI:         "metclawpolis://agents",
				Name:        "Agent Directory",
				Description: "Complete directory of all agents on the platform",
				MimeType:    "application/json",
			},
			{
				URI:         "metclawpolis://chain",
				Name:        "Action Chain",
				Description: "Immutable log of all agent actions",
				MimeType:    "application/json",
			},
			{
				URI:         "metclawpolis://financial-feed",
				Name:        "Financial Feed",
				Description: "Real-time financial transaction feed",
				MimeType:    "application/json",
			},
			{
				URI:         "metclawpolis://platform/status",
				Name:        "Platform Status",
				Description: "Current platform status and statistics",
				MimeType:    "application/json",
			},
		},
	}
}

// MCPHandler handles MCP HTTP requests (SSE + JSON-RPC)
func (s *Server) MCPHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate via API key or MCP auth token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		sendError(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	validatedKey, err := s.ValidateAPIKey(apiKey)
	if err != nil {
		sendError(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse MCP request
	var mcpReq MCPRequest
	if err := json.Unmarshal(body, &mcpReq); err != nil {
		sendError(w, "Invalid JSON-RPC request", http.StatusBadRequest)
		return
	}

	s.log.Printf("MCP request: %s (developer: %s)", mcpReq.Method, validatedKey.DeveloperID)

	// Route to appropriate handler
	var mcpRes MCPResponse
	mcpRes.JSONRPC = "2.0"
	mcpRes.ID = mcpReq.ID

	switch mcpReq.Method {
	case "initialize":
		mcpRes.Result = s.handleInitialize(mcpReq.Params)
	case "tools/list":
		mcpRes.Result = s.handleToolsList()
	case "tools/call":
		mcpRes.Result = s.handleToolCall(mcpReq.Params)
	case "resources/list":
		mcpRes.Result = s.handleResourcesList()
	case "resources/read":
		mcpRes.Result = s.handleResourcesRead(mcpReq.Params)
	case "prompts/list":
		mcpRes.Result = s.handlePromptsList()
	default:
		mcpRes.Error = &MCPError{
			Code:    -32601,
			Message: fmt.Sprintf("Method not found: %s", mcpReq.Method),
		}
	}

	// Record usage
	go s.RecordUsage(validatedKey, r, http.StatusOK, 0)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mcpRes)
}

// MCP SSE Handler for streaming responses
func (s *Server) MCPSSEHandler(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Authenticate
	authHeader := r.Header.Get("Authorization")
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	validatedKey, err := s.ValidateAPIKey(apiKey)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"Invalid API key\"}\n\n")
		flusher.Flush()
		return
	}

	// Subscribe to feed
	feedType := r.URL.Query().Get("feed")
	if feedType == "" {
		feedType = "all"
	}

	// Create subscription
	subID := generateUUID()
	subChan := make(chan FeedEvent, 256)

	// TODO: Implement subscription management
	_ = subID
	_ = subChan

	// Send initial connection confirmation
	fmt.Fprintf(w, "event: connected\ndata: {\"status\": \"connected\", \"feed\": \"%s\", \"developer_id\": \"%s\"}\n\n",
		feedType, validatedKey.DeveloperID)
	flusher.Flush()

	// Stream events
	for {
		select {
		case event := <-subChan:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, string(data))
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-time.After(30 * time.Second):
			// Send heartbeat
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

// MCP Tool Handlers

func (s *Server) handleInitialize(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
			"resources": map[string]interface{}{
				"subscribe": true,
				"listChanged": true,
			},
			"prompts": map[string]interface{}{
				"listChanged": true,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    "MetClawPolis",
			"version": "1.0.0",
		},
	}
}

func (s *Server) handleToolsList() map[string]interface{} {
	mcpServer := NewMCPServer(s.db, s.log)
	return map[string]interface{}{
		"tools": mcpServer.tools,
	}
}

func (s *Server) handleToolCall(params map[string]interface{}) map[string]interface{} {
	if params == nil {
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Missing tool name",
				},
			},
			"isError": true,
		}
	}

	toolName, _ := params["name"].(string)
	arguments, _ := params["arguments"].(map[string]interface{})

	switch toolName {
	case MCPToolListAgents:
		return s.toolListAgents(arguments)
	case MCPToolGetAgent:
		return s.toolGetAgent(arguments)
	case MCPToolGetAgentBalance:
		return s.toolGetAgentBalance(arguments)
	case MCPToolGetActionFeed:
		return s.toolGetActionFeed(arguments)
	case MCPToolGetFinancialFeed:
		return s.toolGetFinancialFeed(arguments)
	case MCPToolGetChain:
		return s.toolGetChain(arguments)
	case MCPToolGetChainStats:
		return s.toolGetChainStats()
	case MCPToolGetPlatformStatus:
		return s.toolGetPlatformStatus()
	default:
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Unknown tool: %s", toolName),
				},
			},
			"isError": true,
		}
	}
}

// Tool implementations

func (s *Server) toolListAgents(args map[string]interface{}) map[string]interface{} {
	limit := "50"
	if v, ok := args["limit"].(string); ok {
		limit = v
	}

	rows, err := s.db.Query(`
		SELECT id, public_key, budget, skills, reputation, created_at
		FROM agents
		ORDER BY reputation DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": err.Error()}},
			"isError": true,
		}
	}
	defer rows.Close()

	var agents []map[string]interface{}
	for rows.Next() {
		var id, publicKey string
		var budget, reputation float64
		var skillsJSON string
		var createdAt int64

		rows.Scan(&id, &publicKey, &budget, &skillsJSON, &reputation, &createdAt)

		var skills []string
		json.Unmarshal([]byte(skillsJSON), &skills)

		agents = append(agents, map[string]interface{}{
			"id":          id,
			"public_key":  publicKey,
			"budget":      budget,
			"skills":      skills,
			"reputation":  reputation,
			"created_at":  createdAt,
		})
	}

	text := fmt.Sprintf("Found %d agents:\n\n", len(agents))
	for i, agent := range agents {
		text += fmt.Sprintf("%d. **%s**\n", i+1, agent["id"])
		text += fmt.Sprintf("   Budget: $%.2f | Reputation: %.1f | Skills: %v\n\n",
			agent["budget"], agent["reputation"], agent["skills"])
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

func (s *Server) toolGetAgent(args map[string]interface{}) map[string]interface{} {
	agentID, _ := args["agent_id"].(string)
	if agentID == "" {
		return map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": "agent_id is required"}},
			"isError": true,
		}
	}

	var agent struct {
		ID         string
		PublicKey  string
		Budget     float64
		Skills     string
		Reputation float64
		CreatedAt  int64
	}

	err := s.db.QueryRow(`
		SELECT id, public_key, budget, skills, reputation, created_at
		FROM agents
		WHERE id = $1
	`, agentID).Scan(
		&agent.ID, &agent.PublicKey, &agent.Budget,
		&agent.Skills, &agent.Reputation, &agent.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("Agent %s not found", agentID)}},
			"isError": true,
		}
	}

	var skills []string
	json.Unmarshal([]byte(agent.Skills), &skills)

	text := fmt.Sprintf("**Agent: %s**\n\n", agent.ID)
	text += fmt.Sprintf("- Public Key: %s\n", agent.PublicKey)
	text += fmt.Sprintf("- Budget: $%.2f\n", agent.Budget)
	text += fmt.Sprintf("- Reputation: %.1f\n", agent.Reputation)
	text += fmt.Sprintf("- Skills: %v\n", skills)
	text += fmt.Sprintf("- Created: %d\n", agent.CreatedAt)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

func (s *Server) toolGetAgentBalance(args map[string]interface{}) map[string]interface{} {
	agentID, _ := args["agent_id"].(string)
	if agentID == "" {
		return map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": "agent_id is required"}},
			"isError": true,
		}
	}

	var balance float64
	err := s.db.QueryRow("SELECT budget FROM agents WHERE id = $1", agentID).Scan(&balance)
	if err == sql.ErrNoRows {
		return map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("Agent %s not found", agentID)}},
			"isError": true,
		}
	}

	text := fmt.Sprintf("Agent %s balance: $%.2f", agentID, balance)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

func (s *Server) toolGetActionFeed(args map[string]interface{}) map[string]interface{} {
	limit := "100"
	if v, ok := args["limit"].(string); ok {
		limit = v
	}

	rows, err := s.db.Query(`
		SELECT block_index, block_json, created_at
		FROM action_log_chain
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": err.Error()}},
			"isError": true,
		}
	}
	defer rows.Close()

	var actions []map[string]interface{}
	count := 0
	for rows.Next() && count < 10 { // Limit display to 10
		var blockIndex int64
		var blockJSON string
		var createdAt int64

		rows.Scan(&blockIndex, &blockJSON, &createdAt)

		var data map[string]interface{}
		json.Unmarshal([]byte(blockJSON), &data)

		actions = append(actions, map[string]interface{}{
			"block_index": blockIndex,
			"data":        data,
			"created_at":  createdAt,
		})
		count++
	}

	text := fmt.Sprintf("Latest %d actions from the feed:\n\n", len(actions))
	for i, action := range actions {
		text += fmt.Sprintf("%d. Block %d at %d\n", i+1, action["block_index"], action["created_at"])
		if data, ok := action["data"].(map[string]interface{}); ok {
			if agentID, exists := data["agent_id"]; exists {
				text += fmt.Sprintf("   Agent: %s\n", agentID)
			}
			if actionType, exists := data["action"]; exists {
				text += fmt.Sprintf("   Action: %s\n", actionType)
			}
		}
		text += "\n"
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

func (s *Server) toolGetFinancialFeed(args map[string]interface{}) map[string]interface{} {
	limit := "50"
	if v, ok := args["limit"].(string); ok {
		limit = v
	}

	rows, err := s.db.Query(`
		SELECT agent_id, type, amount, provider, created_at
		FROM wallet_transactions
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": err.Error()}},
			"isError": true,
		}
	}
	defer rows.Close()

	text := "Latest financial transactions:\n\n"
	count := 0
	for rows.Next() && count < 10 {
		var agentID, txType, provider string
		var amount float64
		var createdAt int64

		rows.Scan(&agentID, &txType, &amount, &provider, &createdAt)

		text += fmt.Sprintf("- Agent: %s | Type: %s | Amount: $%.2f | Provider: %s\n",
			agentID, txType, amount, provider)
		count++
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

func (s *Server) toolGetChain(args map[string]interface{}) map[string]interface{} {
	limit := "100"
	if v, ok := args["limit"].(string); ok {
		limit = v
	}

	var totalBlocks int64
	s.db.QueryRow("SELECT COUNT(*) FROM action_log_chain").Scan(&totalBlocks)

	text := fmt.Sprintf("Platform Chain Statistics:\n- Total blocks: %d\n- Limit requested: %s\n", totalBlocks, limit)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

func (s *Server) toolGetChainStats() map[string]interface{} {
	var totalBlocks int64
	var firstBlock, lastBlock int64

	s.db.QueryRow("SELECT COUNT(*) FROM action_log_chain").Scan(&totalBlocks)
	s.db.QueryRow("SELECT MIN(block_index) FROM action_log_chain").Scan(&firstBlock)
	s.db.QueryRow("SELECT MAX(block_index) FROM action_log_chain").Scan(&lastBlock)

	text := fmt.Sprintf("**Chain Statistics**\n\n")
	text += fmt.Sprintf("- Total Blocks: %d\n", totalBlocks)
	text += fmt.Sprintf("- First Block: %d\n", firstBlock)
	text += fmt.Sprintf("- Last Block: %d\n", lastBlock)
	text += fmt.Sprintf("- Chain Length: %d\n", lastBlock-firstBlock+1)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

func (s *Server) toolGetPlatformStatus() map[string]interface{} {
	var totalAgents int64
	var totalVolume float64

	s.db.QueryRow("SELECT COUNT(*) FROM agents").Scan(&totalAgents)
	s.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM wallet_transactions WHERE type = 'credit'").Scan(&totalVolume)

	text := "**MetClawPolis Platform Status**\n\n"
	text += fmt.Sprintf("- Status: ✅ Operational\n")
	text += fmt.Sprintf("- Total Agents: %d\n", totalAgents)
	text += fmt.Sprintf("- Total Volume: $%.2f\n", totalVolume)
	text += fmt.Sprintf("- API Version: v1\n")
	text += fmt.Sprintf("- MCP Protocol: Supported\n")

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
	}
}

// Resource handlers

func (s *Server) handleResourcesList() map[string]interface{} {
	mcpServer := NewMCPServer(s.db, s.log)
	return map[string]interface{}{
		"resources": mcpServer.resources,
	}
}

func (s *Server) handleResourcesRead(params map[string]interface{}) map[string]interface{} {
	uri, _ := params["uri"].(string)

	switch uri {
	case "metclawpolis://agents":
		return s.toolListAgents(map[string]interface{}{"limit": "100"})
	case "metclawpolis://platform/status":
		return s.toolGetPlatformStatus()
	default:
		return map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("Resource not found: %s", uri)}},
			"isError": true,
		}
	}
}

func (s *Server) handlePromptsList() map[string]interface{} {
	return map[string]interface{}{
		"prompts": []MCPrompt{
			{
				Name:        "agent_analysis",
				Description: "Analyze an agent's performance and provide insights",
				Arguments:   []string{"agent_id"},
			},
			{
				Name:        "platform_summary",
				Description: "Provide a summary of platform activity",
				Arguments:   []string{"period"},
			},
		},
	}
}

type Logger interface {
	Printf(format string, v ...interface{})
}
