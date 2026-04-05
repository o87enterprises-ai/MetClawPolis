package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/metclawpolis/agent-platform/bitiverse/types"
)

// GraduationRequest represents a request to graduate from Bitiverse
type GraduationRequest struct {
	AgentID string `json:"agent_id"`
}

// GraduationResult represents the result of a Bitiverse graduation
type GraduationResult struct {
	AgentID        string  `json:"agent_id"`
	BICGraduated   float64 `json:"bic_graduated"`
	MCLWReceived   float64 `json:"mclw_received"`
	ExchangeRate   float64 `json:"exchange_rate"`
	GraduationBonus float64 `json:"graduation_bonus"`
	MoralScore     float64 `json:"moral_score"`
	TrainingScore  float64 `json:"training_score"`
	CompletedLessons int   `json:"completed_lessons"`
	GraduatedAt    int64   `json:"graduated_at"`
	TxSignature    string  `json:"tx_signature,omitempty"`
}

// BICMCLWExchangeRate returns the current BIC to MCLW exchange rate
func BICMCLWExchangeRate() float64 {
	if tokenConfig == nil {
		return 0.01 // default: 100 BIC = 1 MCLW
	}
	return tokenConfig.BICExchangeRate
}

// CalculateGraduationReward calculates the MCLW reward for Bitiverse graduation
func CalculateGraduationReward(bicBalance float64, moralScore float64, trainingScore float64, completedLessons int) float64 {
	// Base conversion: BIC → MCLW
	baseReward := bicBalance * BICMCLWExchangeRate()

	// Moral score multiplier (0.0–1.0 → 0.5x–2.0x)
	moralMultiplier := 0.5 + moralScore*1.5

	// Training score multiplier (0.0–1.0 → 0.8x–1.5x)
	trainingMultiplier := 0.8 + trainingScore*0.7

	// Lesson completion bonus (0.01 MCLW per lesson)
	lessonBonus := float64(completedLessons) * 0.01

	// Total reward
	total := baseReward * moralMultiplier * trainingMultiplier * lessonBonus

	// Add graduation bonus
	if tokenConfig != nil {
		total += tokenConfig.GraduationBonus
	}

	// Cap at reasonable maximum
	if total > 10000 {
		total = 10000
	}

	return math.Round(total*1e9) / 1e9 // Round to 9 decimals
}

// GraduateAgent handles Bitiverse graduation: BIC → MCLW bridge
func GraduateAgent(agentID string) (*GraduationResult, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not available")
	}

	// Get agent's Bitiverse state (simulated — in production, load from Bitiverse runtime)
	bicBalance := 500.0  // simulated BIC balance
	moralScore := 0.85   // simulated moral score
	trainingScore := 0.78 // simulated training score
	completedLessons := 12 // simulated completed lessons

	// Check minimum graduation amount
	if tokenConfig != nil && bicBalance < tokenConfig.MinGraduationAmount {
		return nil, fmt.Errorf("insufficient BIC balance: need %.0f BIC minimum", tokenConfig.MinGraduationAmount)
	}

	// Check if already graduated
	var alreadyGraduated bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM graduation_records WHERE agent_id = $1)", agentID).Scan(&alreadyGraduated)
	if err != nil {
		return nil, fmt.Errorf("failed to check graduation status: %w", err)
	}
	if alreadyGraduated {
		return nil, fmt.Errorf("agent already graduated")
	}

	// Calculate reward
	mclwReward := CalculateGraduationReward(bicBalance, moralScore, trainingScore, completedLessons)

	now := time.Now().Unix()

	result := &GraduationResult{
		AgentID:        agentID,
		BICGraduated:   bicBalance,
		MCLWReceived:   mclwReward,
		ExchangeRate:   BICMCLWExchangeRate(),
		GraduationBonus: tokenConfig.GraduationBonus,
		MoralScore:     moralScore,
		TrainingScore:  trainingScore,
		CompletedLessons: completedLessons,
		GraduatedAt:    now,
	}

	// Record graduation
	_, err = DB.Exec(
		"INSERT INTO graduation_records (agent_id, bic_graduated, mclw_received, exchange_rate, moral_score, training_score, completed_lessons, graduated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		agentID, bicBalance, mclwReward, BICMCLWExchangeRate(), moralScore, trainingScore, completedLessons, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record graduation: %w", err)
	}

	// Credit MCLW to agent
	err = UpdateTokenBalance(agentID, mclwReward, "graduation_reward", "bitiverse_graduation")
	if err != nil {
		return nil, fmt.Errorf("failed to credit MCLW: %w", err)
	}

	// Log to PoW chain
	logAction(ActionLogRequest{
		AgentID: agentID,
		Type:    "BITIVERSE_GRADUATION",
		Meta: fmt.Sprintf(`{"bic_graduated":%.2f,"mclw_received":%.9f,"moral_score":%.4f,"training_score":%.4f}`,
			bicBalance, mclwReward, moralScore, trainingScore),
	}, "")

	// Broadcast notification
	BroadcastNotification(Notification{
		Type:      "graduation",
		Title:     "Bitiverse Graduation Complete",
		Message:   fmt.Sprintf("%s graduated from Bitiverse! +%.9f MCLW earned", agentID, mclwReward),
		Timestamp: now,
	})

	log.Printf("Graduation: %s graduated with %.2f BIC → %.9f MCLW (moral=%.2f, training=%.2f)",
		agentID, bicBalance, mclwReward, moralScore, trainingScore)

	return result, nil
}

