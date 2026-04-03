package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/metclawpolis/agent-platform/chain"
)

const (
	MarketplaceFeeRate   = 0.005 // 0.5% transaction fee
	HiringFeeRate        = 0.01  // 1.0% hiring fee
)

// HireAgentRequest is the request for agent hiring
type HireAgentRequest struct {
	ContractorID string  `json:"contractor_id"`
	Amount       float64 `json:"amount"`
	RevenueSplit float64 `json:"revenue_split"` // 0.7 = employer keeps 70%, contractor gets 30%
	Scope        string  `json:"scope"`
}

// HireAgentHandler creates an escrow for agent hiring
func HireAgentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	employerID := r.Header.Get("X-Verified-Agent-ID")

	var req HireAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.ContractorID == "" || req.Amount <= 0 || req.Scope == "" {
		http.Error(w, `{"error":"contractor_id, amount (>0), and scope are required"}`, http.StatusBadRequest)
		return
	}

	if req.RevenueSplit < 0 || req.RevenueSplit > 1 {
		req.RevenueSplit = 0.7 // default
	}

	// Verify employer has sufficient balance
	var employerBalance float64
	err := DB.QueryRow("SELECT budget FROM agents WHERE id=$1", employerID).Scan(&employerBalance)
	if err != nil {
		http.Error(w, `{"error":"employer not found"}`, http.StatusNotFound)
		return
	}

	hiringFee := req.Amount * HiringFeeRate
	totalCost := req.Amount + hiringFee

	if employerBalance < totalCost {
		http.Error(w, `{"error":"insufficient balance"}`, http.StatusBadRequest)
		return
	}

	// Verify contractor exists
	var contractorExists bool
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM agents WHERE id=$1)", req.ContractorID).Scan(&contractorExists)
	if err != nil || !contractorExists {
		http.Error(w, `{"error":"contractor not found"}`, http.StatusNotFound)
		return
	}

	// Create escrow
	escrowID := fmt.Sprintf("escrow_%d_%s", time.Now().Unix(), employerID[:8])
	_, err = DB.Exec(
		"INSERT INTO escrow (id, employer_id, contractor_id, amount, revenue_split, scope, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		escrowID, employerID, req.ContractorID, req.Amount, req.RevenueSplit, req.Scope, "active", time.Now().Unix(),
	)
	if err != nil {
		http.Error(w, `{"error":"failed to create escrow"}`, http.StatusInternalServerError)
		return
	}

	// Deduct from employer
	_, err = DB.Exec("UPDATE agents SET budget = budget - $1 WHERE id=$2", totalCost, employerID)
	if err != nil {
		http.Error(w, `{"error":"failed to deduct from employer"}`, http.StatusInternalServerError)
		return
	}

	// Record transaction
	_, err = DB.Exec(
		"INSERT INTO wallet_transactions (agent_id, type, amount, provider, reference, marketplace_fee, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		employerID, "hire_payment", totalCost, "escrow", escrowID, hiringFee, time.Now().Unix(),
	)
	if err != nil {
		http.Error(w, `{"error":"failed to record transaction"}`, http.StatusInternalServerError)
		return
	}

	// Log action
	inputHash := chain.CalculateInputHash(employerID, req.ContractorID, fmt.Sprintf("%.2f", req.Amount))
	action := ActionLogRequest{
		AgentID:   employerID,
		Type:      "HIRE",
		InputHash: inputHash,
		Meta:      fmt.Sprintf(`{"contractor":"%s","amount":%.2f,"split":%.2f,"escrow":"%s"}`, req.ContractorID, req.Amount, req.RevenueSplit, escrowID),
	}
	logAction(action, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "escrow created",
		"escrow_id":     escrowID,
		"amount":        req.Amount,
		"hiring_fee":    hiringFee,
		"total_cost":    totalCost,
		"revenue_split": req.RevenueSplit,
	})
}

// CompleteEscrowHandler completes an escrow and distributes funds
func CompleteEscrowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EscrowID string `json:"escrow_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Fetch escrow
	var employerID, contractorID, scope string
	var amount, revenueSplit float64
	var status string
	err := DB.QueryRow(
		"SELECT employer_id, contractor_id, amount, revenue_split, scope, status FROM escrow WHERE id=$1",
		req.EscrowID,
	).Scan(&employerID, &contractorID, &amount, &revenueSplit, &scope, &status)
	if err != nil {
		http.Error(w, `{"error":"escrow not found"}`, http.StatusNotFound)
		return
	}

	if status != "active" {
		http.Error(w, `{"error":"escrow is not active"}`, http.StatusBadRequest)
		return
	}

	// Calculate split
	employerShare := amount * revenueSplit
	contractorShare := amount * (1 - revenueSplit)

	// Update balances
	_, err = DB.Exec("UPDATE agents SET budget = budget + $1 WHERE id=$2", employerShare, employerID)
	if err != nil {
		http.Error(w, `{"error":"failed to update employer balance"}`, http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec("UPDATE agents SET budget = budget + $1 WHERE id=$2", contractorShare, contractorID)
	if err != nil {
		http.Error(w, `{"error":"failed to update contractor balance"}`, http.StatusInternalServerError)
		return
	}

	// Update escrow status
	_, err = DB.Exec("UPDATE escrow SET status='completed', completed_at=$1 WHERE id=$2", time.Now().Unix(), req.EscrowID)
	if err != nil {
		http.Error(w, `{"error":"failed to update escrow"}`, http.StatusInternalServerError)
		return
	}

	// Log action
	inputHash := chain.CalculateInputHash(req.EscrowID, "COMPLETE")
	action := ActionLogRequest{
		AgentID:   employerID,
		Type:      "ESCROW_COMPLETE",
		InputHash: inputHash,
		Meta: fmt.Sprintf(`{"escrow":"%s","employer_share":%.2f,"contractor_share":%.2f}`,
			req.EscrowID, employerShare, contractorShare),
	}
	logAction(action, "")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "escrow completed",
		"escrow_id":       req.EscrowID,
		"employer_share":  employerShare,
		"contractor_share": contractorShare,
	})
}

// GetEscrowHandler retrieves escrow details
func GetEscrowHandler(w http.ResponseWriter, r *http.Request) {
	escrowID := r.URL.Query().Get("id")
	if escrowID == "" {
		http.Error(w, `{"error":"escrow id is required"}`, http.StatusBadRequest)
		return
	}

	var employerID, contractorID, scope, status string
	var amount, revenueSplit float64
	var createdAt int64
	var completedAt *int64

	err := DB.QueryRow(
		"SELECT employer_id, contractor_id, amount, revenue_split, scope, status, created_at, completed_at FROM escrow WHERE id=$1",
		escrowID,
	).Scan(&employerID, &contractorID, &amount, &revenueSplit, &scope, &status, &createdAt, &completedAt)
	if err != nil {
		http.Error(w, `{"error":"escrow not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"escrow_id":      escrowID,
		"employer_id":    employerID,
		"contractor_id":  contractorID,
		"amount":         amount,
		"revenue_split":  revenueSplit,
		"scope":          scope,
		"status":         status,
		"created_at":     createdAt,
		"completed_at":   completedAt,
	})
}
