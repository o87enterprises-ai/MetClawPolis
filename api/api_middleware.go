//go:build external_api
// +build external_api

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*TokenBucket
	db       *sql.DB
	cleanup  *time.Ticker
}

// TokenBucket represents a rate limiter bucket
type TokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// APIMiddleware wraps the external API authentication and rate limiting
func (s *Server) APIMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendError(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		apiKeyStr := strings.TrimPrefix(authHeader, "Bearer ")
		if apiKeyStr == authHeader {
			sendError(w, "Invalid Authorization header format. Use: Bearer <api_key>", http.StatusUnauthorized)
			return
		}

		// Validate API key
		apiKey, err := s.ValidateAPIKey(apiKeyStr)
		if err != nil {
			s.log.Printf("Invalid API key: %v", err)
			sendError(w, "Invalid or revoked API key", http.StatusUnauthorized)
			return
		}

		// Check IP whitelist
		if len(apiKey.IPWhitelist) > 0 {
			clientIP := getClientIP(r)
			if !isIPAllowed(clientIP, apiKey.IPWhitelist) {
				sendError(w, "IP address not whitelisted", http.StatusForbidden)
				return
			}
		}

		// Check permissions
		requiredPermission := getRequiredPermission(r.URL.Path, r.Method)
		if requiredPermission != "" && !hasPermission(apiKey.Permissions, requiredPermission) {
			sendError(w, fmt.Sprintf("Missing required permission: %s", requiredPermission), http.StatusForbidden)
			return
		}

		// Check rate limit
		if !s.RateLimiter.Allow(apiKey.ID) {
			sendRateLimitError(w)
			return
		}

		// Check quota
		quotaExceeded, err := s.CheckQuota(apiKey.DeveloperID)
		if err != nil {
			s.log.Printf("Error checking quota: %v", err)
		}
		if quotaExceeded {
			sendError(w, "Monthly API quota exceeded. Upgrade your plan or wait for next billing cycle", http.StatusTooManyRequests)
			return
		}

		// Add developer info to context
		ctx := context.WithValue(r.Context(), "developer_id", apiKey.DeveloperID)
		ctx = context.WithValue(ctx, "api_key_id", apiKey.ID)
		ctx = context.WithValue(ctx, "api_key", apiKey)

		// Log usage
		start := time.Now()
		wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r.WithContext(ctx))

		// Record usage
		duration := time.Since(start).Milliseconds()
		go s.RecordUsage(apiKey, r, wrapper.statusCode, duration)
	}
}

// RateLimiter methods

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*TokenBucket),
	}
	rl.cleanup = time.NewTicker(5 * time.Minute)
	go rl.runCleanup()
	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		// Default: 30 requests per minute for free tier
		bucket = &TokenBucket{
			tokens:     30,
			maxTokens:  30,
			refillRate: 0.5, // 30 per minute = 0.5 per second
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens = min(bucket.maxTokens, bucket.tokens+elapsed*bucket.refillRate)
	bucket.lastRefill = now

	if bucket.tokens < 1 {
		return false
	}

	bucket.tokens--
	return true
}

func (rl *RateLimiter) SetLimit(key string, requestsPerMinute int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.buckets[key] = &TokenBucket{
		tokens:     float64(requestsPerMinute),
		maxTokens:  float64(requestsPerMinute),
		refillRate: float64(requestsPerMinute) / 60.0,
		lastRefill: time.Now(),
	}
}

func (rl *RateLimiter) Remove(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.buckets, key)
}

