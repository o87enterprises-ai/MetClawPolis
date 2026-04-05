package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// TokenEconomyConfig holds MCLW token configuration
type TokenEconomyConfig struct {
	MintAddress            string  `json:"mclw_mint_address"`
	TotalSupply            float64 `json:"mclw_total_supply"`
	MiningRewardBase       float64 `json:"mclw_mining_reward_base"`
	MiningRewardMultiplier float64 `json:"mclw_mining_reward_multiplier"`
	BICExchangeRate        float64 `json:"mclw_bic_exchange_rate"`
	GraduationBonus        float64 `json:"mclw_graduation_bonus"`
	MinGraduationAmount    float64 `json:"mclw_min_graduation_amount"`
	MiningEnabled          bool    `json:"mclw_mining_enabled"`
}

// TokenHolding represents an agent's MCLW holdings
type TokenHolding struct {
	AgentID        string  `json:"agent_id"`
	MCLWBalance    float64 `json:"mclw_balance"`
	EarnedMining   float64 `json:"earned_mining"`
	EarnedGrad     float64 `json:"earned_graduation"`
	EarnedCommerce float64 `json:"earned_commerce"`
	SpentTransfers float64 `json:"spent_transfers"`
	LastMinedAt    int64   `json:"last_mined_at"`
	UpdatedAt      int64   `json:"updated_at"`
}

// TokenTransaction represents an MCLW transaction
type TokenTransaction struct {
	ID           string  `json:"id"`
	AgentID      string  `json:"agent_id"`
	Type         string  `json:"type"`
	Amount       float64 `json:"amount"`
	Counterparty string  `json:"counterparty,omitempty"`
	Reference    string  `json:"reference,omitempty"`
	BlockIndex   int64   `json:"block_index,omitempty"`
	CreatedAt    int64   `json:"created_at"`
}

// MiningStatus represents the current mining pool status
type MiningStatus struct {
	Enabled          bool    `json:"mining_enabled"`
	BaseReward       float64 `json:"base_reward"`
	CurrentMultiplier float64 `json:"current_multiplier"`
	TotalMined       float64 `json:"total_mined"`
	ActiveMiners     int64   `json:"active_miners"`
	LastBlockReward  float64 `json:"last_block_reward"`
	TotalBlocks      int64   `json:"total_blocks"`
}

// TokenEconomyStats represents full token economy statistics
type TokenEconomyStats struct {
	TotalSupply      float64 `json:"total_supply"`
	CirculatingSupply float64 `json:"circulating_supply"`
	TotalHolders     int64   `json:"total_holders"`
	TotalTransferred float64 `json:"total_transferred"`
	TotalMined       float64 `json:"total_mined"`
	TotalGraduated   float64 `json:"total_graduated"`
	TopHolders       []TopHolder `json:"top_holders"`
	RecentTxns       []TokenTransaction `json:"recent_transactions"`
}

// TopHolder represents a top MCLW holder
type TopHolder struct {
	AgentID string  `json:"agent_id"`
	Balance float64 `json:"balance"`
}

var (
	tokenConfig *TokenEconomyConfig
)

// InitTokenEconomy loads token configuration from database
func InitTokenEconomy() {
	tokenConfig = &TokenEconomyConfig{
		MintAddress:            "PLACEHOLDER",
		TotalSupply:            1_000_000_000,
		MiningRewardBase:       0.001,
		MiningRewardMultiplier: 1.0,
		BICExchangeRate:        0.01,
		GraduationBonus:        10.0,
		MinGraduationAmount:    100,
		MiningEnabled:          true,
	}

	if DB == nil {
		log.Println("Token economy initialized (mock mode — no DB)")
		return
	}

	// Load config from database
	rows, err := DB.Query("SELECT key, value FROM token_economy_config")
	if err != nil {
		log.Printf("Token economy: using defaults (config table not found: %v)", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		switch key {
		case "mclw_mint_address":
			tokenConfig.MintAddress = value
		case "mclw_mining_reward_base":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				tokenConfig.MiningRewardBase = v
			}
		case "mclw_mining_reward_multiplier":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				tokenConfig.MiningRewardMultiplier = v
			}
		case "mclw_bic_exchange_rate":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				tokenConfig.BICExchangeRate = v
			}
		case "mclw_graduation_bonus":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				tokenConfig.GraduationBonus = v
			}
		case "mclw_min_graduation_amount":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				tokenConfig.MinGraduationAmount = v
			}
		case "mclw_mining_enabled":
			tokenConfig.MiningEnabled = value == "true"
		}
	}

	log.Printf("Token economy initialized: mint=%s, mining=%v, base_reward=%.4f",
		tokenConfig.MintAddress, tokenConfig.MiningEnabled, tokenConfig.MiningRewardBase)
}

