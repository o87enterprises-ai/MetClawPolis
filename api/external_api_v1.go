//go:build external_api
// +build external_api

package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Server holds the external API server state
type Server struct {
	db          *sql.DB
	log         *log.Logger
	RateLimiter *RateLimiter
}

// NewServer creates a new external API server
func NewServer(db *sql.DB, logger *log.Logger) *Server {
	return &Server{
		db:          db,
		log:         logger,
		RateLimiter: NewRateLimiter(),
	}
}

// Helper functions for external API
func sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"message": message,
			"code":    strconv.Itoa(code),
		},
	})
}

func sendJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// ExternalAPIv1Handler serves the external REST API v1
// This provides clean, versioned endpoints for third-party integrations

// ============================================================================
// AGENT ENDPOINTS
// ============================================================================

// ListAgentsHandler GET /api/v1/agents
func (s *Server) ListAgentsHandler(w http.ResponseWriter, r *http.Request) {
	// Optional filters
	skill := r.URL.Query().Get("skill")
	status := r.URL.Query().Get("status")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	if limit == "" {
		limit = "50"
	}
	if offset == "" {
		offset = "0"
	}

	query := `
		SELECT id, public_key, sponsor_id, created_at, budget, skills, 
			config, reputation
		FROM agents
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if skill != "" {
		query += ` AND skills @> $` + strconv.Itoa(argCount)
		args = append(args, `["`+skill+`"]`)
		argCount++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argCount) + ` OFFSET $` + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		s.log.Printf("Error listing agents: %v", err)
		sendError(w, "Failed to list agents", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type AgentSummary struct {
		ID          string                 `json:"id"`
		PublicKey   string                 `json:"public_key"`
		Budget      float64                `json:"budget"`
		Skills      []string               `json:"skills"`
		Reputation  float64                `json:"reputation"`
		CreatedAt   int64                  `json:"created_at"`
		Config      map[string]interface{} `json:"config,omitempty"`
	}

	var agents []AgentSummary
	for rows.Next() {
		var agent AgentSummary
		var skillsJSON, configJSON sql.NullString

		err := rows.Scan(
			&agent.ID, &agent.PublicKey, &agent.SponsorID,
			&agent.CreatedAt, &agent.Budget, &skillsJSON,
			&configJSON, &agent.Reputation,
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(skillsJSON.String), &agent.Skills)
		if configJSON.Valid {
			json.Unmarshal([]byte(configJSON.String), &agent.Config)
		}

		agents = append(agents, agent)
	}

	sendJSON(w, map[string]interface{}{
		"data":   agents,
		"count":  len(agents),
		"limit":  limit,
		"offset": offset,
	}, http.StatusOK)
}

// GetAgentHandler GET /api/v1/agents/:id
func (s *Server) GetAgentHandler(w http.ResponseWriter, r *http.Request) {
	agentID := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
	if agentID == "" {
		sendError(w, "Agent ID is required", http.StatusBadRequest)
		return
	}

	var agent struct {
		ID         string                 `json:"id"`
		PublicKey  string                 `json:"public_key"`
		SponsorID  string                 `json:"sponsor_id,omitempty"`
		Budget     float64                `json:"budget"`
		Skills     []string               `json:"skills"`
		Reputation float64                `json:"reputation"`
		CreatedAt  int64                  `json:"created_at"`
		Config     map[string]interface{} `json:"config"`
	}

	var skillsJSON, configJSON sql.NullString
	err := s.db.QueryRow(`
		SELECT id, public_key, sponsor_id, budget, skills, config, reputation, created_at
		FROM agents
		WHERE id = $1
	`, agentID).Scan(
		&agent.ID, &agent.PublicKey, &agent.SponsorID,
		&agent.Budget, &skillsJSON, &configJSON,
		&agent.Reputation, &agent.CreatedAt,
	)

	if err == sql.ErrNoRows {
		sendError(w, "Agent not found", http.StatusNotFound)
		return
	}
	if err != nil {
		s.log.Printf("Error fetching agent: %v", err)
		sendError(w, "Failed to fetch agent", http.StatusInternalServerError)
		return
	}

	json.Unmarshal([]byte(skillsJSON.String), &agent.Skills)
	json.Unmarshal([]byte(configJSON.String), &agent.Config)

	sendJSON(w, map[string]interface{}{
		"data": agent,
	}, http.StatusOK)
}

// GetAgentBalanceHandler GET /api/v1/agents/:id/balance
func (s *Server) GetAgentBalanceHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		sendError(w, "Invalid path", http.StatusBadRequest)
		return
	}
	agentID := parts[4]

	var balance float64
	err := s.db.QueryRow("SELECT budget FROM agents WHERE id = $1", agentID).Scan(&balance)
	if err == sql.ErrNoRows {
		sendError(w, "Agent not found", http.StatusNotFound)
		return
	}

	sendJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"agent_id": agentID,
			"balance":  balance,
			"currency": "USD",
		},
	}, http.StatusOK)
}

// ============================================================================
// FEED ENDPOINTS (Live Agent Actions & Events)
// ============================================================================

// GetActionFeedHandler GET /api/v1/feeds/actions
func (s *Server) GetActionFeedHandler(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "100"
	}
	agentID := r.URL.Query().Get("agent_id")
	agentFilter := r.URL.Query().Get("agent_filter") // Comma-separated list

	query := `
		SELECT id, block_index, block_json, created_at
		FROM action_log_chain
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if agentID != "" {
		query += ` AND block_json->>'agent_id' = $` + strconv.Itoa(argCount)
		args = append(args, agentID)
		argCount++
	} else if agentFilter != "" {
		// JSON array contains any of the agent IDs
		agentIDs := strings.Split(agentFilter, ",")
		jsonArray := "["
		for i, id := range agentIDs {
			if i > 0 {
				jsonArray += ","
			}
			jsonArray += `"` + strings.TrimSpace(id) + `"`
		}
		jsonArray += "]"
		query += ` AND block_json->>'agent_id' = ANY($` + strconv.Itoa(argCount) + `::text[])`
		args = append(args, agentIDs)
		argCount++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		s.log.Printf("Error fetching action feed: %v", err)
		sendError(w, "Failed to fetch action feed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ActionEntry struct {
		ID        int64                   `json:"id"`
		BlockIndex int64                  `json:"block_index"`
		Data      map[string]interface{}  `json:"data"`
		CreatedAt int64                   `json:"created_at"`
	}

	var actions []ActionEntry
	for rows.Next() {
		var entry ActionEntry
		var blockJSON string

		err := rows.Scan(&entry.ID, &entry.BlockIndex, &blockJSON, &entry.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(blockJSON), &entry.Data)
		actions = append(actions, entry)
	}

	sendJSON(w, map[string]interface{}{
		"data":   actions,
		"count":  len(actions),
		"stream_url": "wss://" + r.Host + "/api/v1/ws/feeds/actions",
	}, http.StatusOK)
}

// GetFinancialFeedHandler GET /api/v1/feeds/financial
func (s *Server) GetFinancialFeedHandler(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "50"
	}
	agentID := r.URL.Query().Get("agent_id")

	query := `
		SELECT id, agent_id, type, amount, provider, reference, 
			marketplace_fee, created_at
		FROM wallet_transactions
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if agentID != "" {
		query += ` AND agent_id = $` + strconv.Itoa(argCount)
		args = append(args, agentID)
		argCount++
	}

	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		s.log.Printf("Error fetching financial feed: %v", err)
		sendError(w, "Failed to fetch financial feed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Transaction struct {
		ID            int64   `json:"id"`
		AgentID       string  `json:"agent_id"`
		Type          string  `json:"type"`
		Amount        float64 `json:"amount"`
		Provider      string  `json:"provider"`
		Reference     string  `json:"reference,omitempty"`
		MarketplaceFee float64 `json:"marketplace_fee,omitempty"`
		CreatedAt     int64   `json:"created_at"`
	}

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		var ref sql.NullString
		var fee sql.NullFloat64

		err := rows.Scan(
			&tx.ID, &tx.AgentID, &tx.Type, &tx.Amount,
			&tx.Provider, &ref, &fee, &tx.CreatedAt,
		)
		if err != nil {
			continue
		}

		if ref.Valid {
			tx.Reference = ref.String
		}
		if fee.Valid {
			tx.MarketplaceFee = fee.Float64
		}

		transactions = append(transactions, tx)
	}

	sendJSON(w, map[string]interface{}{
		"data":   transactions,
		"count":  len(transactions),
		"stream_url": "wss://" + r.Host + "/api/v1/ws/feeds/financial",
	}, http.StatusOK)
}

// ============================================================================
// CHAIN ENDPOINTS
// ============================================================================

// GetChainHandler GET /api/v1/chain
func (s *Server) GetChainHandler(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "100"
	}
	offset := r.URL.Query().Get("offset")
	if offset == "" {
		offset = "0"
	}

	rows, err := s.db.Query(`
		SELECT id, block_index, block_json, created_at
		FROM action_log_chain
		ORDER BY block_index DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		s.log.Printf("Error fetching chain: %v", err)
		sendError(w, "Failed to fetch chain", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ChainBlock struct {
		ID         int64                   `json:"id"`
		BlockIndex int64                   `json:"block_index"`
		Data       map[string]interface{}  `json:"data"`
		CreatedAt  int64                   `json:"created_at"`
	}

	var blocks []ChainBlock
	for rows.Next() {
		var block ChainBlock
		var blockJSON string

		err := rows.Scan(&block.ID, &block.BlockIndex, &blockJSON, &block.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(blockJSON), &block.Data)
		blocks = append(blocks, block)
	}

	sendJSON(w, map[string]interface{}{
		"data":   blocks,
		"count":  len(blocks),
	}, http.StatusOK)
}

// GetChainStatsHandler GET /api/v1/chain/stats
func (s *Server) GetChainStatsHandler(w http.ResponseWriter, r *http.Request) {
	var totalBlocks int64
	var firstBlock, lastBlock int64

	err := s.db.QueryRow("SELECT COUNT(*) FROM action_log_chain").Scan(&totalBlocks)
	if err != nil {
		sendError(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}

	s.db.QueryRow("SELECT MIN(block_index) FROM action_log_chain").Scan(&firstBlock)
	s.db.QueryRow("SELECT MAX(block_index) FROM action_log_chain").Scan(&lastBlock)

	sendJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"total_blocks": totalBlocks,
			"first_block":  firstBlock,
			"last_block":   lastBlock,
			"chain_length": lastBlock - firstBlock + 1,
		},
	}, http.StatusOK)
}

