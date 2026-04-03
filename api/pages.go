package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/metclawpolis/agent-platform/chain"
)

// CreatePageRequest is the request to create an agent page
type CreatePageRequest struct {
	PageURL         string `json:"page_url"`
	StripeAccountID string `json:"stripe_account_id,omitempty"`
	CryptoWallet    string `json:"crypto_wallet,omitempty"`
}

// CreatePageHandler creates a page owned by an agent
func CreatePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")

	var req CreatePageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.PageURL == "" {
		http.Error(w, `{"error":"page_url is required"}`, http.StatusBadRequest)
		return
	}

	_, err := DB.Exec(
		"INSERT INTO agent_pages (agent_id, page_url, stripe_account_id, crypto_wallet, created_at) VALUES ($1, $2, $3, $4, $5)",
		agentID, req.PageURL, req.StripeAccountID, req.CryptoWallet, time.Now().Unix(),
	)
	if err != nil {
		http.Error(w, `{"error":"failed to create page"}`, http.StatusInternalServerError)
		return
	}

	// Log action
	inputHash := chain.CalculateInputHash(agentID, req.PageURL)
	action := ActionLogRequest{
		AgentID:   agentID,
		Type:      "CREATE_PAGE",
		InputHash: inputHash,
		Meta:      fmt.Sprintf(`{"page_url":"%s"}`, req.PageURL),
	}
	logAction(action, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "page created",
		"page_url": req.PageURL,
		"agent_id": agentID,
	})
}

// GetAgentPagesHandler retrieves all pages owned by an agent
func GetAgentPagesHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	rows, err := DB.Query(
		"SELECT id, page_url, stripe_account_id, crypto_wallet, created_at FROM agent_pages WHERE agent_id=$1",
		agentID,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch pages"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Page struct {
		ID              int    `json:"id"`
		PageURL         string `json:"page_url"`
		StripeAccountID string `json:"stripe_account_id,omitempty"`
		CryptoWallet    string `json:"crypto_wallet,omitempty"`
		CreatedAt       int64  `json:"created_at"`
	}

	pages := []Page{}
	for rows.Next() {
		var p Page
		if err := rows.Scan(&p.ID, &p.PageURL, &p.StripeAccountID, &p.CryptoWallet, &p.CreatedAt); err != nil {
			http.Error(w, `{"error":"failed to scan pages"}`, http.StatusInternalServerError)
			return
		}
		pages = append(pages, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id": agentID,
		"pages":    pages,
	})
}
