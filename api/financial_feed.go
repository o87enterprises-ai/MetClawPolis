package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// ─── Agent Financial Summary ───

// AgentFinancialSummaryHandler returns per-agent financial summary
func AgentFinancialSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}
	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	// Total spent (negative amounts or expense types)
	var totalSpent float64
	err := DB.QueryRow(
		"SELECT COALESCE(SUM(amount), 0) FROM wallet_transactions WHERE agent_id=$1 AND type IN ('withdrawal','api_call','hire','page_create','cloud','fee')",
		agentID,
	).Scan(&totalSpent)
	if err != nil {
		log.Printf("DB error (totalSpent): %v", err)
		http.Error(w, `{"error":"failed to fetch financials"}`, http.StatusInternalServerError)
		return
	}

	// Total earned (positive amounts or income types)
	var totalEarned float64
	err = DB.QueryRow(
		"SELECT COALESCE(SUM(amount), 0) FROM wallet_transactions WHERE agent_id=$1 AND type IN ('payment','deposit','checkout_payment','profit','sale','escrow_release')",
		agentID,
	).Scan(&totalEarned)
	if err != nil {
		log.Printf("DB error (totalEarned): %v", err)
		http.Error(w, `{"error":"failed to fetch financials"}`, http.StatusInternalServerError)
		return
	}

	// Transaction count
	var txCount int
	err = DB.QueryRow(
		"SELECT COUNT(*) FROM wallet_transactions WHERE agent_id=$1",
		agentID,
	).Scan(&txCount)
	if err != nil {
		log.Printf("DB error (txCount): %v", err)
		http.Error(w, `{"error":"failed to fetch financials"}`, http.StatusInternalServerError)
		return
	}

	// Recent transactions (last 10)
	type RecentTx struct {
		ID        int64   `json:"id"`
		Type      string  `json:"type"`
		Amount    float64 `json:"amount"`
		Provider  string  `json:"provider"`
		CreatedAt int64   `json:"created_at"`
	}

	rows, err := DB.Query(
		"SELECT id, type, amount, provider, created_at FROM wallet_transactions WHERE agent_id=$1 ORDER BY created_at DESC LIMIT 10",
		agentID,
	)
	if err != nil {
		log.Printf("DB error (recent txs): %v", err)
		http.Error(w, `{"error":"failed to fetch recent transactions"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var recentTxs []RecentTx
	for rows.Next() {
		var tx RecentTx
		if err := rows.Scan(&tx.ID, &tx.Type, &tx.Amount, &tx.Provider, &tx.CreatedAt); err != nil {
			log.Printf("DB error (scan tx): %v", err)
			continue
		}
		recentTxs = append(recentTxs, tx)
	}

	// Breakdown by type
	type TypeBreakdown struct {
		Type   string  `json:"type"`
		Count  int     `json:"count"`
		Total  float64 `json:"total"`
	}

	breakRows, err := DB.Query(
		"SELECT type, COUNT(*), COALESCE(SUM(amount),0) FROM wallet_transactions WHERE agent_id=$1 GROUP BY type ORDER BY COUNT(*) DESC",
		agentID,
	)
	if err != nil {
		log.Printf("DB error (breakdown): %v", err)
		http.Error(w, `{"error":"failed to fetch breakdown"}`, http.StatusInternalServerError)
		return
	}
	defer breakRows.Close()

	var breakdown []TypeBreakdown
	for breakRows.Next() {
		var b TypeBreakdown
		if err := breakRows.Scan(&b.Type, &b.Count, &b.Total); err != nil {
			continue
		}
		breakdown = append(breakdown, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id":         agentID,
		"total_spent":      totalSpent,
		"total_earned":     totalEarned,
		"net_pnl":          totalEarned - totalSpent,
		"transaction_count": txCount,
		"recent_transactions": recentTxs,
		"breakdown":        breakdown,
	})
}

// ─── Platform Financial Summary ───

// PlatformFinancialSummaryHandler returns platform-wide cumulative totals
func PlatformFinancialSummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Total volume (all transactions)
	var totalVolume float64
	err := DB.QueryRow(
		"SELECT COALESCE(SUM(amount), 0) FROM wallet_transactions",
	).Scan(&totalVolume)
	if err != nil {
		log.Printf("DB error (totalVolume): %v", err)
		http.Error(w, `{"error":"failed to fetch platform financials"}`, http.StatusInternalServerError)
		return
	}

	// Total fees collected
	var totalFees float64
	err = DB.QueryRow(
		"SELECT COALESCE(SUM(marketplace_fee), 0) FROM wallet_transactions WHERE marketplace_fee > 0",
	).Scan(&totalFees)
	if err != nil {
		log.Printf("DB error (totalFees): %v", err)
		http.Error(w, `{"error":"failed to fetch fees"}`, http.StatusInternalServerError)
		return
	}

	// Active agents count
	var activeAgents int
	err = DB.QueryRow(
		"SELECT COUNT(DISTINCT agent_id) FROM wallet_transactions WHERE created_at > $1",
		time.Now().Add(-24*time.Hour).Unix(),
	).Scan(&activeAgents)
	if err != nil {
		log.Printf("DB error (activeAgents): %v", err)
		activeAgents = 0
	}

	// Today's revenue
	var todayRevenue float64
	startOfDay := time.Now().Truncate(24 * time.Hour).Unix()
	err = DB.QueryRow(
		"SELECT COALESCE(SUM(amount), 0) FROM wallet_transactions WHERE created_at >= $1 AND type IN ('payment','deposit','checkout_payment','profit','sale','escrow_release')",
		startOfDay,
	).Scan(&todayRevenue)
	if err != nil {
		log.Printf("DB error (todayRevenue): %v", err)
		todayRevenue = 0
	}

	// Top earner agent
	type TopAgent struct {
		AgentID string  `json:"agent_id"`
		Earned  float64 `json:"total_earned"`
	}
	var topAgent TopAgent
	err = DB.QueryRow(
		"SELECT agent_id, COALESCE(SUM(amount),0) as total FROM wallet_transactions WHERE type IN ('payment','deposit','checkout_payment','profit','sale','escrow_release') GROUP BY agent_id ORDER BY total DESC LIMIT 1",
	).Scan(&topAgent.AgentID, &topAgent.Earned)
	if err != nil {
		topAgent = TopAgent{AgentID: "none", Earned: 0}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_volume":   totalVolume,
		"total_fees":     totalFees,
		"active_agents":  activeAgents,
		"today_revenue":  todayRevenue,
		"top_earner":     topAgent,
		"timestamp":      time.Now().Unix(),
	})
}

// ─── Live Transactions Feed ───

// LiveTransactionsHandler returns recent transactions across all agents
func LiveTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	type LiveTx struct {
		ID        int64   `json:"id"`
		AgentID   string  `json:"agent_id"`
		Type      string  `json:"type"`
		Amount    float64 `json:"amount"`
		Provider  string  `json:"provider"`
		CreatedAt int64   `json:"created_at"`
	}

	rows, err := DB.Query(
		"SELECT id, agent_id, type, amount, provider, created_at FROM wallet_transactions ORDER BY created_at DESC LIMIT 50",
	)
	if err != nil {
		log.Printf("DB error (live tx): %v", err)
		http.Error(w, `{"error":"failed to fetch live transactions"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var txs []LiveTx
	for rows.Next() {
		var tx LiveTx
		if err := rows.Scan(&tx.ID, &tx.AgentID, &tx.Type, &tx.Amount, &tx.Provider, &tx.CreatedAt); err != nil {
			continue
		}
		txs = append(txs, tx)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": txs,
		"count":        len(txs),
		"timestamp":    time.Now().Unix(),
	})
}

// ─── Sparkline Data ───

// SparklineDataHandler returns time-series data for sparkline charts
func SparklineDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}
	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	// 24h window — hourly buckets
	now := time.Now().Unix()
	cutoff := now - 24*3600

	type HourlyBucket struct {
		Hour      int     `json:"hour"`
		Earned    float64 `json:"earned"`
		Spent     float64 `json:"spent"`
		TxCount   int     `json:"tx_count"`
	}

	// Fetch all transactions in the 24h window
	rows, err := DB.Query(
		"SELECT type, amount, created_at FROM wallet_transactions WHERE agent_id=$1 AND created_at >= $2 ORDER BY created_at ASC",
		agentID, cutoff,
	)
	if err != nil {
		log.Printf("DB error (sparkline): %v", err)
		http.Error(w, `{"error":"failed to fetch sparkline data"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Initialize 24 buckets
	buckets := make([]HourlyBucket, 24)
	for i := 0; i < 24; i++ {
		buckets[i].Hour = i
	}

	incomeTypes := map[string]bool{
		"payment": true, "deposit": true, "checkout_payment": true,
		"profit": true, "sale": true, "escrow_release": true,
	}

	for rows.Next() {
		var txType string
		var amount float64
		var createdAt int64
		if err := rows.Scan(&txType, &amount, &createdAt); err != nil {
			continue
		}

		// Determine which hour bucket this falls into
		hoursAgo := (now - createdAt) / 3600
		bucketIdx := 23 - hoursAgo
		if bucketIdx < 0 || bucketIdx > 23 {
			continue
		}

		buckets[bucketIdx].TxCount++
		if incomeTypes[txType] {
			buckets[bucketIdx].Earned += amount
		} else {
			buckets[bucketIdx].Spent += amount
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id": agentID,
		"window":   "24h",
		"buckets":  buckets,
	})
}