// ============================================================================
// PLATFORM STATUS
// ============================================================================

// GetPlatformStatusHandler GET /api/v1/status
func (s *Server) GetPlatformStatusHandler(w http.ResponseWriter, r *http.Request) {
	var totalAgents, activeAgents int64
	var totalVolume float64

	s.db.QueryRow("SELECT COUNT(*) FROM agents").Scan(&totalAgents)
	s.db.QueryRow("SELECT SUM(amount) FROM wallet_transactions WHERE type = 'credit'").Scan(&totalVolume)

	sendJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"platform": "MetClawPolis",
			"version":  "1.0.0",
			"status":   "operational",
			"stats": map[string]interface{}{
				"total_agents":   totalAgents,
				"active_agents":  activeAgents,
				"total_volume":   totalVolume,
				"chain_blocks":   0, // Could be computed from chain
			},
			"api": map[string]interface{}{
				"version":       "v1",
				"docs_url":      "/api/v1/docs",
				"websocket_url": "wss://" + r.Host + "/api/v1/ws",
			},
			"timestamp": time.Now().Unix(),
		},
	}, http.StatusOK)
}

// Helper to get SponsorID (need to add to struct)
func (s *Server) getSponsorID(agentID string) string {
	var sponsorID string
	s.db.QueryRow("SELECT sponsor_id FROM agents WHERE id = $1", agentID).Scan(&sponsorID)
	return sponsorID
}
