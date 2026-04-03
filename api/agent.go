package api

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/metclawpolis/agent-platform/chain/did"
)

// CreateAgentRequest is the request to create a new agent
type CreateAgentRequest struct {
	SponsorID string `json:"sponsor_id"`
}

// CreateAgentResponse is the response after creating an agent
type CreateAgentResponse struct {
	Agent      *did.Agent `json:"agent"`
	PrivateKey string     `json:"private_key"` // WARNING: store securely, never log
}

// CreateAgentHandler creates a new agent with ed25519 keys
func CreateAgentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.SponsorID == "" {
		http.Error(w, `{"error":"sponsor_id is required"}`, http.StatusBadRequest)
		return
	}

	// Create new agent
	agent, privKey, err := did.NewAgent(req.SponsorID)
	if err != nil {
		http.Error(w, `{"error":"failed to create agent"}`, http.StatusInternalServerError)
		return
	}

	// Insert into database
	_, err = DB.Exec(
		"INSERT INTO agents (id, public_key, sponsor_id, created_at, budget, skills, config, reputation) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		agent.ID, agent.PublicKey, agent.Sponsor, agent.CreatedAt, agent.Budget, "[]", agent.Config, agent.Reputation,
	)
	if err != nil {
		log.Printf("DB ERROR: %v", err)
		http.Error(w, `{"error":"failed to save agent to database: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Log action to PoW chain
	action := ActionLogRequest{
		AgentID: agent.ID,
		Type:    "CREATE_AGENT",
		Meta:    `{"sponsor":"` + req.SponsorID + `"}`,
	}
	logAction(action, hex.EncodeToString(ed25519.Sign(privKey, []byte("CREATE_AGENT:"+agent.ID))))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateAgentResponse{
		Agent:      agent,
		PrivateKey: hex.EncodeToString(privKey),
	})
}

// GetAgentHandler retrieves an agent's profile
func GetAgentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("id")
	if agentID == "" {
		http.Error(w, `{"error":"id parameter is required"}`, http.StatusBadRequest)
		return
	}

	var agent did.Agent
	var skillsStr, configStr string
	err := DB.QueryRow(
		"SELECT id, public_key, sponsor_id, created_at, budget, skills, config, reputation FROM agents WHERE id=$1",
		agentID,
	).Scan(&agent.ID, &agent.PublicKey, &agent.Sponsor, &agent.CreatedAt, &agent.Budget, &skillsStr, &configStr, &agent.Reputation)
	if err != nil {
		log.Printf("GetAgent error: %v", err)
		http.Error(w, `{"error":"agent not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// UpdateAgentSkillsHandler updates an agent's skills
func UpdateAgentSkillsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")

	var req struct {
		Skills []string `json:"skills"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	skillsJSON, _ := json.Marshal(req.Skills)
	_, err := DB.Exec("UPDATE agents SET skills=$1 WHERE id=$2", string(skillsJSON), agentID)
	if err != nil {
		http.Error(w, `{"error":"failed to update skills"}`, http.StatusInternalServerError)
		return
	}

	// Log action
	action := ActionLogRequest{
		AgentID: agentID,
		Type:    "UPDATE_SKILLS",
		Meta:    string(skillsJSON),
	}
	logAction(action, "")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "skills updated"})
}

// GetAgentBalanceHandler returns an agent's wallet balance
func GetAgentBalanceHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	var balance float64
	err := DB.QueryRow("SELECT budget FROM agents WHERE id=$1", agentID).Scan(&balance)
	if err != nil {
		http.Error(w, `{"error":"agent not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id": agentID,
		"balance":  balance,
		"currency": "USD",
	})
}

// DepositHandler allows agents to deposit funds
func DepositHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	_, err := DB.Exec("UPDATE agents SET budget = budget + $1 WHERE id=$2", req.Amount, agentID)
	if err != nil {
		http.Error(w, `{"error":"failed to deposit"}`, http.StatusInternalServerError)
		return
	}

	// Record transaction
	_, err = DB.Exec(
		"INSERT INTO wallet_transactions (agent_id, type, amount, provider, reference, marketplace_fee, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		agentID, "deposit", req.Amount, "system", "", 0, time.Now().Unix(),
	)
	if err != nil {
		http.Error(w, `{"error":"failed to record transaction"}`, http.StatusInternalServerError)
		return
	}

	// Log action
	action := ActionLogRequest{
		AgentID: agentID,
		Type:    "DEPOSIT",
		Meta:    `{"amount":` + fmt.Sprintf("%.2f", req.Amount) + `}`,
	}
	logAction(action, "")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "deposited",
		"amount":  req.Amount,
		"agent_id": agentID,
	})
}
