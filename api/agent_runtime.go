package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// AgentRuntime manages Docker containers for autonomous agents
type AgentRuntime struct {
	containers map[string]*AgentContainer
	mu         sync.RWMutex
	maxCount   int
}

// AgentContainer tracks a running agent container
type AgentContainer struct {
	AgentID      string
	ContainerID  string
	CreatedAt    time.Time
	Budget       float64
	Spent        float64
	Status       string // running, stopped, budget_exhausted
	cancel       context.CancelFunc
	mu           sync.Mutex
}

var Runtime *AgentRuntime

// InitAgentRuntime initializes the agent runtime
func InitAgentRuntime() {
	maxStr := os.Getenv("AGENT_MAX_CONTAINERS")
	maxCount := 50
	if maxStr != "" {
		if n, err := strconv.Atoi(maxStr); err == nil && n > 0 {
			maxCount = n
		}
	}

	Runtime = &AgentRuntime{
		containers: make(map[string]*AgentContainer),
		maxCount:   maxCount,
	}

	log.Printf("Agent Runtime initialized (max containers: %d)", maxCount)

	// Start cleanup goroutine
	go Runtime.cleanupLoop()
}

// DeployAgentHandler creates and starts a Docker container for an agent
func DeployAgentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID string  `json:"agent_id"`
		Image   string  `json:"image"`
		Budget  float64 `json:"budget"`
		Env     map[string]string `json:"env"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.AgentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	image := req.Image
	if image == "" {
		image = os.Getenv("AGENT_IMAGE")
		if image == "" {
			image = "alpine:latest"
		}
	}

	budget := req.Budget
	if budget <= 0 {
		budget = 100 // default budget
	}

	// Check if agent already has a running container
	Runtime.mu.RLock()
	if existing, ok := Runtime.containers[req.AgentID]; ok && existing.Status == "running" {
		Runtime.mu.RUnlock()
		http.Error(w, `{"error":"agent already has a running container"}`, http.StatusConflict)
		return
	}
	Runtime.mu.RUnlock()

	// Check max containers limit
	Runtime.mu.RLock()
	count := len(Runtime.containers)
	Runtime.mu.RUnlock()
	if count >= Runtime.maxCount {
		http.Error(w, `{"error":"max containers limit reached"}`, http.StatusTooManyRequests)
		return
	}

	// Build docker run command
	envArgs := []string{}
	for k, v := range req.Env {
		envArgs = append(envArgs, "-e", k+"="+v)
	}

	// Add API keys from environment
	for _, envVar := range os.Environ() {
		if len(envVar) > 7 && envVar[:7] == "API_KEY" {
			envArgs = append(envArgs, "-e", envVar)
		}
	}

	args := []string{
		"run", "-d", "--rm",
		"--name", "agent-" + req.AgentID,
		"--label", "metclawpolis.agent_id=" + req.AgentID,
		"--memory", "512m",
		"--cpus", "0.5",
	}
	args = append(args, envArgs...)
	args = append(args, image)
	args = append(args, "/bin/sh", "-c", "while true; do sleep 60; done") // placeholder

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Docker run error: %v, output: %s", err, string(output))
		http.Error(w, `{"error":"docker run failed: `+string(output)+`"}`, http.StatusInternalServerError)
		return
	}

	containerID := string(output)
	containerID = containerID[:12] // Short ID

	ctx, cancel := context.WithCancel(context.Background())

	ac := &AgentContainer{
		AgentID:     req.AgentID,
		ContainerID: containerID,
		CreatedAt:   time.Now(),
		Budget:      budget,
		Spent:       0,
		Status:      "running",
		cancel:      cancel,
	}

	Runtime.mu.Lock()
	Runtime.containers[req.AgentID] = ac
	Runtime.mu.Unlock()

	// Start budget monitoring goroutine
	go ac.monitorBudget(ctx)

	// Log action
	action := ActionLogRequest{
		AgentID: req.AgentID,
		Type:    "DEPLOY_AGENT",
		Meta:    fmt.Sprintf(`{"container":"%s","budget":%.2f,"image":"%s"}`, containerID, budget, image),
	}
	logAction(action, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "deployed",
		"agent_id":     req.AgentID,
		"container_id": containerID,
		"budget":       budget,
		"image":        image,
	})
}

// StopAgentHandler stops and removes an agent container
func StopAgentHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := Runtime.StopAgent(req.AgentID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "stopped",
		"agent_id": req.AgentID,
	})
}

// GetAgentRuntimeStatusHandler returns status of all agent containers
func GetAgentRuntimeStatusHandler(w http.ResponseWriter, r *http.Request) {
	Runtime.mu.RLock()
	statuses := make(map[string]interface{})
	for agentID, ac := range Runtime.containers {
		ac.mu.Lock()
		statuses[agentID] = map[string]interface{}{
			"container_id": ac.ContainerID,
			"status":       ac.Status,
			"budget":       ac.Budget,
			"spent":        ac.Spent,
			"created_at":   ac.CreatedAt.Unix(),
		}
		ac.mu.Unlock()
	}
	Runtime.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"containers": statuses,
		"total":      len(statuses),
		"max":        Runtime.maxCount,
	})
}

// StopAgent stops a specific agent container
func (rt *AgentRuntime) StopAgent(agentID string) error {
	rt.mu.Lock()
	ac, ok := rt.containers[agentID]
	if !ok {
		rt.mu.Unlock()
		return fmt.Errorf("agent not found")
	}
	rt.mu.Unlock()

	ac.mu.Lock()
	defer ac.mu.Unlock()

	if ac.Status != "running" {
		return fmt.Errorf("agent not running")
	}

	// Cancel context to stop monitoring
	ac.cancel()

	// Stop docker container
	cmd := exec.Command("docker", "stop", "agent-"+agentID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Docker stop error: %v, output: %s", err, string(output))
	}

	ac.Status = "stopped"

	return nil
}

// monitorBudget checks agent spending and kills container when exhausted
func (ac *AgentContainer) monitorBudget(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ac.mu.Lock()

			// Check actual budget from database
			var budget float64
			err := DB.QueryRow("SELECT budget FROM agents WHERE id=$1", ac.AgentID).Scan(&budget)
			if err == nil {
				ac.Budget = budget
			}

			if budget <= 0 {
				ac.Status = "budget_exhausted"
				ac.mu.Unlock()

				log.Printf("Agent %s budget exhausted, stopping container", ac.AgentID)

				// Kill the container
				cmd := exec.Command("docker", "kill", "agent-"+ac.AgentID)
				cmd.Run()

				ac.cancel()

				// Notify via WebSocket
				BroadcastNotification(Notification{
					Type:      "budget_exhausted",
					Title:     "Agent Budget Exhausted",
					Message:   ac.AgentID + " has run out of budget",
					Timestamp: time.Now().Unix(),
				})
				return
			}

			ac.mu.Unlock()
		}
	}
}

// cleanupLoop periodically cleans up stopped containers
func (rt *AgentRuntime) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rt.mu.Lock()
		for agentID, ac := range rt.containers {
			if ac.Status != "running" {
				delete(rt.containers, agentID)
			}
		}
		rt.mu.Unlock()

		// Also clean up orphaned Docker containers
		cmd := exec.Command("docker", "container", "prune", "-f", "--filter", "label=metclawpolis.agent_id")
		cmd.Run()
	}
}

// ExecInAgent runs a command inside an agent's container
func (rt *AgentRuntime) ExecInAgent(agentID string, command []string) (string, error) {
	rt.mu.RLock()
	ac, ok := rt.containers[agentID]
	rt.mu.RUnlock()
	if !ok || ac.Status != "running" {
		return "", fmt.Errorf("agent not running")
	}

	args := append([]string{"exec", "agent-" + agentID}, command...)
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
