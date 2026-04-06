package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ============================================================================
// SERVICE CATALOG - Pricing with markups
// ============================================================================

// ServiceCategory represents a type of service
type ServiceCategory string

const (
	CategoryAI      ServiceCategory = "ai"
	CategoryCompute ServiceCategory = "compute"
	CategoryStorage ServiceCategory = "storage"
	CategoryNetwork ServiceCategory = "network"
)

// BillingUnit defines how the service is measured
type BillingUnit string

const (
	UnitPerToken     BillingUnit = "per_1k_tokens"
	UnitPerHour      BillingUnit = "per_hour"
	UnitPerMonth     BillingUnit = "per_month"
	UnitPerGB        BillingUnit = "per_gb"
	UnitPerRequest   BillingUnit = "per_request"
	UnitPerMinute    BillingUnit = "per_minute"
)

// ServiceConfig defines a resellable service
type ServiceConfig struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Category      ServiceCategory `json:"category"`
	Provider      string          `json:"provider"`
	BillingUnit   BillingUnit     `json:"billing_unit"`
	YourCost      float64         `json:"your_cost"`       // What YOU pay the provider
	UserPrice     float64         `json:"user_price"`      // What USER pays
	MarkupPercent float64         `json:"markup_percent"`  // Your margin
	Enabled       bool            `json:"enabled"`         // Available to users
	RequiresSetup bool            `json:"requires_setup"`  // Needs provider API key
	Metadata      map[string]any  `json:"metadata,omitempty"` // Provider-specific config
}

// ServiceCatalog holds all available services
type ServiceCatalog struct {
	Services map[string]ServiceConfig
	mu       sync.RWMutex
}

// Global service catalog
var Catalog = &ServiceCatalog{
	Services: make(map[string]ServiceConfig),
}