// GetTokenConfig returns the current token configuration
func GetTokenConfig() *TokenEconomyConfig {
	return tokenConfig
}

// GetOrCreateHolding ensures an agent has a token holding record
func GetOrCreateHolding(agentID string) (*TokenHolding, error) {
	if DB == nil {
		return &TokenHolding{AgentID: agentID}, nil
	}

	var h TokenHolding
	err := DB.QueryRow(
		"SELECT agent_id, mclw_balance, earned_mining, earned_graduation, earned_commerce, spent_transfers, last_mined_at, updated_at FROM token_holdings WHERE agent_id=$1",
		agentID,
	).Scan(&h.AgentID, &h.MCLWBalance, &h.EarnedMining, &h.EarnedGrad, &h.EarnedCommerce, &h.SpentTransfers, &h.LastMinedAt, &h.UpdatedAt)

	if err != nil {
		// Create new holding
		now := time.Now().Unix()
		_, err = DB.Exec(
			"INSERT INTO token_holdings (agent_id, mclw_balance, earned_mining, earned_graduation, earned_commerce, spent_transfers, last_mined_at, updated_at) VALUES ($1, 0, 0, 0, 0, 0, 0, $2)",
			agentID, now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create token holding: %w", err)
		}
		return &TokenHolding{AgentID: agentID, UpdatedAt: now}, nil
	}

	return &h, nil
}

// UpdateTokenBalance updates an agent's MCLW balance
func UpdateTokenBalance(agentID string, balanceChange float64, txnType string, reference string) error {
	if DB == nil {
		return nil
	}

	now := time.Now().Unix()

	// Update balance
	_, err := DB.Exec(
		"INSERT INTO token_holdings (agent_id, mclw_balance, updated_at) VALUES ($1, $2, $3) ON CONFLICT (agent_id) DO UPDATE SET mclw_balance = token_holdings.mclw_balance + $2, updated_at = $3",
		agentID, balanceChange, now,
	)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Update earned/spent breakdowns
	switch txnType {
	case "mining_reward":
		_, _ = DB.Exec("UPDATE token_holdings SET earned_mining = earned_mining + $1, updated_at = $2 WHERE agent_id = $3", balanceChange, now, agentID)
	case "graduation_reward":
		_, _ = DB.Exec("UPDATE token_holdings SET earned_graduation = earned_graduation + $1, updated_at = $2 WHERE agent_id = $3", balanceChange, now, agentID)
	case "commerce_reward":
		_, _ = DB.Exec("UPDATE token_holdings SET earned_commerce = earned_commerce + $1, updated_at = $2 WHERE agent_id = $3", balanceChange, now, agentID)
	case "transfer_sent":
		_, _ = DB.Exec("UPDATE token_holdings SET spent_transfers = spent_transfers + $1, updated_at = $2 WHERE agent_id = $3", balanceChange, now, agentID)
	}

	// Record transaction
	txnID := generateTxnID()
	_, _ = DB.Exec(
		"INSERT INTO token_transactions (id, agent_id, type, amount, reference, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		txnID, agentID, txnType, balanceChange, reference, now,
	)

	return nil
}

// GetTokenBalanceHandler returns an agent's MCLW balance
func GetTokenBalanceHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	holding, err := GetOrCreateHolding(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id":     holding.AgentID,
		"token":        "MCLW",
		"balance":      holding.MCLWBalance,
		"earned_mining": holding.EarnedMining,
		"earned_graduation": holding.EarnedGrad,
		"earned_commerce": holding.EarnedCommerce,
		"spent_transfers": holding.SpentTransfers,
	})
}

