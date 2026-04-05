package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/metclawpolis/agent-platform/chain"
)

var Blockchain *chain.Blockchain

// ActionLogRequest is the request to log an action
type ActionLogRequest struct {
	AgentID    string `json:"agent_id"`
	Type       string `json:"type"`
	InputHash  string `json:"input_hash"`
	OutputHash string `json:"output_hash"`
	Meta       string `json:"meta"`
}

// InitBlockchain initializes the PoW blockchain
func InitBlockchain() {
	Blockchain = chain.NewBlockchain(2) // difficulty 2 for dev
}

// logAction logs an action to the PoW chain and persists to DB
func logAction(req ActionLogRequest, signature string) chain.Block {
	action := chain.Action{
		AgentID:    req.AgentID,
		Type:       req.Type,
		InputHash:  req.InputHash,
		OutputHash: req.OutputHash,
		Timestamp:  time.Now().Unix(),
		Meta:       req.Meta,
	}

	block := Blockchain.AddAction(action)

	// Persist to database
	blockJSON, _ := json.Marshal(block)
	_, _ = DB.Exec(
		"INSERT INTO action_log_chain (block_index, block_json, created_at) VALUES ($1, $2, $3)",
		block.Index, string(blockJSON), time.Now().Unix(),
	)

	return block
}

// LogActionHandler handles POST /api/action
func LogActionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req ActionLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Use verified agent ID from auth middleware
	if agentID := r.Header.Get("X-Verified-Agent-ID"); agentID != "" {
		req.AgentID = agentID
	}

	block := logAction(req, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "logged",
		"block_hash": block.Hash,
		"block_index": block.Index,
		"nonce":      block.Nonce,
	})
}

// GetChainHandler returns the full PoW chain
func GetChainHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chain":      Blockchain.GetChainJSON(),
		"valid":      Blockchain.Verify(),
		"difficulty": Blockchain.Difficulty,
		"length":     len(Blockchain.GetChainJSON()),
	})
}

// LogBitiverseAction logs a Bitiverse action to the immutable chain
func LogBitiverseAction(agentID, actionType, action string, result string, meta string) chain.Block {
	req := ActionLogRequest{
		AgentID: agentID,
		Type:    "BITIVERSE_" + actionType,
		Meta:    `{"action":"` + action + `","result":"` + result + `",` + meta + `}`,
	}
	return logAction(req, "")
}
