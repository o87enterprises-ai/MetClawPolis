package api

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"io"
	"log"
	"net/http"
)

// AuthMiddleware verifies agent signatures on each request
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC in AuthMiddleware: %v", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()

		// Skip auth for public endpoints
		publicPaths := map[string]bool{
			"/api/chain":         true,
			"/health":            true,
			"/":                  true,
			"/api/agent/create":  true,
			"/api/agent":         true,
			"/api/providers":     true,
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

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error":"failed to read body"}`, http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Fetch public key from DB
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

		// Verify signature
		if !ed25519.Verify(pubKey, body, sig) {
			http.Error(w, `{"error":"invalid signature"}`, http.StatusUnauthorized)
			return
		}

		// Add agent ID to context for downstream handlers
		r.Header.Set("X-Verified-Agent-ID", agentID)
		next(w, r)
	}
}