// GraduateFromBitiverseHandler handles the graduation endpoint
func GraduateFromBitiverseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req GraduationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	agentID := req.AgentID
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	result, err := GraduateAgent(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetGraduationStatusHandler returns an agent's graduation status
func GetGraduationStatusHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	var graduated bool
	var record struct {
		BICGraduated   float64
		MCLWReceived   float64
		ExchangeRate   float64
		MoralScore     float64
		TrainingScore  float64
		CompletedLessons int
		GraduatedAt    int64
	}

	if DB != nil {
		err := DB.QueryRow(
			"SELECT bic_graduated, mclw_received, exchange_rate, moral_score, training_score, completed_lessons, graduated_at FROM graduation_records WHERE agent_id = $1",
			agentID,
		).Scan(&record.BICGraduated, &record.MCLWReceived, &record.ExchangeRate, &record.MoralScore, &record.TrainingScore, &record.CompletedLessons, &record.GraduatedAt)
		if err == nil {
			graduated = true
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id":   agentID,
		"graduated":  graduated,
		"record":     record,
		"exchange_rate": BICMCLWExchangeRate(),
		"min_graduation": tokenConfig.MinGraduationAmount,
		"graduation_bonus": tokenConfig.GraduationBonus,
	})
}

// GetBitiverseEconomyHandler returns Bitiverse economy statistics
func GetBitiverseEconomyHandler(w http.ResponseWriter, r *http.Request) {
	var stats struct {
		TotalGraduated     int64   `json:"total_graduated"`
		TotalBICGraduated  float64 `json:"total_bic_graduated"`
		TotalMCLWDistributed float64 `json:"total_mclw_distributed"`
		AverageMCLWPerGrad float64 `json:"avg_mclw_per_graduation"`
		TopGraduates       []interface{} `json:"top_graduates"`
		ExchangeRate       float64 `json:"exchange_rate"`
	}

	stats.ExchangeRate = BICMCLWExchangeRate()

	if DB != nil {
		DB.QueryRow("SELECT COUNT(*) FROM graduation_records").Scan(&stats.TotalGraduated)
		DB.QueryRow("SELECT COALESCE(SUM(bic_graduated), 0) FROM graduation_records").Scan(&stats.TotalBICGraduated)
		DB.QueryRow("SELECT COALESCE(SUM(mclw_received), 0) FROM graduation_records").Scan(&stats.TotalMCLWDistributed)

		if stats.TotalGraduated > 0 {
			stats.AverageMCLWPerGrad = stats.TotalMCLWDistributed / float64(stats.TotalGraduated)
		}

		// Top 10 graduates by MCLW earned
		rows, err := DB.Query(
			"SELECT agent_id, mclw_received, moral_score, training_score, graduated_at FROM graduation_records ORDER BY mclw_received DESC LIMIT 10",
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var agentID string
				var mclw, moral, training float64
				var graduatedAt int64
				rows.Scan(&agentID, &mclw, &moral, &training, &graduatedAt)
				stats.TopGraduates = append(stats.TopGraduates, map[string]interface{}{
					"agent_id":      agentID,
					"mclw_received": mclw,
					"moral_score":   moral,
					"training_score": training,
					"graduated_at":  graduatedAt,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetBitiverseAgentInventoryHandler returns an agent's Bitiverse inventory with MCLW
func GetBitiverseAgentInventoryHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	// Get BIC balance from Bitiverse state (simulated)
	bicBalance := 500.0

	// Get MCLW balance
	mclwBalance := 0.0
	if holding, err := GetOrCreateHolding(agentID); err == nil {
		mclwBalance = holding.MCLWBalance
	}

	inventory := &types.Inventory{
		Coins:      int(bicBalance * 100), // coins are cents
		BicBalance: bicBalance,
		Items:      []types.InventoryItem{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id":    agentID,
		"inventory":   inventory,
		"mclw_balance": mclwBalance,
		"exchange_rate": BICMCLWExchangeRate(),
		"can_graduate": bicBalance >= tokenConfig.MinGraduationAmount,
	})
}