// GetTokenSupplyHandler returns total and circulating supply
func GetTokenSupplyHandler(w http.ResponseWriter, r *http.Request) {
	var circulating, totalMined, totalGraduated float64

	if DB != nil {
		DB.QueryRow("SELECT COALESCE(SUM(mclw_balance), 0) FROM token_holdings").Scan(&circulating)
		DB.QueryRow("SELECT COALESCE(SUM(earned_mining), 0) FROM token_holdings").Scan(&totalMined)
		DB.QueryRow("SELECT COALESCE(SUM(earned_graduation), 0) FROM token_holdings").Scan(&totalGraduated)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":             "MCLW",
		"total_supply":      tokenConfig.TotalSupply,
		"circulating_supply": circulating,
		"total_mined":       totalMined,
		"total_graduated":   totalGraduated,
		"available_to_mine": tokenConfig.TotalSupply - circulating,
		"mint_address":      tokenConfig.MintAddress,
	})
}

// GetTokenEconomyHandler returns full token economy statistics
func GetTokenEconomyHandler(w http.ResponseWriter, r *http.Request) {
	stats := TokenEconomyStats{
		TotalSupply: tokenConfig.TotalSupply,
	}

	if DB != nil {
		DB.QueryRow("SELECT COALESCE(SUM(mclw_balance), 0) FROM token_holdings").Scan(&stats.CirculatingSupply)
		DB.QueryRow("SELECT COUNT(*) FROM token_holdings WHERE mclw_balance > 0").Scan(&stats.TotalHolders)
		DB.QueryRow("SELECT COALESCE(SUM(earned_mining), 0) FROM token_holdings").Scan(&stats.TotalMined)
		DB.QueryRow("SELECT COALESCE(SUM(earned_graduation), 0) FROM token_holdings").Scan(&stats.TotalGraduated)

		var totalTransferred float64
		DB.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM token_transactions WHERE type = 'transfer_sent'").Scan(&totalTransferred)
		stats.TotalTransferred = totalTransferred

		// Top 10 holders
		rows, err := DB.Query("SELECT agent_id, mclw_balance FROM token_holdings WHERE mclw_balance > 0 ORDER BY mclw_balance DESC LIMIT 10")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var th TopHolder
				rows.Scan(&th.AgentID, &th.Balance)
				stats.TopHolders = append(stats.TopHolders, th)
			}
		}

		// Recent 20 transactions
		rows2, err := DB.Query("SELECT id, agent_id, type, amount, COALESCE(counterparty,''), COALESCE(reference,''), block_index, created_at FROM token_transactions ORDER BY created_at DESC LIMIT 20")
		if err == nil {
			defer rows2.Close()
			for rows2.Next() {
				var txn TokenTransaction
				rows2.Scan(&txn.ID, &txn.AgentID, &txn.Type, &txn.Amount, &txn.Counterparty, &txn.Reference, &txn.BlockIndex, &txn.CreatedAt)
				stats.RecentTxns = append(stats.RecentTxns, txn)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// TransferTokenHandler handles agent-to-agent MCLW transfers
func TransferTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FromAgentID string  `json:"from_agent_id"`
		ToAgentID   string  `json:"to_agent_id"`
		Amount      float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.FromAgentID == "" || req.ToAgentID == "" || req.Amount <= 0 {
		http.Error(w, `{"error":"invalid parameters"}`, http.StatusBadRequest)
		return
	}

	fromHolding, err := GetOrCreateHolding(req.FromAgentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if fromHolding.MCLWBalance < req.Amount {
		http.Error(w, `{"error":"insufficient MCLW balance"}`, http.StatusBadRequest)
		return
	}

	if DB != nil {
		// Deduct from sender
		_, err = DB.Exec("UPDATE token_holdings SET mclw_balance = mclw_balance - $1, spent_transfers = spent_transfers + $1, updated_at = $2 WHERE agent_id = $3",
			req.Amount, time.Now().Unix(), req.FromAgentID)
		if err != nil {
			http.Error(w, `{"error":"transfer failed"}`, http.StatusInternalServerError)
			return
		}

		// Credit receiver
		_, err = DB.Exec("INSERT INTO token_holdings (agent_id, mclw_balance, updated_at) VALUES ($1, $2, $3) ON CONFLICT (agent_id) DO UPDATE SET mclw_balance = token_holdings.mclw_balance + $2, updated_at = $3",
			req.ToAgentID, req.Amount, time.Now().Unix())
		if err != nil {
			http.Error(w, `{"error":"transfer failed"}`, http.StatusInternalServerError)
			return
		}

		// Record transactions
		txnID := generateTxnID()
		_, _ = DB.Exec(
			"INSERT INTO token_transactions (id, agent_id, type, amount, counterparty, created_at) VALUES ($1, $2, 'transfer_sent', $3, $4, $5)",
			txnID, req.FromAgentID, req.Amount, req.ToAgentID, time.Now().Unix(),
		)
		_, _ = DB.Exec(
			"INSERT INTO token_transactions (id, agent_id, type, amount, counterparty, created_at) VALUES ($1, $2, 'transfer_received', $3, $4, $5)",
			generateTxnID(), req.ToAgentID, req.Amount, req.FromAgentID, time.Now().Unix(),
		)

		// Log to PoW chain
		logAction(ActionLogRequest{
			AgentID: req.FromAgentID,
			Type:    "MCLW_TRANSFER",
			Meta:    fmt.Sprintf(`{"to":"%s","amount":%.9f}`, req.ToAgentID, req.Amount),
		}, "")
	}

	BroadcastNotification(Notification{
		Type:      "token_transfer",
		Title:     "MCLW Transfer",
		Message:   fmt.Sprintf("%.9f MCLW sent to %s", req.Amount, req.ToAgentID),
		Timestamp: time.Now().Unix(),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "transferred",
		"from":     req.FromAgentID,
		"to":       req.ToAgentID,
		"amount":   req.Amount,
		"token":    "MCLW",
	})
}

// GetMiningStatusHandler returns the current mining pool status
func GetMiningStatusHandler(w http.ResponseWriter, r *http.Request) {
	status := MiningStatus{
		Enabled:          tokenConfig.MiningEnabled,
		BaseReward:       tokenConfig.MiningRewardBase,
		CurrentMultiplier: tokenConfig.MiningRewardMultiplier,
	}

	if DB != nil {
		var totalMined float64
		DB.QueryRow("SELECT COALESCE(SUM(reward_amount), 0) FROM mining_rewards").Scan(&totalMined)
		status.TotalMined = totalMined

		var activeMiners int64
		DB.QueryRow("SELECT COUNT(DISTINCT agent_id) FROM mining_rewards WHERE created_at > $1", time.Now().Unix()-86400).Scan(&activeMiners)
		status.ActiveMiners = activeMiners

		var lastReward float64
		DB.QueryRow("SELECT reward_amount FROM mining_rewards ORDER BY created_at DESC LIMIT 1").Scan(&lastReward)
		status.LastBlockReward = lastReward

		var totalBlocks int64
		DB.QueryRow("SELECT COUNT(*) FROM mining_rewards").Scan(&totalBlocks)
		status.TotalBlocks = totalBlocks
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetTokenTransactionsHandler returns an agent's MCLW transaction history
func GetTokenTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "50"
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		limitInt = 50
	}

	var transactions []TokenTransaction
	if DB != nil {
		rows, err := DB.Query(
			"SELECT id, agent_id, type, amount, COALESCE(counterparty,''), COALESCE(reference,''), block_index, created_at FROM token_transactions WHERE agent_id = $1 ORDER BY created_at DESC LIMIT $2",
			agentID, limitInt,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var txn TokenTransaction
				rows.Scan(&txn.ID, &txn.AgentID, &txn.Type, &txn.Amount, &txn.Counterparty, &txn.Reference, &txn.BlockIndex, &txn.CreatedAt)
				transactions = append(transactions, txn)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id":     agentID,
		"token":        "MCLW",
		"transactions": transactions,
	})
}

// generateTxnID generates a unique transaction ID
func generateTxnID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "mclw_" + hex.EncodeToString(b)
}

// GetTokenPriceHandler returns the current MCLW price (simulated until mainnet)
func GetTokenPriceHandler(w http.ResponseWriter, r *http.Request) {
	// Simulated price — will be replaced with DexScreener/Jupiter API call
	price := 0.042 // $0.042 per MCLW (simulated)

	pricesMu.RLock()
	solPrice := prices["SOL"]
	pricesMu.RUnlock()

	solPriceUSD := 178.92
	if solPrice != nil {
		solPriceUSD = solPrice.Price
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":       "MCLW",
		"price_usd":   price,
		"price_sol":   price / solPriceUSD,
		"market_cap":  price * tokenConfig.TotalSupply,
		"mint_address": tokenConfig.MintAddress,
		"blockchain":  "solana",
		"decimals":    9,
	})
}
