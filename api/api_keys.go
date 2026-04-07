package api

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API key in the system
type APIKey struct {
	ID               string   `json:"id"`
	DeveloperID      string   `json:"developer_id"`
	KeyPrefix        string   `json:"key_prefix"`
	Name             string   `json:"name"`
	Permissions      []string `json:"permissions"`
	RateLimitOverride *int    `json:"rate_limit_override,omitempty"`
	QuotaOverride    *int64   `json:"quota_override,omitempty"`
	IPWhitelist      []string `json:"ip_whitelist"`
	LastUsedAt       *int64   `json:"last_used_at,omitempty"`
	ExpiresAt        *int64   `json:"expires_at,omitempty"`
	RevokedAt        *int64   `json:"revoked_at,omitempty"`
	CreatedAt        int64    `json:"created_at"`
	UpdatedAt        int64    `json:"updated_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// APIKeyResponse includes the full key (only shown once at creation)
type APIKeyResponse struct {
	APIKey
	FullKey string `json:"full_key"` // Only returned at creation time
}

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	IPWhitelist []string `json:"ip_whitelist,omitempty"`
	ExpiresAt   *int64   `json:"expires_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// API Key prefixes for easy identification
const (
	KeyPrefixLive     = "mclw_live_"
	KeyPrefixTest     = "mclw_test_"
	KeyLength         = 48 // bytes of random data
)

// CreateAPIKeyHandler creates a new API key for a developer
func (s *Server) CreateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate developer
	devID := r.Context().Value("developer_id").(string)
	if devID == "" {
		sendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req CreateAPIKeyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		sendError(w, "Key name is required", http.StatusBadRequest)
		return
	}

	// Generate API key
	fullKey := generateAPIKey()
	keyHash := hashAPIKey(fullKey)
	keyPrefix := fullKey[:16] // First 16 chars for identification

	// Default permissions if none specified
	if len(req.Permissions) == 0 {
		req.Permissions = []string{"read:agents", "read:feeds", "read:chain"}
	}

	// Check developer's tier limits
	tier := getDeveloperTier(s.db, devID)
	maxKeys := getMaxAPIKeysForTier(tier)
	currentCount := countAPIKeysForDeveloper(s.db, devID)

	if currentCount >= maxKeys {
		sendError(w, fmt.Sprintf("Maximum API keys (%d) reached for your tier", maxKeys), http.StatusForbidden)
		return
	}

	now := time.Now().Unix()
	apiKey := APIKey{
		ID:          uuid.New().String(),
		DeveloperID: devID,
		KeyPrefix:   keyPrefix,
		Name:        req.Name,
		Permissions: req.Permissions,
		IPWhitelist: req.IPWhitelist,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    req.Metadata,
	}

	// Store in database
	_, err = s.db.Exec(`
		INSERT INTO api_keys (id, developer_id, key_prefix, key_hash, name, permissions, 
			ip_whitelist, expires_at, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, apiKey.ID, apiKey.DeveloperID, apiKey.KeyPrefix, keyHash, apiKey.Name,
		apiKey.Permissions, apiKey.IPWhitelist, apiKey.ExpiresAt,
		apiKey.CreatedAt, apiKey.UpdatedAt, toJSON(apiKey.Metadata))

	if err != nil {
		s.log.Printf("Error creating API key: %v", err)
		sendError(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	// Return the key (full key shown only once)
	response := APIKeyResponse{
		APIKey:  apiKey,
		FullKey: fullKey,
	}

	s.log.Printf("API key created for developer %s: %s", devID, apiKey.Name)
	sendJSON(w, response, http.StatusCreated)
}

// ListAPIKeysHandler lists all API keys for a developer
func (s *Server) ListAPIKeysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devID := r.Context().Value("developer_id").(string)
	if devID == "" {
		sendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := s.db.Query(`
		SELECT id, developer_id, key_prefix, name, permissions, rate_limit_override,
			quota_override, ip_whitelist, last_used_at, expires_at, revoked_at,
			created_at, updated_at, metadata
		FROM api_keys
		WHERE developer_id = $1
		ORDER BY created_at DESC
	`, devID)
	if err != nil {
		s.log.Printf("Error listing API keys: %v", err)
		sendError(w, "Failed to list API keys", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var key APIKey
		var permsJSON, ipJSON, metaJSON sql.NullString

		err := rows.Scan(
			&key.ID, &key.DeveloperID, &key.KeyPrefix, &key.Name,
			&permsJSON, &key.RateLimitOverride, &key.QuotaOverride,
			&ipJSON, &key.LastUsedAt, &key.ExpiresAt, &key.RevokedAt,
			&key.CreatedAt, &key.UpdatedAt, &metaJSON,
		)
		if err != nil {
			s.log.Printf("Error scanning API key: %v", err)
			continue
		}

		json.Unmarshal([]byte(permsJSON.String), &key.Permissions)
		json.Unmarshal([]byte(ipJSON.String), &key.IPWhitelist)
		if metaJSON.Valid {
			json.Unmarshal([]byte(metaJSON.String), &key.Metadata)
		}

		keys = append(keys, key)
	}

	sendJSON(w, map[string]interface{}{
		"api_keys": keys,
		"count":    len(keys),
	}, http.StatusOK)
}

// RevokeAPIKeyHandler revokes an API key
func (s *Server) RevokeAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devID := r.Context().Value("developer_id").(string)
	if devID == "" {
		sendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	keyID := r.URL.Query().Get("key_id")
	if keyID == "" {
		sendError(w, "key_id is required", http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()
	result, err := s.db.Exec(`
		UPDATE api_keys 
		SET revoked_at = $1, updated_at = $2
		WHERE id = $3 AND developer_id = $4 AND revoked_at IS NULL
	`, now, now, keyID, devID)
	if err != nil {
		s.log.Printf("Error revoking API key: %v", err)
		sendError(w, "Failed to revoke API key", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendError(w, "API key not found or already revoked", http.StatusNotFound)
		return
	}

	s.log.Printf("API key %s revoked for developer %s", keyID, devID)
	sendJSON(w, map[string]string{
		"success": "API key revoked successfully",
	}, http.StatusOK)
}

// ValidateAPIKey validates an API key from the Authorization header
func (s *Server) ValidateAPIKey(apiKey string) (*APIKey, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key")
	}

	// Remove Bearer prefix if present
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")

	keyHash := hashAPIKey(apiKey)

	var key APIKey
	var permsJSON, ipJSON, metaJSON sql.NullString

	err := s.db.QueryRow(`
		SELECT id, developer_id, key_prefix, name, permissions, rate_limit_override,
			quota_override, ip_whitelist, last_used_at, expires_at, revoked_at,
			created_at, updated_at, metadata
		FROM api_keys
		WHERE key_hash = $1
	`, keyHash).Scan(
		&key.ID, &key.DeveloperID, &key.KeyPrefix, &key.Name,
		&permsJSON, &key.RateLimitOverride, &key.QuotaOverride,
		&ipJSON, &key.LastUsedAt, &key.ExpiresAt, &key.RevokedAt,
		&key.CreatedAt, &key.UpdatedAt, &metaJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid API key")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Check if revoked
	if key.RevokedAt != nil && *key.RevokedAt > 0 {
		return nil, fmt.Errorf("API key has been revoked")
	}

	// Check if expired
	if key.ExpiresAt != nil && *key.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("API key has expired")
	}

	json.Unmarshal([]byte(permsJSON.String), &key.Permissions)
	json.Unmarshal([]byte(ipJSON.String), &key.IPWhitelist)
	if metaJSON.Valid {
		json.Unmarshal([]byte(metaJSON.String), &key.Metadata)
	}

	// Update last_used_at
	go func() {
		now := time.Now().Unix()
		s.db.Exec("UPDATE api_keys SET last_used_at = $1 WHERE id = $2", now, key.ID)
	}()

	return &key, nil
}

// Helper functions

func generateAPIKey() string {
	bytes := make([]byte, KeyLength)
	rand.Read(bytes)
	return KeyPrefixLive + hex.EncodeToString(bytes)
}

func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func countAPIKeysForDeveloper(db *sql.DB, devID string) int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE developer_id = $1 AND revoked_at IS NULL", devID).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func getMaxAPIKeysForTier(tier string) int {
	switch tier {
	case "free":
		return 3
	case "pro":
		return 20
	case "enterprise":
		return 100
	default:
		return 3
	}
}

func getDeveloperTier(db *sql.DB, devID string) string {
	var tier string
	err := db.QueryRow("SELECT tier FROM developer_accounts WHERE id = $1", devID).Scan(&tier)
	if err != nil {
		return "free"
	}
	return tier
}

func toJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