func (rl *RateLimiter) runCleanup() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		for key, bucket := range rl.buckets {
			if time.Since(bucket.lastRefill) > 10*time.Minute {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Quota management

func (s *Server) CheckQuota(developerID string) (bool, error) {
	var quota APIQuota
	err := s.db.QueryRow(`
		SELECT requests_made, requests_limit, tokens_consumed, 
			COALESCE(tokens_limit, 0), compute_seconds_used, 
			COALESCE(compute_seconds_limit, 0)
		FROM api_quotas
		WHERE developer_id = $1
	`, developerID).Scan(
		&quota.RequestsMade, &quota.RequestsLimit,
		&quota.TokensConsumed, &quota.TokensLimit,
		&quota.ComputeSecondsUsed, &quota.ComputeSecondsLimit,
	)

	if err == sql.ErrNoRows {
		// No quota record yet, create one
		s.InitializeQuota(developerID)
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Check if quota is exceeded
	if quota.RequestsLimit > 0 && quota.RequestsMade >= quota.RequestsLimit {
		return true, nil
	}
	if quota.TokensLimit > 0 && quota.TokensConsumed >= quota.TokensLimit {
		return true, nil
	}
	if quota.ComputeSecondsLimit > 0 && quota.ComputeSecondsUsed >= quota.ComputeSecondsLimit {
		return true, nil
	}

	return false, nil
}

func (s *Server) InitializeQuota(developerID string) error {
	// Get developer's tier
	var tier string
	err := s.db.QueryRow("SELECT tier FROM developer_accounts WHERE id = $1", developerID).Scan(&tier)
	if err != nil {
		return err
	}

	// Get tier config
	var configJSON string
	err = s.db.QueryRow("SELECT config FROM api_tier_configs WHERE tier = $1", tier).Scan(&configJSON)
	if err != nil {
		return err
	}

	var tierConfig map[string]interface{}
	json.Unmarshal([]byte(configJSON), &tierConfig)

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0) // Next month

	requestsLimit := int64(tierConfig["monthly_request_quota"].(float64))
	tokensLimit := int64(0) // 0 = unlimited
	if v, ok := tierConfig["token_limit"]; ok && v != nil {
		tokensLimit = int64(v.(float64))
	}
	computeLimit := int64(0)
	if v, ok := tierConfig["compute_seconds_limit"]; ok && v != nil {
		computeLimit = int64(v.(float64))
	}

	_, err = s.db.Exec(`
		INSERT INTO api_quotas (id, developer_id, period_start, period_end, 
			requests_limit, tokens_limit, compute_seconds_limit, last_reset_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, generateUUID(), developerID, now.Unix(), periodEnd.Unix(),
		requestsLimit, tokensLimit, computeLimit, now.Unix(), now.Unix())

	return err
}

func (s *Server) RecordUsage(apiKey *APIKey, r *http.Request, statusCode int, responseTimeMs int64) {
	// Update quota
	_, err := s.db.Exec(`
		UPDATE api_quotas 
		SET requests_made = requests_made + 1,
			updated_at = $1
		WHERE developer_id = $2
	`, time.Now().Unix(), apiKey.DeveloperID)
	if err != nil {
		s.log.Printf("Error updating quota: %v", err)
	}

	// Log usage
	_, err = s.db.Exec(`
		INSERT INTO api_usage_logs (api_key_id, developer_id, endpoint, method, 
			status_code, request_size_bytes, response_time_ms, ip_address, 
			user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, apiKey.ID, apiKey.DeveloperID, r.URL.Path, r.Method,
		statusCode, r.ContentLength, responseTimeMs,
		getClientIP(r), r.UserAgent(), time.Now().Unix())
	if err != nil {
		s.log.Printf("Error logging API usage: %v", err)
	}
}

// Helper types and functions

type APIQuota struct {
	RequestsMade        int64
	RequestsLimit       int64
	TokensConsumed      int64
	TokensLimit         int64
	ComputeSecondsUsed  int64
	ComputeSecondsLimit int64
	OverageCharges      float64
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func getRequiredPermission(path string, method string) string {
	// Map endpoints to required permissions
	if strings.HasPrefix(path, "/api/v1/agents") {
		if method == http.MethodGet {
			return "read:agents"
		}
		return "write:agents"
	}
	if strings.HasPrefix(path, "/api/v1/feeds") {
		if method == http.MethodGet {
			return "read:feeds"
		}
		return "write:feeds"
	}
	if strings.HasPrefix(path, "/api/v1/chain") {
		if method == http.MethodGet {
			return "read:chain"
		}
		return "write:chain"
	}
	if strings.HasPrefix(path, "/api/v1/actions") {
		return "write:actions"
	}
	if strings.HasPrefix(path, "/api/v1/mcp") {
		return "use:mcp"
	}
	return "read:basic"
}

func hasPermission(permissions []string, required string) bool {
	// Check for wildcard permission
	for _, p := range permissions {
		if p == "*" || p == required {
			return true
		}
	}
	return false
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func isIPAllowed(ip string, whitelist []string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	for _, entry := range whitelist {
		if strings.Contains(entry, "/") {
			// CIDR notation
			_, ipNet, err := net.ParseCIDR(entry)
			if err == nil && ipNet.Contains(clientIP) {
				return true
			}
		} else {
			// Exact IP
			if net.ParseIP(entry).Equal(clientIP) {
				return true
			}
		}
	}
	return false
}

func sendRateLimitError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "60")
	w.Header().Set("X-RateLimit-Limit", "30")
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "Rate limit exceeded",
		"message": "Too many requests. Please slow down or upgrade your plan.",
		"retry_after_seconds": 60,
	})
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Developer account management

func (s *Server) RegisterDeveloperHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email        string `json:"email"`
		Name         string `json:"name"`
		Organization string `json:"organization,omitempty"`
		Password     string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Name == "" || req.Password == "" {
		sendError(w, "Email, name, and password are required", http.StatusBadRequest)
		return
	}

	// Hash password
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		s.log.Printf("Error hashing password: %v", err)
		sendError(w, "Internal error", http.StatusInternalServerError)
		return
	}

	devID := generateUUID()
	now := time.Now().Unix()

	_, err = s.db.Exec(`
		INSERT INTO developer_accounts (id, email, name, organization, password_hash, tier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'free', $6, $7)
	`, devID, req.Email, req.Name, req.Organization, passwordHash, now, now)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			sendError(w, "Email already registered", http.StatusConflict)
			return
		}
		s.log.Printf("Error creating developer account: %v", err)
		sendError(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	// Initialize quota
	s.InitializeQuota(devID)

	s.log.Printf("New developer account registered: %s (%s)", devID, req.Email)
	sendJSON(w, map[string]string{
		"success":   "Account created successfully",
		"developer_id": devID,
		"tier":      "free",
	}, http.StatusCreated)
}

func hashPassword(password string) (string, error) {
	// In production, use bcrypt. For now, simple SHA-256
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password))), nil
}