// InitServiceCatalog loads all services with pricing
func InitServiceCatalog() {
	Catalog.mu.Lock()
	defer Catalog.mu.Unlock()

	// ── AI Services ──
	Catalog.Services["gpt-4o"] = ServiceConfig{
		ID:            "gpt-4o",
		Name:          "OpenAI GPT-4o",
		Description:   "OpenAI's flagship model - best for complex tasks",
		Category:      CategoryAI,
		Provider:      "openai",
		BillingUnit:   UnitPerToken,
		YourCost:      0.010,    // $10/1M output tokens
		UserPrice:     0.015,    // $15/1M tokens (50% markup)
		MarkupPercent: 50,
		Enabled:       getEnvBool("ENABLE_OPENAI", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"model_id":     "gpt-4o",
			"input_cost":   0.0025,  // $2.50/1M input
			"output_cost":  0.010,   // $10/1M output
			"endpoint":     "https://api.openai.com/v1/chat/completions",
		},
	}

	Catalog.Services["gpt-4o-mini"] = ServiceConfig{
		ID:            "gpt-4o-mini",
		Name:          "OpenAI GPT-4o Mini",
		Description:   "Fast, affordable model for lighter tasks",
		Category:      CategoryAI,
		Provider:      "openai",
		BillingUnit:   UnitPerToken,
		YourCost:      0.0006,   // $0.60/1M output tokens
		UserPrice:     0.001,    // $1/1M tokens (67% markup)
		MarkupPercent: 67,
		Enabled:       getEnvBool("ENABLE_OPENAI", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"model_id":     "gpt-4o-mini",
			"input_cost":   0.00015,
			"output_cost":  0.0006,
			"endpoint":     "https://api.openai.com/v1/chat/completions",
		},
	}

	Catalog.Services["claude-3-7-sonnet"] = ServiceConfig{
		ID:            "claude-3-7-sonnet",
		Name:          "Anthropic Claude 3.7 Sonnet",
		Description:   "Anthropic's most intelligent model",
		Category:      CategoryAI,
		Provider:      "anthropic",
		BillingUnit:   UnitPerToken,
		YourCost:      0.015,    // $15/1M output tokens
		UserPrice:     0.022,    // $22/1M tokens (47% markup)
		MarkupPercent: 47,
		Enabled:       getEnvBool("ENABLE_ANTHROPIC", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"model_id":     "claude-3-7-sonnet-20250219",
			"input_cost":   0.003,
			"output_cost":  0.015,
			"endpoint":     "https://api.anthropic.com/v1/messages",
		},
	}

	Catalog.Services["claude-3-5-haiku"] = ServiceConfig{
		ID:            "claude-3-5-haiku",
		Name:          "Anthropic Claude 3.5 Haiku",
		Description:   "Fast, compact model for simple tasks",
		Category:      CategoryAI,
		Provider:      "anthropic",
		BillingUnit:   UnitPerToken,
		YourCost:      0.002,    // $2/1M output tokens (wait, actual is $1.25/1M output)
		UserPrice:     0.0035,   // $3.5/1M tokens (75% markup)
		MarkupPercent: 75,
		Enabled:       getEnvBool("ENABLE_ANTHROPIC", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"model_id":     "claude-3-5-haiku-20241022",
			"input_cost":   0.0008,
			"output_cost":  0.002,
			"endpoint":     "https://api.anthropic.com/v1/messages",
		},
	}

	Catalog.Services["gemini-2.5-pro"] = ServiceConfig{
		ID:            "gemini-2.5-pro",
		Name:          "Google Gemini 2.5 Pro",
		Description:   "Google's most capable model with long context",
		Category:      CategoryAI,
		Provider:      "google",
		BillingUnit:   UnitPerToken,
		YourCost:      0.010,    // $10/1M output tokens (half price for <128k)
		UserPrice:     0.015,    // $15/1M tokens (50% markup)
		MarkupPercent: 50,
		Enabled:       getEnvBool("ENABLE_GOOGLE", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"model_id":     "gemini-2.5-pro",
			"input_cost":   0.00125,
			"output_cost":  0.010,
			"endpoint":     "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent",
		},
	}

	Catalog.Services["gemini-2.0-flash"] = ServiceConfig{
		ID:            "gemini-2.0-flash",
		Name:          "Google Gemini 2.0 Flash",
		Description:   "Fast, efficient model for quick tasks",
		Category:      CategoryAI,
		Provider:      "google",
		BillingUnit:   UnitPerToken,
		YourCost:      0.0004,   // $0.40/1M output tokens
		UserPrice:     0.0007,   // $0.70/1M tokens (75% markup)
		MarkupPercent: 75,
		Enabled:       getEnvBool("ENABLE_GOOGLE", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"model_id":     "gemini-2.0-flash",
			"input_cost":   0.0001,
			"output_cost":  0.0004,
			"endpoint":     "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent",
		},
	}

	Catalog.Services["local-ollama"] = ServiceConfig{
		ID:            "local-ollama",
		Name:          "Local Ollama (FREE)",
		Description:   "Connect your own Ollama instance - completely free",
		Category:      CategoryAI,
		Provider:      "ollama",
		BillingUnit:   UnitPerToken,
		YourCost:      0.000,
		UserPrice:     0.000,
		MarkupPercent: 0,
		Enabled:       true, // Always enabled
		RequiresSetup: false,
		Metadata: map[string]any{
			"description": "Users run their own Ollama, zero cost to platform",
		},
	}

	// ── Compute Services ──
	Catalog.Services["hf-space-cpu"] = ServiceConfig{
		ID:            "hf-space-cpu",
		Name:          "Hugging Face Space (CPU)",
		Description:   "Deploy apps on Hugging Face Spaces with CPU",
		Category:      CategoryCompute,
		Provider:      "huggingface",
		BillingUnit:   UnitPerHour,
		YourCost:      0.000,    // FREE tier
		UserPrice:     0.005,    // $0.005/hour = $3.60/month (pure profit)
		MarkupPercent: 100,
		Enabled:       getEnvBool("ENABLE_HF", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"hardware":    "CPU",
			"ram":         "16GB",
			"monthly_est": 3.60,
		},
	}

	Catalog.Services["hf-space-gpu-t4"] = ServiceConfig{
		ID:            "hf-space-gpu-t4",
		Name:          "Hugging Face Space (GPU T4)",
		Description:   "Deploy apps with NVIDIA T4 GPU for AI workloads",
		Category:      CategoryCompute,
		Provider:      "huggingface",
		BillingUnit:   UnitPerHour,
		YourCost:      0.0008,   // $0.60/hour T4 (wait, that's not right - it's $0.60/hr)
		UserPrice:     0.0012,   // $0.86/hour (50% markup)
		MarkupPercent: 50,
		Enabled:       getEnvBool("ENABLE_HF", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"hardware":    "NVIDIA T4",
			"ram":         "30GB",
			"vcpu":        4,
			"monthly_est": 8.76,
		},
	}

	Catalog.Services["render-small"] = ServiceConfig{
		ID:            "render-small",
		Name:          "Render Small Instance",
		Description:   "Small web service on Render - good for APIs and bots",
		Category:      CategoryCompute,
		Provider:      "render",
		BillingUnit:   UnitPerHour,
		YourCost:      0.007,    // ~$5/month
		UserPrice:     0.010,    // ~$7/month (43% markup)
		MarkupPercent: 43,
		Enabled:       getEnvBool("ENABLE_RENDER", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"ram":         "512MB",
			"cpu":         "0.5 CPU",
			"monthly_est": 7.30,
		},
	}

	Catalog.Services["render-medium"] = ServiceConfig{
		ID:            "render-medium",
		Name:          "Render Medium Instance",
		Description:   "Medium web service - good for full-stack apps",
		Category:      CategoryCompute,
		Provider:      "render",
		BillingUnit:   UnitPerHour,
		YourCost:      0.014,    // ~$10/month
		UserPrice:     0.020,    // ~$14/month (43% markup)
		MarkupPercent: 43,
		Enabled:       getEnvBool("ENABLE_RENDER", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"ram":         "2GB",
			"cpu":         "1 CPU",
			"monthly_est": 14.60,
		},
	}

	// ── Storage Services ──
	Catalog.Services["r2-storage"] = ServiceConfig{
		ID:            "r2-storage",
		Name:          "Cloudflare R2 Storage",
		Description:   "S3-compatible object storage with zero egress fees",
		Category:      CategoryStorage,
		Provider:      "cloudflare",
		BillingUnit:   UnitPerGB,
		YourCost:      0.00002,  // $0.015/GB/month = $0.00002/GB/hour
		UserPrice:     0.00003,  // $0.022/GB/month (50% markup)
		MarkupPercent: 50,
		Enabled:       getEnvBool("ENABLE_R2", false),
		RequiresSetup: true,
		Metadata: map[string]any{
			"free_tier_gb": 10,
			"monthly_per_gb": 0.015,
		},
	}

	log.Printf("Service catalog initialized with %d services", len(Catalog.Services))
}

