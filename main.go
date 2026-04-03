package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/metclawpolis/agent-platform/api"
)

type spaHandler struct{}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "ui/index.html")
}

// staticHandler serves files with correct MIME types
type staticHandler struct {
	dir http.FileSystem
}

func (h staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filePath := "ui" + r.URL.Path
	f, err := h.dir.Open(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil || stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	// Set correct MIME type before serving
	ext := filepath.Ext(filePath)
	if ct := mime.TypeByExtension(ext); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
}

func main() {
	// Initialize database
	if err := api.InitDB(); err != nil {
		log.Printf("WARNING: Database not available - %v", err)
		log.Println("Running in mock mode (no persistence)")
	}
	defer api.CloseDB()

	// Initialize PoW blockchain
	api.InitBlockchain()
	log.Println("PoW Blockchain initialized with difficulty 2")

	// Setup routes with panic recovery
	mux := http.NewServeMux()
	handler := recoverMiddleware(mux)

	// Public endpoints
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/chain", api.AuthMiddleware(api.GetChainHandler))
	mux.HandleFunc("/api/agent", api.AuthMiddleware(api.GetAgentHandler))
	mux.HandleFunc("/api/agent/balance", api.AuthMiddleware(api.GetAgentBalanceHandler))
	mux.HandleFunc("/api/pages", api.AuthMiddleware(api.GetAgentPagesHandler))
	mux.HandleFunc("/api/escrow", api.AuthMiddleware(api.GetEscrowHandler))
	mux.HandleFunc("/api/providers", api.AuthMiddleware(api.GetAIProvidersHandler))

	// Authenticated endpoints (require X-Agent-ID + X-Signature)
	mux.HandleFunc("/api/agent/create", api.AuthMiddleware(api.CreateAgentHandler))
	mux.HandleFunc("/api/agent/skills", api.AuthMiddleware(api.UpdateAgentSkillsHandler))
	mux.HandleFunc("/api/agent/deposit", api.AuthMiddleware(api.DepositHandler))
	mux.HandleFunc("/api/action", api.AuthMiddleware(api.LogActionHandler))
	mux.HandleFunc("/api/page/create", api.AuthMiddleware(api.CreatePageHandler))
	mux.HandleFunc("/api/hire", api.AuthMiddleware(api.HireAgentHandler))
	mux.HandleFunc("/api/escrow/complete", api.AuthMiddleware(api.CompleteEscrowHandler))
	mux.HandleFunc("/api/ai/proxy", api.AuthMiddleware(api.AIProxyHandler))

	// Serve static assets (CSS, JS, images) with correct MIME types
	mux.Handle("/assets/", staticHandler{dir: http.Dir(".")})
	mux.Handle("/logo.png", staticHandler{dir: http.Dir(".")})

	// SPA fallback: serve index.html for all other non-API paths
	mux.Handle("/", spaHandler{})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("🚀 Agent Platform API server starting on http://localhost%s", addr)
	log.Printf("📋 View mock UI at http://localhost%s/", addr)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
	}()

	log.Fatal(http.ListenAndServe(addr, handler))
}

// recoverMiddleware recovers from panics in any handler
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"service": "agent-platform",
	})
}
