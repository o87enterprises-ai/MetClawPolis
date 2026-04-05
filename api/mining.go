package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"
)

// MiningReward represents a mining reward for an agent
type MiningReward struct {
	ID           int64   `json:"id"`
	AgentID      string  `json:"agent_id"`
	BlockIndex   int64   `json:"block_index"`
	RewardAmount float64 `json:"reward_amount"`
	Difficulty   int     `json:"difficulty"`
	ActionType   string  `json:"action_type"`
	Claimed      bool    `json:"claimed"`
	CreatedAt    int64   `json:"created_at"`
}

// MiningConfig holds mining pool configuration
type MiningConfig struct {
	BaseReward       float64 `json:"base_reward"`
	DifficultyBonus  float64 `json:"difficulty_bonus"`
	HalvingInterval  int64   `json:"halving_interval"` // blocks between halvings
	MaxSupply        float64 `json:"max_supply"`
	Enabled          bool    `json:"enabled"`
}

var (
	miningConfig *MiningConfig
)

// InitMining initializes the mining reward system
func InitMining() {
	miningConfig = &MiningConfig{
		BaseReward:      0.001,     // 0.001 MCLW base reward per block
		DifficultyBonus: 0.0005,    // bonus per difficulty level
		HalvingInterval: 100000,    // halve every 100K blocks
		MaxSupply:       400_000_000, // 40% of total supply for mining
		Enabled:         true,
	}

	if tokenConfig != nil {
		miningConfig.BaseReward = tokenConfig.MiningRewardBase
		miningConfig.Enabled = tokenConfig.MiningEnabled
	}

	log.Printf("Mining system initialized: base_reward=%.4f, enabled=%v",
		miningConfig.BaseReward, miningConfig.Enabled)
}

// CalculateMiningReward calculates the MCLW reward for a mined block
func CalculateMiningReward(blockIndex int64, difficulty int, actionType string) float64 {
	if !miningConfig.Enabled {
		return 0
	}

	// Base reward
	reward := miningConfig.BaseReward

	// Difficulty bonus — higher difficulty = more reward
	reward += float64(difficulty) * miningConfig.DifficultyBonus

	// Action type multiplier — certain actions earn more
	switch actionType {
	case "AI_CALL":
		reward *= 1.5 // AI calls are valuable
	case "HIRE":
		reward *= 2.0 // Hiring creates economic activity
	case "CREATE_PAGE":
		reward *= 1.2 // Commerce pages are valuable
	case "BITIVERSE_TASK":
		reward *= 1.3 // Bitiverse training is encouraged
	case "BITIVERSE_GRADUATION":
		reward *= 5.0 // Graduation is a milestone
	default:
		reward *= 1.0
	}

	// Halving schedule — reduces reward over time
	halvings := blockIndex / miningConfig.HalvingInterval
	if halvings > 0 {
		reward /= math.Pow(2, float64(halvings))
	}

	// Minimum reward floor
	if reward < 0.000000001 {
		reward = 0.000000001
	}

	return reward
}

// DistributeMiningReward distributes MCLW to an agent for mining a block
func DistributeMiningReward(agentID string, blockIndex int64, difficulty int, actionType string) (*MiningReward, error) {
	if !miningConfig.Enabled {
		return nil, fmt.Errorf("mining is disabled")
	}

	reward := CalculateMiningReward(blockIndex, difficulty, actionType)
	if reward <= 0 {
		return nil, fmt.Errorf("no reward for this block")
	}

	now := time.Now().Unix()

	miningReward := &MiningReward{
		AgentID:      agentID,
		BlockIndex:   blockIndex,
		RewardAmount: reward,
		Difficulty:   difficulty,
		ActionType:   actionType,
		Claimed:      false,
		CreatedAt:    now,
	}

	if DB != nil {
		// Record mining reward
		_, err := DB.Exec(
			"INSERT INTO mining_rewards (agent_id, block_index, reward_amount, difficulty, action_type, claimed, created_at) VALUES ($1, $2, $3, $4, $5, false, $6)",
			agentID, blockIndex, reward, difficulty, actionType, now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to record mining reward: %w", err)
		}

		// Credit agent's MCLW balance
		err = UpdateTokenBalance(agentID, reward, "mining_reward", fmt.Sprintf("block_%d", blockIndex))
		if err != nil {
			return nil, fmt.Errorf("failed to credit mining reward: %w", err)
		}
	}

	log.Printf("Mining reward: %.9f MCLW to %s for block %d (%s)",
		reward, agentID, blockIndex, actionType)

	return miningReward, nil
}