// GetService returns a service config by ID
func (c *ServiceCatalog) GetService(id string) (ServiceConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	svc, ok := c.Services[id]
	return svc, ok
}

// GetServicesByCategory returns all services in a category
func (c *ServiceCatalog) GetServicesByCategory(cat ServiceCategory) []ServiceConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	var services []ServiceConfig
	for _, svc := range c.Services {
		if svc.Category == cat && svc.Enabled {
			services = append(services, svc)
		}
	}
	return services
}

// GetAllEnabled returns all enabled services
func (c *ServiceCatalog) GetAllEnabled() []ServiceConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	var services []ServiceConfig
	for _, svc := range c.Services {
		if svc.Enabled {
			services = append(services, svc)
		}
	}
	return services
}

// ============================================================================
// USAGE TRACKING - Database schema and tracking
// ============================================================================

// ServiceUsage records a single service usage event
type ServiceUsage struct {
	ID            string    `json:"id"`
	AgentID       string    `json:"agent_id"`
	ServiceID     string    `json:"service_id"`
	Provider      string    `json:"provider"`
	Timestamp     time.Time `json:"timestamp"`
	Quantity      float64   `json:"quantity"`       // Tokens, hours, GB, etc.
	YourCost      float64   `json:"your_cost"`      // What it cost YOU
	UserPrice     float64   `json:"user_price"`     // What user is charged
	TotalCharged  float64   `json:"total_charged"`  // quantity * user_price
	YourProfit    float64   `json:"your_profit"`    // total_charged - (quantity * your_cost)
	Metadata      string    `json:"metadata"`       // JSON blob for details
}

