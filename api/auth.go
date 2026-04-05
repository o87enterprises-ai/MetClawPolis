package api

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ── In-memory auth store (replace with Redis/DB in production) ──

type AuthSession struct {
	Email       string    `json:"email"`
	Token       string    `json:"token"`
	TokenExpiry time.Time `json:"token_expiry"`
	Verified    bool      `json:"verified"`
	CreatedAt   time.Time `json:"created_at"`
	Remember    bool      `json:"remember"`
	Username    string    `json:"username"`
}

var (
	authSessions   = make(map[string]*AuthSession) // key: email
	authSessionsMu sync.RWMutex
)

// Generate a random hex token
func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Clean up expired sessions periodically
func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			authSessionsMu.Lock()
			for email, sess := range authSessions {
				if sess.TokenExpiry.Before(time.Now()) {
					delete(authSessions, email)
				}
			}
			authSessionsMu.Unlock()
		}
	}()
}

// ── Handlers ──

// RequestTokenHandler: Step 1 — User enters email, receives a timed verification token
func RequestTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Remember bool   `json:"remember"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		http.Error(w, `{"error":"invalid email address"}`, http.StatusBadRequest)
		return
	}

	// Generate 6-digit numeric token
	token := fmt.Sprintf("%06d", int(time.Now().UnixNano()%1000000))

	authSessionsMu.Lock()
	authSessions[req.Email] = &AuthSession{
		Email:       req.Email,
		Token:       token,
		TokenExpiry: time.Now().Add(10 * time.Minute),
		Verified:    false,
		CreatedAt:   time.Now(),
		Remember:    req.Remember,
	}
	authSessionsMu.Unlock()

	// In production: send email via SendGrid/AWS SES etc.
	// For now: log the token (and echo it back for dev/testing)
	log.Printf("AUTH TOKEN for %s: %s (expires in 10 min)", req.Email, token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Verification token sent to " + req.Email,
		"dev_token": token, // Remove in production
	})
}

// VerifyTokenHandler: Step 2 — User enters the 6-digit token
func VerifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Token    string `json:"token"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	authSessionsMu.Lock()
	sess, exists := authSessions[req.Email]
	authSessionsMu.Unlock()

	if !exists || sess.TokenExpiry.Before(time.Now()) {
		http.Error(w, `{"error":"token expired or not found. Please request a new one."}`, http.StatusUnauthorized)
		return
	}

	if sess.Token != req.Token {
		http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
		return
	}

	// Mark as verified
	authSessionsMu.Lock()
	sess.Verified = true
	if req.Username != "" {
		sess.Username = req.Username
	}
	authSessionsMu.Unlock()

	// Generate session token (long-lived if remember me)
	sessionDuration := 24 * time.Hour
	if sess.Remember {
		sessionDuration = 30 * 24 * time.Hour // 30 days
	}
	_ = generateToken(32) // session token placeholder
	expiresAt := time.Now().Add(sessionDuration)

	// Store session (simple: encode email into token with expiry)
	sessionData := fmt.Sprintf("%s|%d", sess.Email, expiresAt.Unix())
	sessionID := generateToken(48)

	// In production: store in Redis/DB. For now: in-memory with cleanup.
	storeSession(sessionID, sessionData, expiresAt, sess.Remember)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "ok",
		"session_token":  sessionID,
		"email":          sess.Email,
		"username":       sess.Username,
		"remember":       sess.Remember,
		"expires_at":     expiresAt.Unix(),
	})
}

// Session store (in-memory, production: Redis)
type SessionEntry struct {
	Data      string
	ExpiresAt time.Time
	Remember  bool
}

var (
	sessions   = make(map[string]*SessionEntry)
	sessionsMu sync.RWMutex
)

func storeSession(id, data string, expiresAt time.Time, remember bool) {
	sessionsMu.Lock()
	sessions[id] = &SessionEntry{Data: data, ExpiresAt: expiresAt, Remember: remember}
	sessionsMu.Unlock()
}

// ValidateSessionHandler: Validate an existing session token
func ValidateSessionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Session-Token")
	if token == "" {
		// Check cookie as fallback
		cookie, err := r.Cookie("mcp_session")
		if err != nil {
			http.Error(w, `{"error":"no session token"}`, http.StatusUnauthorized)
			return
		}
		token = cookie.Value
	}

	sessionsMu.Lock()
	entry, exists := sessions[token]
	if exists && entry.ExpiresAt.Before(time.Now()) {
		delete(sessions, token)
		exists = false
	}
	sessionsMu.Unlock()

	if !exists {
		http.Error(w, `{"error":"session expired or invalid"}`, http.StatusUnauthorized)
		return
	}

	parts := strings.SplitN(entry.Data, "|", 2)
	email := parts[0]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"email":   email,
		"remember": entry.Remember,
	})
}

// LogoutHandler: Invalidate session
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("X-Session-Token")
	if token != "" {
		sessionsMu.Lock()
		delete(sessions, token)
		sessionsMu.Unlock()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
	})
}

// MockEmailSender simulates sending a verification email
// In production, replace with SendGrid, AWS SES, etc.
func sendVerificationEmail(email, token string) error {
	// Production implementation:
	// return emailService.Send(email, "Your MetClawPolis Verification Code",
	//     fmt.Sprintf("Your verification code is: %s\nThis code expires in 10 minutes.", token))

	// For now: log it
	log.Printf("📧 EMAIL to %s: Your MetClawPolis verification code is %s", email, token)
	return nil
}

// GetEnv with fallback
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

// ── Agent ed25519 Auth Middleware (preserved from original) ──

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC in AuthMiddleware: %v", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()

		publicPaths := map[string]bool{
			"/api/chain":         true,
			"/health":            true,
			"/":                  true,
			"/api/agent/create":  true,
			"/api/agent":         true,
			"/api/providers":     true,
			"/api/auth/request":  true,
			"/api/auth/verify":   true,
			"/api/auth/validate": true,
		}
		if publicPaths[r.URL.Path] {
			next(w, r)
			return
		}

		agentID := r.Header.Get("X-Agent-ID")
		signatureHex := r.Header.Get("X-Signature")

		if agentID == "" || signatureHex == "" {
			http.Error(w, `{"error":"missing X-Agent-ID or X-Signature header"}`, http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error":"failed to read body"}`, http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		var pubKeyHex string
		err = DB.QueryRow("SELECT public_key FROM agents WHERE id=$1", agentID).Scan(&pubKeyHex)
		if err != nil {
			http.Error(w, `{"error":"agent not found"}`, http.StatusUnauthorized)
			return
		}

		pubKey, err := hex.DecodeString(pubKeyHex)
		if err != nil {
			http.Error(w, `{"error":"invalid public key"}`, http.StatusInternalServerError)
			return
		}

		sig, err := hex.DecodeString(signatureHex)
		if err != nil {
			http.Error(w, `{"error":"invalid signature format"}`, http.StatusBadRequest)
			return
		}

		if !ed25519.Verify(pubKey, body, sig) {
			http.Error(w, `{"error":"invalid signature"}`, http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-Verified-Agent-ID", agentID)
		next(w, r)
	}
}
