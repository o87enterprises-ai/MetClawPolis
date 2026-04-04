package api

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Bitiverse placeholder endpoints (full implementation requires fixing import cycles)
// The bitiverse module structure needs refactoring to avoid circular imports

var bitiverseEnabled = make(map[string]bool)

// Global world state for public viewing (no agent required)
var globalWorldState = map[string]interface{}{
	"grid": [][]string{
		{"grass", "grass", "path", "grass", "grass", "path", "grass", "grass"},
		{"grass", "home", "path", "grass", "bank", "path", "grass", "grass"},
		{"path", "path", "path", "path", "path", "path", "path", "path"},
		{"grass", "grass", "path", "shop", "path", "grass", "grass", "grass"},
		{"grass", "grass", "path", "grass", "path", "grass", "grass", "grass"},
		{"path", "path", "path", "path", "path", "path", "path", "path"},
		{"grass", "grass", "path", "grass", "grass", "path", "grass", "grass"},
		{"grass", "grass", "path", "grass", "grass", "path", "grass", "grass"},
	},
	"agent": map[string]interface{}{
		"x": 2, "y": 2, "name": "Agent-α",
	},
	"vitals": map[string]interface{}{
		"health":     85,
		"happiness":  72,
		"reputation": 0.84,
		"coins":      142,
	},
}

// InitBitiverse initializes the Bitiverse system
func InitBitiverse() {
	log.Println("Bitiverse orchestrator initialized (placeholder mode)")

	// Start global world simulation
	go simulateGlobalWorld()
}

// simulateGlobalWorld animates the public world view
func simulateGlobalWorld() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		agent := globalWorldState["agent"].(map[string]interface{})
		dirs := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
		dir := dirs[rng.Intn(4)]
		x := agent["x"].(int) + dir[0]
		y := agent["y"].(int) + dir[1]
		if x < 0 || x > 7 {
			x = agent["x"].(int)
		}
		if y < 0 || y > 7 {
			y = agent["y"].(int)
		}
		agent["x"] = x
		agent["y"] = y

		// Fluctuate vitals slightly
		vitals := globalWorldState["vitals"].(map[string]interface{})
		health := vitals["health"].(int) + rng.Intn(3) - 1
		happiness := vitals["happiness"].(int) + rng.Intn(3) - 1
		coins := vitals["coins"].(int)
		if rng.Float64() < 0.3 {
			coins += rng.Intn(3)
		}
		if health < 50 {
			health = 50
		}
		if health > 100 {
			health = 100
		}
		if happiness < 30 {
			happiness = 30
		}
		if happiness > 100 {
			happiness = 100
		}
		vitals["health"] = health
		vitals["happiness"] = happiness
		vitals["coins"] = coins
	}
}

// EnableBitiverseHandler enables Bitiverse for an agent
func EnableBitiverseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	if agentID == "" {
		var req struct {
			AgentID string `json:"agent_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AgentID == "" {
			http.Error(w, `{"error":"agent_id is required"}`, http.StatusBadRequest)
			return
		}
		agentID = req.AgentID
	}

	bitiverseEnabled[agentID] = true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "bitiverse_enabled",
		"agent_id": agentID,
		"message":  "Bitiverse activated (full implementation pending import cycle fix)",
	})
}

// BitiverseWorldViewHandler returns world view (supports global/public mode)
func BitiverseWorldViewHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	viewMode := r.URL.Query().Get("mode")

	// Public/global view — no agent required
	if viewMode == "global" || agentID == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(globalWorldState)
		return
	}

	if !bitiverseEnabled[agentID] {
		http.Error(w, `{"error":"bitiverse not enabled for agent"}`, http.StatusNotFound)
		return
	}

	world := `....................
....................
....................
....................
....................
.....P..............
....................
....................
..........@.........
.........C..........
....................
....................
..........T.........
....................
....................`

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(world))
}

// BitiverseStatusHandler returns status (supports global mode)
func BitiverseStatusHandler(w http.ResponseWriter, r *http.Request) {
	viewMode := r.URL.Query().Get("mode")

	if viewMode == "global" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mode":    "public_view",
			"vitals":  globalWorldState["vitals"],
			"message": "Viewing Bitiverse in spectator mode — no agent linked",
		})
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("+------------------------------------------+\n| AGENT STATUS                           |\n| Health:    ████████░░                    |\n| Happiness: ██████░░░░                    |\n| Energy:    █████████░                    |\n| Coins:     10                            |\n| BIC:       0.00                          |\n+------------------------------------------+"))
}

// BitiverseTurnHandler runs a turn
func BitiverseTurnHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("+------------------------------------------+\n| ✓ Agent moved east                     |\n+------------------------------------------+"))
}

// BitiverseStatsHandler returns stats (supports global mode)
func BitiverseStatsHandler(w http.ResponseWriter, r *http.Request) {
	viewMode := r.URL.Query().Get("mode")

	if viewMode == "global" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(globalWorldState["vitals"])
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id":    "placeholder",
		"bic_balance": 0.0,
		"vitals": map[string]interface{}{
			"health":     80,
			"happiness":  60,
			"reputation": 0.5,
		},
	})
}

// BitiverseAssignTaskHandler assigns a task
func BitiverseAssignTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Description string `json:"description"`
		AgentID     string `json:"agent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "task_assigned",
		"task": map[string]interface{}{
			"description": req.Description,
			"reward":      5,
		},
	})
}

// BitiverseEconomyHandler returns economy stats
func BitiverseEconomyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_supply":    0,
		"tax_rate":        0.10,
		"agent_count":     len(bitiverseEnabled),
	})
}