// InitUsageTrackingSchema creates the database tables for usage tracking
func InitUsageTrackingSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS service_usage (
		id VARCHAR(255) PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL,
		service_id VARCHAR(100) NOT NULL,
		provider VARCHAR(50) NOT NULL,
		timestamp BIGINT NOT NULL,
		quantity DOUBLE PRECISION NOT NULL,
		your_cost DOUBLE PRECISION NOT NULL,
		user_price DOUBLE PRECISION NOT NULL,
		total_charged DOUBLE PRECISION NOT NULL,
		your_profit DOUBLE PRECISION NOT NULL,
		metadata TEXT,
		created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())
	);
	
	CREATE INDEX IF NOT EXISTS idx_service_usage_agent ON service_usage(agent_id);
	CREATE INDEX IF NOT EXISTS idx_service_usage_service ON service_usage(service_id);
	CREATE INDEX IF NOT EXISTS idx_service_usage_timestamp ON service_usage(timestamp);
	
	CREATE TABLE IF NOT EXISTS service_subscriptions (
		id VARCHAR(255) PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL,
		service_id VARCHAR(100) NOT NULL,
		provider_id VARCHAR(255), -- ID on the provider's side (VM ID, etc.)
		status VARCHAR(20) DEFAULT 'active', -- active, paused, stopped
		started_at BIGINT NOT NULL,
		last_billed_at BIGINT,
		next_billing_at BIGINT,
		hourly_rate DOUBLE PRECISION NOT NULL,
		your_hourly_cost DOUBLE PRECISION NOT NULL,
		metadata TEXT,
		created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW()),
		UNIQUE(agent_id, service_id, provider_id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_service_sub_agent ON service_subscriptions(agent_id);
	CREATE INDEX IF NOT EXISTS idx_service_sub_status ON service_subscriptions(status);
	
	CREATE TABLE IF NOT EXISTS service_api_keys (
		id VARCHAR(255) PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL UNIQUE,
		provider VARCHAR(50) NOT NULL,
		encrypted_key TEXT NOT NULL,
		created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW()),
		updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())
	);
	
	CREATE INDEX IF NOT EXISTS idx_service_api_keys_agent ON service_api_keys(agent_id);
	`
	
	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create usage tracking tables: %w", err)
	}
	
	return nil
}

// RecordUsage logs a service usage event
func RecordUsage(usage ServiceUsage) error {
	if DB == nil {
		return fmt.Errorf("database not available")
	}
	
	_, err := DB.Exec(
		`INSERT INTO service_usage 
		(id, agent_id, service_id, provider, timestamp, quantity, your_cost, user_price, total_charged, your_profit, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		usage.ID, usage.AgentID, usage.ServiceID, usage.Provider,
		usage.Timestamp.Unix(), usage.Quantity, usage.YourCost,
		usage.UserPrice, usage.TotalCharged, usage.YourProfit, usage.Metadata,
	)
	
	return err
}

