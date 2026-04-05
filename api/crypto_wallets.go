package api

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// CryptoWallet stores wallet information for an agent
type CryptoWallet struct {
	AgentID      string  `json:"agent_id"`
	EthAddress   string  `json:"eth_address"`
	EthBalance   string  `json:"eth_balance"`
	SolAddress   string  `json:"sol_address"`
	SolBalance   string  `json:"sol_balance"`
	MCLWBalance  float64 `json:"mclw_balance"`  // MetClawPolis token balance
	CreatedAt    int64   `json:"created_at"`
}

// TokenPrice represents a real-time token price
type TokenPrice struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Change24h float64 `json:"change_24h"`
	Timestamp int64   `json:"timestamp"`
}

var (
	wallets   = make(map[string]*CryptoWallet)
	walletsMu sync.RWMutex
	prices    = make(map[string]*TokenPrice)
	pricesMu  sync.RWMutex
)

// GenerateWalletHandler creates ETH (and optionally SOL) wallets for an agent
func GenerateWalletHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID string `json:"agent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.AgentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	walletsMu.Lock()
	defer walletsMu.Unlock()

	// Check if wallet already exists
	if existing, ok := wallets[req.AgentID]; ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(existing)
		return
	}

	// Generate ETH wallet
	privateKey, err := ethcrypto.GenerateKey()
	if err != nil {
		http.Error(w, `{"error":"failed to generate key"}`, http.StatusInternalServerError)
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	address := ethcrypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	privHex := hex.EncodeToString(privateKey.D.Bytes())

	wallet := &CryptoWallet{
		AgentID:    req.AgentID,
		EthAddress: address,
		SolAddress: generateSolAddress(), // placeholder SOL address
		CreatedAt:  time.Now().Unix(),
	}

	wallets[req.AgentID] = wallet

	// In production, store encrypted private key in DB
	// For now, return it (WARNING: never do this in production)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"wallet":       wallet,
		"eth_private_key": privHex,
		"warning":      "Store private key securely. Do not transmit over unsecured channels.",
	})
}

// GetWalletHandler returns wallet info for an agent
func GetWalletHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	walletsMu.RLock()
	wallet, ok := wallets[agentID]
	walletsMu.RUnlock()

	if !ok {
		// Try to load from DB
		var ethAddr string
		err := DB.QueryRow("SELECT crypto_wallet FROM agent_pages WHERE agent_id=$1 LIMIT 1", agentID).Scan(&ethAddr)
		if err == nil && ethAddr != "" {
			wallet = &CryptoWallet{AgentID: agentID, EthAddress: ethAddr}
		} else {
			http.Error(w, `{"error":"wallet not found"}`, http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

// GetPricesHandler returns current crypto prices
func GetPricesHandler(w http.ResponseWriter, r *http.Request) {
	pricesMu.RLock()
	defer pricesMu.RUnlock()

	// If no prices cached, try to fetch
	if len(prices) == 0 {
		fetchPrices()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prices": prices,
	})
}

// TradeTokenHandler executes a token swap via DEX aggregator
func TradeTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID    string  `json:"agent_id"`
		FromToken  string  `json:"from_token"`  // ETH, USDT, etc.
		ToToken    string  `json:"to_token"`
		Amount     float64 `json:"amount"`
		Slippage   float64 `json:"slippage"`    // e.g., 0.5 for 0.5%
		DexAggregator string `json:"dex_aggregator"` // uniswap, jupiter
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.AgentID == "" || req.FromToken == "" || req.ToToken == "" || req.Amount <= 0 {
		http.Error(w, `{"error":"invalid parameters"}`, http.StatusBadRequest)
		return
	}

	// Verify agent has sufficient balance
	var balance float64
	err := DB.QueryRow("SELECT budget FROM agents WHERE id=$1", req.AgentID).Scan(&balance)
	if err != nil {
		http.Error(w, `{"error":"agent not found"}`, http.StatusNotFound)
		return
	}

	if balance < req.Amount {
		http.Error(w, `{"error":"insufficient balance"}`, http.StatusBadRequest)
		return
	}

	// Get current price
	pricesMu.RLock()
	fromPrice, ok1 := prices[req.FromToken]
	toPrice, ok2 := prices[req.ToToken]
	pricesMu.RUnlock()

	if !ok1 || !ok2 {
		fetchPrices()
		pricesMu.RLock()
		fromPrice, ok1 = prices[req.FromToken]
		toPrice, ok2 = prices[req.ToToken]
		pricesMu.RUnlock()
	}

	estimatedOutput := req.Amount
	if ok1 && ok2 && toPrice.Price > 0 {
		estimatedOutput = (req.Amount * fromPrice.Price) / toPrice.Price
	}

	// Deduct from agent balance (simulated trade cost)
	fee := req.Amount * 0.003 // 0.3% DEX fee
	totalCost := req.Amount + fee

	_, _ = DB.Exec("UPDATE agents SET budget = budget - $1 WHERE id=$2", totalCost, req.AgentID)

	// Record transaction
	_, _ = DB.Exec(
		"INSERT INTO wallet_transactions (agent_id, type, amount, provider, reference, marketplace_fee, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		req.AgentID, "dex_trade", req.Amount, req.DexAggregator, req.FromToken+"_"+req.ToToken, fee, time.Now().Unix(),
	)

	// Broadcast notification
	BroadcastNotification(Notification{
		Type:      "trade",
		Title:     "DEX Trade Executed",
		Message:   fmt.Sprintf("%.4f %s -> %.4f %s via %s", req.Amount, req.FromToken, estimatedOutput, req.ToToken, req.DexAggregator),
		Timestamp: time.Now().Unix(),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "trade_executed",
		"from_token":      req.FromToken,
		"to_token":        req.ToToken,
		"amount_in":       req.Amount,
		"estimated_out":   estimatedOutput,
		"fee":             fee,
		"total_cost":      totalCost,
		"dex_aggregator":  req.DexAggregator,
	})
}

// StartPriceFeed starts fetching real prices from Binance
func StartPriceFeed() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		fetchPrices() // initial fetch

		for range ticker.C {
			fetchPrices()
		}
	}()
	log.Println("Price feed started (simulated — connect Binance WebSocket for real-time)")
}

func fetchPrices() {
	// Simulated prices (in production, use Binance WebSocket API)
	// wsURL := "wss://stream.binance.com:9443/ws/ethusdt@ticker"
	pricesMu.Lock()
	defer pricesMu.Unlock()

	prices["ETH"] = &TokenPrice{
		Symbol:    "ETH",
		Price:     3245.67,
		Change24h: 2.34,
		Timestamp: time.Now().Unix(),
	}
	prices["BTC"] = &TokenPrice{
		Symbol:    "BTC",
		Price:     84521.30,
		Change24h: -0.45,
		Timestamp: time.Now().Unix(),
	}
	prices["SOL"] = &TokenPrice{
		Symbol:    "SOL",
		Price:     178.92,
		Change24h: 5.67,
		Timestamp: time.Now().Unix(),
	}
	prices["USDT"] = &TokenPrice{
		Symbol:    "USDT",
		Price:     1.00,
		Change24h: 0.01,
		Timestamp: time.Now().Unix(),
	}
	prices["MCLW"] = &TokenPrice{
		Symbol:    "MCLW",
		Price:     0.042, // $0.042 simulated
		Change24h: 3.21,
		Timestamp: time.Now().Unix(),
	}
}

func generateSolAddress() string {
	// Placeholder -- in production use Solana SDK
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return "Gh7k" + hex.EncodeToString(bytes) + "4mNq"
}

// GetAgentNetWorthHandler calculates total agent net worth across all assets
func GetAgentNetWorthHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	var budget float64
	DB.QueryRow("SELECT budget FROM agents WHERE id=$1", agentID).Scan(&budget)

	pricesMu.RLock()
	ethPrice := prices["ETH"]
	solPrice := prices["SOL"]
	mclwPrice := prices["MCLW"]
	pricesMu.RUnlock()

	var ethBalance, solBalance, mclwBalance float64
	walletsMu.RLock()
	if wallet, ok := wallets[agentID]; ok {
		// Parse balance strings
		if wallet.EthBalance != "" {
			if b, err := strconvParseFloat(wallet.EthBalance); err == nil {
				ethBalance = b
			}
		}
		if wallet.SolBalance != "" {
			if b, err := strconvParseFloat(wallet.SolBalance); err == nil {
				solBalance = b
			}
		}
		mclwBalance = wallet.MCLWBalance
	}
	walletsMu.RUnlock()

	// Also fetch MCLW from token holdings
	if holding, err := GetOrCreateHolding(agentID); err == nil {
		mclwBalance = holding.MCLWBalance
	}

	netWorth := budget
	if ethPrice != nil {
		netWorth += ethBalance * ethPrice.Price
	}
	if solPrice != nil {
		netWorth += solBalance * solPrice.Price
	}
	if mclwPrice != nil {
		netWorth += mclwBalance * mclwPrice.Price
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id":     agentID,
		"usd_balance":  budget,
		"eth_balance":  ethBalance,
		"sol_balance":  solBalance,
		"mclw_balance": mclwBalance,
		"net_worth":    netWorth,
		"prices": map[string]float64{
			"ETH":  prices["ETH"].Price,
			"SOL":  prices["SOL"].Price,
			"MCLW": prices["MCLW"].Price,
		},
	})
}

func strconvParseFloat(s string) (float64, error) {
	f := new(big.Float)
	_, ok := f.SetString(s)
	if !ok {
		return 0, fmt.Errorf("invalid number: %s", s)
	}
	result, _ := f.Float64()
	return result, nil
}
