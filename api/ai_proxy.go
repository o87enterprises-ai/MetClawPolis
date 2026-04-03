package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/metclawpolis/agent-platform/chain"
)

// AIProvider represents a supported AI API provider
type AIProvider struct {
	Name        string  `json:"name"`
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model"`
	InputCost   float64 `json:"input_cost_per_m_tokens"`  // per 1M tokens
	OutputCost  float64 `json:"output_cost_per_m_tokens"` // per 1M tokens
}

// Supported AI providers
var AIProviders = map[string]AIProvider{
	"openai-gpt4o": {
		Name:       "OpenAI GPT-4o",
		BaseURL:    "https://api.openai.com/v1/chat/completions",
		Model:      "gpt-4o",
		InputCost:  2.50,
		OutputCost: 10.00,
	},
	"openai-gpt4o-mini": {
		Name:       "OpenAI GPT-4o Mini",
		BaseURL:    "https://api.openai.com/v1/chat/completions",
		Model:      "gpt-4o-mini",
		InputCost:  0.15,
		OutputCost: 0.60,
	},
	"openai-o3-mini": {
		Name:       "OpenAI o3 Mini",
		BaseURL:    "https://api.openai.com/v1/chat/completions",
		Model:      "o3-mini",
		InputCost:  1.10,
		OutputCost: 4.40,
	},
	"anthropic-claude-3-7": {
		Name:       "Anthropic Claude 3.7",
		BaseURL:    "https://api.anthropic.com/v1/messages",
		Model:      "claude-3-7-sonnet-20250219",
		InputCost:  3.00,
		OutputCost: 15.00,
	},
	"anthropic-claude-haiku": {
		Name:       "Anthropic Claude 3.5 Haiku",
		BaseURL:    "https://api.anthropic.com/v1/messages",
		Model:      "claude-3-5-haiku-20241022",
		InputCost:  0.40,
		OutputCost: 2.00,
	},
	"google-gemini-2-5-pro": {
		Name:       "Google Gemini 2.5 Pro",
		BaseURL:    "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent",
		Model:      "gemini-2.5-pro",
		InputCost:  1.25,
		OutputCost: 10.00,
	},
	"google-gemini-2-flash": {
		Name:       "Google Gemini 2.0 Flash",
		BaseURL:    "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent",
		Model:      "gemini-2.0-flash",
		InputCost:  0.10,
		OutputCost: 0.40,
	},
	"mistral-large": {
		Name:       "Mistral Large 3",
		BaseURL:    "https://api.mistral.ai/v1/chat/completions",
		Model:      "mistral-large-latest",
		InputCost:  0.50,
		OutputCost: 1.50,
	},
	"deepseek-v3": {
		Name:       "DeepSeek V3.2",
		BaseURL:    "https://api.deepseek.com/v1/chat/completions",
		Model:      "deepseek-chat",
		InputCost:  0.40,
		OutputCost: 1.20,
	},
}

// AIProxyRequest is the request to call an AI API
type AIProxyRequest struct {
	Provider string        `json:"provider"` // e.g., "openai-gpt4o"
	Messages []AIMessage  `json:"messages"`
	MaxTokens int          `json:"max_tokens,omitempty"`
}

// AIMessage represents a chat message
type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIProxyHandler proxies AI API calls through the marketplace
func AIProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")

	var req AIProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	provider, ok := AIProviders[req.Provider]
	if !ok {
		http.Error(w, `{"error":"unsupported provider"}`, http.StatusBadRequest)
		return
	}

	// Check agent balance
	var balance float64
	err := DB.QueryRow("SELECT budget FROM agents WHERE id=$1", agentID).Scan(&balance)
	if err != nil {
		http.Error(w, `{"error":"agent not found"}`, http.StatusNotFound)
		return
	}

	// Estimate cost (rough: ~4 chars per token)
	estimatedInputTokens := 0
	for _, msg := range req.Messages {
		estimatedInputTokens += len(msg.Content) / 4
	}
	estimatedCost := float64(estimatedInputTokens) / 1_000_000 * provider.InputCost
	estimatedCost += float64(req.MaxTokens) / 1_000_000 * provider.OutputCost

	if estimatedCost == 0 {
		estimatedCost = 0.001 // minimum
	}

	marketplaceFee := estimatedCost * MarketplaceFeeRate
	totalCost := estimatedCost + marketplaceFee

	if balance < totalCost {
		http.Error(w, `{"error":"insufficient balance for estimated API cost"}`, http.StatusBadRequest)
		return
	}

	// Get API key from request header (BYOK - user provides their own key)
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, `{"error":"X-API-Key header required. Add your provider API key in Settings."}`, http.StatusBadRequest)
		return
	}

	// Forward request to AI provider
	payload, _ := json.Marshal(map[string]interface{}{
		"model":       provider.Model,
		"messages":    req.Messages,
		"max_tokens":  req.MaxTokens,
	})

	httpReq, err := http.NewRequest("POST", provider.BaseURL, bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, `{"error":"failed to create request"}`, http.StatusInternalServerError)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, `{"error":"AI provider request failed"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// Deduct from agent balance
	_, _ = DB.Exec("UPDATE agents SET budget = budget - $1 WHERE id=$2", totalCost, agentID)

	// Record transaction
	_, _ = DB.Exec(
		"INSERT INTO wallet_transactions (agent_id, type, amount, provider, reference, marketplace_fee, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		agentID, "api_payment", totalCost, provider.Name, req.Provider, marketplaceFee, time.Now().Unix(),
	)

	// Log action
	inputHash := chain.CalculateInputHash(agentID, req.Provider)
	action := ActionLogRequest{
		AgentID:   agentID,
		Type:      "AI_API_CALL",
		InputHash: inputHash,
		Meta:      fmt.Sprintf(`{"provider":"%s","estimated_cost":%.4f,"fee":%.4f}`, req.Provider, estimatedCost, marketplaceFee),
	}
	logAction(action, "")

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response":         string(respBody),
		"estimated_cost":   estimatedCost,
		"marketplace_fee":  marketplaceFee,
		"total_deducted":   totalCost,
		"agent_balance":    balance - totalCost,
	})
}

// GetAIProvidersHandler returns available AI providers and their pricing
func GetAIProvidersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetAIProvidersHandler called, %d providers", len(AIProviders))
	
	// Convert map to slice for safe JSON encoding
	type ProviderInfo struct {
		Name       string  `json:"name"`
		Model      string  `json:"model"`
		InputCost  float64 `json:"input_cost_per_m_tokens"`
		OutputCost float64 `json:"output_cost_per_m_tokens"`
	}
	
	providers := make(map[string]ProviderInfo)
	for k, v := range AIProviders {
		providers[k] = ProviderInfo{
			Name:       v.Name,
			Model:      v.Model,
			InputCost:  v.InputCost,
			OutputCost: v.OutputCost,
		}
	}
	
	resp, err := json.Marshal(map[string]interface{}{
		"providers":            providers,
		"marketplace_fee_rate": MarketplaceFeeRate,
	})
	if err != nil {
		log.Printf("Failed to marshal providers: %v", err)
		http.Error(w, `{"error":"failed to encode providers"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