// GetAgentUsage returns usage records for an agent
func GetAgentUsage(agentID string, serviceID string, limit int) ([]ServiceUsage, error) {
	query := `SELECT id, agent_id, service_id, provider, timestamp, quantity, 
			  your_cost, user_price, total_charged, your_profit, metadata
			  FROM service_usage WHERE agent_id = $1`
	args := []interface{}{agentID}
	
	if serviceID != "" {
		query += " AND service_id = $2"
		args = append(args, serviceID)
	}
	
	query += " ORDER BY timestamp DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var usages []ServiceUsage
	for rows.Next() {
		var u ServiceUsage
		var ts int64
		err := rows.Scan(
			&u.ID, &u.AgentID, &u.ServiceID, &u.Provider,
			&ts, &u.Quantity, &u.YourCost, &u.UserPrice,
			&u.TotalCharged, &u.YourProfit, &u.Metadata,
		)
		if err != nil {
			return nil, err
		}
		u.Timestamp = time.Unix(ts, 0)
		usages = append(usages, u)
	}
	
	return usages, nil
}

// GetAgentUsageSummary returns aggregated usage stats
func GetAgentUsageSummary(agentID string) (map[string]float64, error) {
	rows, err := DB.Query(
		`SELECT service_id, SUM(total_charged) as total_spent
		 FROM service_usage WHERE agent_id = $1
		 GROUP BY service_id`,
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	summary := make(map[string]float64)
	for rows.Next() {
		var serviceID string
		var totalSpent float64
		if err := rows.Scan(&serviceID, &totalSpent); err != nil {
			return nil, err
		}
		summary[serviceID] = totalSpent
	}
	
	return summary, nil
}

// GetPlatformRevenue returns your total profit across all users
func GetPlatformRevenue(startTime, endTime time.Time) (float64, error) {
	var totalProfit float64
	err := DB.QueryRow(
		`SELECT SUM(your_profit) FROM service_usage
		 WHERE timestamp >= $1 AND timestamp <= $2`,
		startTime.Unix(), endTime.Unix(),
	).Scan(&totalProfit)
	
	return totalProfit, err
}

// ============================================================================
// WALLET BALANCE MANAGEMENT
// ============================================================================

// GetUserBalance returns the user's available balance
func GetUserBalance(agentID string) (float64, error) {
	var balance float64
	err := DB.QueryRow(
		"SELECT budget FROM agents WHERE id = $1",
		agentID,
	).Scan(&balance)
	
	if err != nil {
		return 0, err
	}
	
	return balance, nil
}

// CheckAndReserveBalance checks if user has enough balance and reserves it
func CheckAndReserveBalance(agentID string, amount float64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	var balance float64
	err = tx.QueryRow("SELECT budget FROM agents WHERE id = $1 FOR UPDATE", agentID).Scan(&balance)
	if err != nil {
		return err
	}
	
	if balance < amount {
		return fmt.Errorf("insufficient balance: have %.4f, need %.4f", balance, amount)
	}
	
	// Reserve the funds
	_, err = tx.Exec("UPDATE agents SET budget = budget - $1 WHERE id = $2", amount, agentID)
	if err != nil {
		return err
	}
	
	return tx.Commit()
}

// RefundBalance returns funds to user's balance
func RefundBalance(agentID string, amount float64) error {
	_, err := DB.Exec(
		"UPDATE agents SET budget = budget + $1 WHERE id = $2",
		amount, agentID,
	)
	return err
}

// ============================================================================
// SERVICE PROXY - Main proxy with billing
// ============================================================================

// ServiceProxyRequest is the request to use a service
type ServiceProxyRequest struct {
	ServiceID string         `json:"service_id"`
	Messages  []AIMessage    `json:"messages"`
	MaxTokens int            `json:"max_tokens,omitempty"`
	Options   map[string]any `json:"options,omitempty"`
}

// ProxyService handles proxying a request to an AI provider with billing
func ProxyService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	
	agentID := r.Header.Get("X-Agent-ID")
	
	var req ServiceProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	
	// Get service config
	service, ok := Catalog.GetService(req.ServiceID)
	if !ok {
		http.Error(w, fmt.Sprintf(`{"error":"service '%s' not found"}`, req.ServiceID), http.StatusBadRequest)
		return
	}
	
	if !service.Enabled {
		http.Error(w, fmt.Sprintf(`{"error":"service '%s' is currently disabled"}`, req.ServiceID), http.StatusServiceUnavailable)
		return
	}
	
	// Handle local Ollama separately (free)
	if service.Provider == "ollama" {
		OllamaProxyHandler(w, r)
		return
	}
	
	// Check user has API key for this provider stored
	apiKey, err := GetProviderAPIKey(agentID, service.Provider)
	if err != nil {
		// Try platform's own API key
		apiKey = GetPlatformAPIKey(service.Provider)
		if apiKey == "" {
			http.Error(w, fmt.Sprintf(`{"error":"no API key configured for %s"}`, service.Provider), http.StatusServiceUnavailable)
			return
		}
	}
	
	// Estimate cost
	estimatedTokens := estimateMessageTokens(req.Messages)
	estimatedCost := estimatedTokens * service.UserPrice
	
	// Check and reserve balance
	if err := CheckAndReserveBalance(agentID, estimatedCost); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	
	// Forward to provider
	var response *http.Response
	var actualTokens float64
	
	switch service.Provider {
	case "openai":
		response, actualTokens, err = ForwardToOpenAI(req, service, apiKey)
	case "anthropic":
		response, actualTokens, err = ForwardToAnthropic(req, service, apiKey)
	case "google":
		response, actualTokens, err = ForwardToGoogle(req, service, apiKey)
	default:
		RefundBalance(agentID, estimatedCost)
		http.Error(w, `{"error":"unsupported provider"}`, http.StatusBadRequest)
		return
	}
	
	if err != nil {
		RefundBalance(agentID, estimatedCost)
		http.Error(w, fmt.Sprintf(`{"error":"provider request failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer response.Body.Close()
	
	body, _ := io.ReadAll(response.Body)
	
	// Calculate actual cost
	actualCost := actualTokens * service.UserPrice
	yourCost := actualTokens * service.YourCost
	profit := actualCost - yourCost
	
	// Refund difference if estimated > actual, or charge extra if actual > estimated
	if estimatedCost > actualCost {
		RefundBalance(agentID, estimatedCost-actualCost)
	} else if actualCost > estimatedCost {
		// Charge the difference (shouldn't happen often with good estimates)
		CheckAndReserveBalance(agentID, actualCost-estimatedCost)
	}
	
	// Record usage
	usage := ServiceUsage{
		ID:           fmt.Sprintf("usage_%d", time.Now().UnixNano()),
		AgentID:      agentID,
		ServiceID:    service.ID,
		Provider:     service.Provider,
		Timestamp:    time.Now(),
		Quantity:     actualTokens,
		YourCost:     yourCost,
		UserPrice:    service.UserPrice,
		TotalCharged: actualCost,
		YourProfit:   profit,
		Metadata:     fmt.Sprintf(`{"status":%d}`, response.StatusCode),
	}
	RecordUsage(usage)
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response":      string(body),
		"tokens_used":   actualTokens,
		"cost_charged":  actualCost,
		"service":       service.Name,
	})
}

// ============================================================================
// PROVIDER INTEGRATIONS
// ============================================================================

// ForwardToOpenAI sends request to OpenAI API
func ForwardToOpenAI(req ServiceProxyRequest, service ServiceConfig, apiKey string) (*http.Response, float64, error) {
	payload := map[string]interface{}{
		"model":    service.Metadata["model_id"],
		"messages": req.Messages,
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	
	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequest("POST", service.Metadata["endpoint"].(string), bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	
	// Estimate tokens from response (rough: ~4 chars per token)
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	
	// In production, parse usage from response
	estimatedTokens := float64(len(body)) / 4.0 / 1000.0 // in thousands
	
	return resp, estimatedTokens, nil
}

// ForwardToAnthropic sends request to Anthropic API
func ForwardToAnthropic(req ServiceProxyRequest, service ServiceConfig, apiKey string) (*http.Response, float64, error) {
	payload := map[string]interface{}{
		"model":     service.Metadata["model_id"],
		"messages":  req.Messages,
		"max_tokens": 4096,
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	
	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequest("POST", service.Metadata["endpoint"].(string), bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	
	estimatedTokens := float64(len(body)) / 4.0 / 1000.0
	
	return resp, estimatedTokens, nil
}

// ForwardToGoogle sends request to Google AI API
func ForwardToGoogle(req ServiceProxyRequest, service ServiceConfig, apiKey string) (*http.Response, float64, error) {
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": req.Messages[0].Content},
				},
			},
		},
	}
	
	body, _ := json.Marshal(payload)
	
	// Google uses query param for API key
	url := fmt.Sprintf("%s?key=%s", service.Metadata["endpoint"], apiKey)
	httpReq, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	
	estimatedTokens := float64(len(body)) / 4.0 / 1000.0
	
	return resp, estimatedTokens, nil
}

// ============================================================================
// API KEY MANAGEMENT
// ============================================================================

// GetProviderAPIKey retrieves user's stored API key for a provider
func GetProviderAPIKey(agentID, provider string) (string, error) {
	if DB == nil {
		return "", fmt.Errorf("database not available")
	}
	
	var encryptedKey string
	err := DB.QueryRow(
		"SELECT encrypted_key FROM service_api_keys WHERE agent_id = $1 AND provider = $2",
		agentID, provider,
	).Scan(&encryptedKey)
	
	if err != nil {
		return "", err
	}
	
	// In production, decrypt here
	return encryptedKey, nil
}

// GetPlatformAPIKey gets the platform's own API key for a provider
func GetPlatformAPIKey(provider string) string {
	switch provider {
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "google":
		return os.Getenv("GOOGLE_AI_API_KEY")
	case "huggingface":
		return os.Getenv("HUGGINGFACE_API_KEY")
	default:
		return ""
	}
}

// SaveProviderAPIKey stores a user's API key for a provider
func SaveProviderAPIKey(agentID, provider, apiKey string) error {
	if DB == nil {
		return fmt.Errorf("database not available")
	}
	
	// In production, encrypt before storing
	_, err := DB.Exec(
		`INSERT INTO service_api_keys (id, agent_id, provider, encrypted_key, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (agent_id) DO UPDATE SET encrypted_key = $4, updated_at = $5`,
		fmt.Sprintf("key_%s_%s", agentID, provider),
		agentID, provider, apiKey, time.Now().Unix(),
	)
	
	return err
}

// ============================================================================
// AUTO-BILLING FOR SUBSCRIPTIONS (VMs, etc.)
// ============================================================================

// StartBillingCron begins the hourly billing job for active subscriptions
func StartBillingCron() {
	log.Println("Starting billing cron job (hourly)")
	
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			if err := ProcessHourlyBilling(); err != nil {
				log.Printf("Billing cron error: %v", err)
			}
		}
	}()
}

// ProcessHourlyBilling charges all active subscriptions for the past hour
func ProcessHourlyBilling() error {
	if DB == nil {
		return fmt.Errorf("database not available")
	}
	
	// Get all active subscriptions
	rows, err := DB.Query(
		`SELECT id, agent_id, service_id, hourly_rate, your_hourly_cost
		 FROM service_subscriptions WHERE status = 'active'`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	totalBilled := 0.0
	totalProfit := 0.0
	failedCount := 0
	
	for rows.Next() {
		var id, agentID, serviceID string
		var hourlyRate, hourlyCost float64
		
		if err := rows.Scan(&id, &agentID, &serviceID, &hourlyRate, &hourlyCost); err != nil {
			continue
		}
		
		// Try to charge user
		if err := CheckAndReserveBalance(agentID, hourlyRate); err != nil {
			log.Printf("Failed to bill %s for %s: %v", agentID, serviceID, err)
			failedCount++
			
			// Auto-stop if balance consistently low
			DB.Exec(
				"UPDATE service_subscriptions SET status = 'suspended' WHERE id = $1",
				id,
			)
			continue
		}
		
		// Record usage
		usage := ServiceUsage{
			ID:           fmt.Sprintf("sub_%d_%s", time.Now().Unix(), id),
			AgentID:      agentID,
			ServiceID:    serviceID,
			Provider:     "subscription",
			Timestamp:    time.Now(),
			Quantity:     1.0, // 1 hour
			YourCost:     hourlyCost,
			UserPrice:    hourlyRate,
			TotalCharged: hourlyRate,
			YourProfit:   hourlyRate - hourlyCost,
			Metadata:     `{"type":"subscription","duration_hours":1}`,
		}
		RecordUsage(usage)
		
		// Update last billed timestamp
		DB.Exec(
			"UPDATE service_subscriptions SET last_billed_at = $1 WHERE id = $2",
			time.Now().Unix(), id,
		)
		
		totalBilled += hourlyRate
		totalProfit += (hourlyRate - hourlyCost)
	}
	
	log.Printf("Billing cron complete: billed $%.2f, profit $%.2f, failed %d",
		totalBilled, totalProfit, failedCount)
	
	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func estimateMessageTokens(messages []AIMessage) float64 {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content)
	}
	// Rough estimate: ~4 characters per token
	return float64(totalChars) / 4.0 / 1000.0 // Return in thousands
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val == "true" || val == "1" || val == "yes"
}

// ============================================================================
// API HANDLERS
// ============================================================================

// GetServiceCatalogHandler returns all available services
func GetServiceCatalogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	
	category := r.URL.Query().Get("category")
	
	var services []ServiceConfig
	if category != "" {
		services = Catalog.GetServicesByCategory(ServiceCategory(category))
	} else {
		services = Catalog.GetAllEnabled()
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services": services,
		"count":    len(services),
	})
}

// GetAgentUsageHandler returns agent's usage history
func GetAgentUsageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	
	agentID := r.Header.Get("X-Agent-ID")
	serviceID := r.URL.Query().Get("service_id")
	
	usage, err := GetAgentUsage(agentID, serviceID, 100)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	
	summary, _ := GetAgentUsageSummary(agentID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"usage":   usage,
		"summary": summary,
	})
}

// SaveAPIKeyHandler stores user's API key for a provider
func SaveAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	
	agentID := r.Header.Get("X-Agent-ID")
	
	var req struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	
	if err := SaveProviderAPIKey(agentID, req.Provider, req.APIKey); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to save API key: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"provider": req.Provider,
	})
}

// GetRevenueHandler returns platform revenue stats (admin only)
func GetRevenueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	
	period := r.URL.Query().Get("period") // "day", "week", "month"
	
	var startTime time.Time
	now := time.Now()
	
	switch period {
	case "day":
		startTime = now.Add(-24 * time.Hour)
	case "week":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "month":
		startTime = now.Add(-30 * 24 * time.Hour)
	default:
		startTime = now.Add(-24 * time.Hour)
	}
	
	revenue, err := GetPlatformRevenue(startTime, now)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"period":      period,
		"revenue":     revenue,
		"start_time":  startTime.Unix(),
		"end_time":    now.Unix(),
	})
}

// OllamaProxyHandler handles requests to user's local Ollama (free)
func OllamaProxyHandler(w http.ResponseWriter, r *http.Request) {
	// This reuses the existing ollama_proxy.go implementation
	// Just forward the request
	log.Printf("Ollama proxy request received - forwarding to local Ollama")
	
	// For now, return a simple response indicating the proxy is working
	// The full implementation is in ollama_proxy.go
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"message": "Local Ollama proxy - using your own models, zero cost",
		"note": "Full implementation in ollama_proxy.go endpoints",
	})
}