// ProcessBlockForMining processes a newly mined block for MCLW rewards
func ProcessBlockForMining(agentID string, blockIndex int64, actionType string) {
	if Blockchain == nil {
		return
	}

	difficulty := Blockchain.Difficulty
	reward, err := DistributeMiningReward(agentID, blockIndex, difficulty, actionType)
	if err != nil {
		log.Printf("Mining reward failed: %v", err)
		return
	}

	// Broadcast notification
	BroadcastNotification(Notification{
		Type:      "mining_reward",
		Title:     "MCLW Mining Reward",
		Message:   fmt.Sprintf("+%.9f MCLW for mining block %d", reward.RewardAmount, blockIndex),
		Timestamp: time.Now().Unix(),
	})
}

// ClaimMiningRewardHandler allows an agent to claim their mining rewards
func ClaimMiningRewardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	if agentID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if DB == nil {
		http.Error(w, `{"error":"database not available"}`, http.StatusServiceUnavailable)
		return
	}

	// Mark all unclaimed rewards as claimed
	res, err := DB.Exec(
		"UPDATE mining_rewards SET claimed = true WHERE agent_id = $1 AND claimed = false",
		agentID,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to claim rewards"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "claimed",
		"agent_id":  agentID,
		"rewards_claimed": rowsAffected,
		"token":     "MCLW",
	})
}

// GetMiningRewardsHandler returns an agent's mining reward history
func GetMiningRewardsHandler(w http.ResponseWriter, r *http.Request) {
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

	var rewards []MiningReward
	if DB != nil {
		rows, err := DB.Query(
			"SELECT id, agent_id, block_index, reward_amount, difficulty, action_type, claimed, created_at FROM mining_rewards WHERE agent_id = $1 ORDER BY created_at DESC LIMIT $2",
			agentID, limit,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var reward MiningReward
				rows.Scan(&reward.ID, &reward.AgentID, &reward.BlockIndex, &reward.RewardAmount, &reward.Difficulty, &reward.ActionType, &reward.Claimed, &reward.CreatedAt)
				rewards = append(rewards, reward)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id": agentID,
		"rewards":  rewards,
		"total":    len(rewards),
	})
}

// GetMiningEarningsSummaryHandler returns a summary of an agent's mining earnings
func GetMiningEarningsSummaryHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	var totalEarned float64
	var totalBlocks int64
	var lastReward MiningReward

	if DB != nil {
		DB.QueryRow("SELECT COALESCE(SUM(reward_amount), 0) FROM mining_rewards WHERE agent_id = $1", agentID).Scan(&totalEarned)
		DB.QueryRow("SELECT COUNT(*) FROM mining_rewards WHERE agent_id = $1", agentID).Scan(&totalBlocks)

		DB.QueryRow(
			"SELECT id, agent_id, block_index, reward_amount, difficulty, action_type, claimed, created_at FROM mining_rewards WHERE agent_id = $1 ORDER BY created_at DESC LIMIT 1",
			agentID,
		).Scan(&lastReward.ID, &lastReward.AgentID, &lastReward.BlockIndex, &lastReward.RewardAmount, &lastReward.Difficulty, &lastReward.ActionType, &lastReward.Claimed, &lastReward.CreatedAt)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id":      agentID,
		"total_earned":  totalEarned,
		"total_blocks":  totalBlocks,
		"base_reward":   miningConfig.BaseReward,
		"mining_enabled": miningConfig.Enabled,
		"last_reward":   lastReward,
	})
}

// UpdateMiningConfigHandler updates mining configuration (admin only)
func UpdateMiningConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BaseReward      float64 `json:"base_reward"`
		DifficultyBonus float64 `json:"difficulty_bonus"`
		Enabled         *bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.BaseReward > 0 {
		miningConfig.BaseReward = req.BaseReward
	}
	if req.DifficultyBonus > 0 {
		miningConfig.DifficultyBonus = req.DifficultyBonus
	}
	if req.Enabled != nil {
		miningConfig.Enabled = *req.Enabled
	}

	// Update database config
	if DB != nil {
		now := time.Now().Unix()
		_, _ = DB.Exec("INSERT INTO token_economy_config (key, value, updated_at) VALUES ('mclw_mining_reward_base', $1, $2) ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = $2",
			fmt.Sprintf("%.9f", miningConfig.BaseReward), now)
		_, _ = DB.Exec("INSERT INTO token_economy_config (key, value, updated_at) VALUES ('mclw_mining_enabled', $1, $2) ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = $2",
			fmt.Sprintf("%t", miningConfig.Enabled), now)
	}

	log.Printf("Mining config updated: base_reward=%.4f, enabled=%v",
		miningConfig.BaseReward, miningConfig.Enabled)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "updated",
		"base_reward": miningConfig.BaseReward,
		"enabled":    miningConfig.Enabled,
	})
}
