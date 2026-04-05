package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"
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
	// Load .env file
	_ = godotenv.Load()

	// Initialize database
	if err := api.InitDB(); err != nil {
		log.Printf("WARNING: Database not available - %v", err)
		log.Println("Running in mock mode (no persistence)")
	}
	defer api.CloseDB()

	// Initialize Redis
	if err := api.InitRedis(); err != nil {
		log.Printf("WARNING: Redis not available - %v", err)
	}

	// Initialize Stripe
	api.InitStripe()

	// Initialize PoW blockchain
	api.InitBlockchain()
	log.Println("PoW Blockchain initialized with difficulty 2")

	// Initialize Agent Runtime
	api.InitAgentRuntime()

	// Start price feed
	api.StartPriceFeed()

	// Initialize Bitiverse (starts global world simulation)
	api.InitBitiverse()

	// Start WebSocket manager
	go api.WS.Run()

	// Setup routes with panic recovery
	mux := http.NewServeMux()
	handler := recoverMiddleware(mux)

	// ── Auth endpoints (email + token sign-in) ──
	mux.HandleFunc("/api/auth/request", api.RequestTokenHandler)
	mux.HandleFunc("/api/auth/verify", api.VerifyTokenHandler)
	mux.HandleFunc("/api/auth/validate", api.ValidateSessionHandler)
	mux.HandleFunc("/api/auth/logout", api.LogoutHandler)

	// ── Public endpoints ──
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/chain", api.AuthMiddleware(api.GetChainHandler))
	mux.HandleFunc("/api/agent", api.AuthMiddleware(api.GetAgentHandler))
	mux.HandleFunc("/api/agent/balance", api.AuthMiddleware(api.GetAgentBalanceHandler))
	mux.HandleFunc("/api/agent/pages", api.AuthMiddleware(api.GetAgentPagesHandler))
	mux.HandleFunc("/api/agent/wallet", api.AuthMiddleware(api.GetWalletHandler))
	mux.HandleFunc("/api/agent/net-worth", api.AuthMiddleware(api.GetAgentNetWorthHandler))
	mux.HandleFunc("/api/agent/transactions", api.AuthMiddleware(api.GetTransactionsHandler))
	mux.HandleFunc("/api/escrow", api.AuthMiddleware(api.GetEscrowHandler))
	mux.HandleFunc("/api/providers", api.AuthMiddleware(api.GetAIProvidersHandler))
	mux.HandleFunc("/api/prices", api.AuthMiddleware(api.GetPricesHandler))
	mux.HandleFunc("/api/tunnel/status", api.AuthMiddleware(api.GetTunnelStatusHandler))
	mux.HandleFunc("/api/pages/analytics", api.AuthMiddleware(api.GetPageAnalyticsHandler))
	mux.HandleFunc("/api/notifications", api.AuthMiddleware(api.GetNotificationsHandler))

	// ── Authenticated endpoints ──
	mux.HandleFunc("/api/agent/create", api.AuthMiddleware(api.CreateAgentHandler))
	mux.HandleFunc("/api/agent/skills", api.AuthMiddleware(api.UpdateAgentSkillsHandler))
	mux.HandleFunc("/api/agent/deposit", api.AuthMiddleware(api.DepositHandler))
	mux.HandleFunc("/api/agent/deploy", api.AuthMiddleware(api.DeployAgentHandler))
	mux.HandleFunc("/api/agent/stop", api.AuthMiddleware(api.StopAgentHandler))
	mux.HandleFunc("/api/agent/runtime", api.AuthMiddleware(api.GetAgentRuntimeStatusHandler))
	mux.HandleFunc("/api/agent/wallet/generate", api.AuthMiddleware(api.GenerateWalletHandler))
	mux.HandleFunc("/api/action", api.AuthMiddleware(api.LogActionHandler))
	mux.HandleFunc("/api/page/create", api.AuthMiddleware(api.CreatePageHandler))
	mux.HandleFunc("/api/page/content", api.AuthMiddleware(api.CreatePageContentHandler))
	mux.HandleFunc("/api/hire", api.AuthMiddleware(api.HireAgentHandler))
	mux.HandleFunc("/api/escrow/complete", api.AuthMiddleware(api.CompleteEscrowHandler))
	mux.HandleFunc("/api/ai/proxy", api.AuthMiddleware(api.AIProxyHandler))
	mux.HandleFunc("/api/sync/services", api.AuthMiddleware(api.SyncServicesHandler))
	mux.HandleFunc("/api/sync/connect", api.AuthMiddleware(api.ConnectServiceHandler))
	mux.HandleFunc("/api/tunnel/start", api.AuthMiddleware(api.StartTunnelHandler))
	mux.HandleFunc("/api/tunnel/stop", api.AuthMiddleware(api.StopTunnelHandler))
	mux.HandleFunc("/api/trade", api.AuthMiddleware(api.TradeTokenHandler))

	// ── Payment endpoints (no agent auth, uses Stripe signatures) ──
	mux.HandleFunc("/api/payments/create", api.CreatePaymentIntentHandler)
	mux.HandleFunc("/api/payments/checkout", api.CreateCheckoutSessionHandler)
	mux.HandleFunc("/api/payments/withdraw", api.AuthMiddleware(api.CreatePayoutHandler))
	mux.HandleFunc("/api/stripe/webhook", api.StripeWebhookHandler)

	// ── Bitiverse endpoints ──
	mux.HandleFunc("/api/bitiverse/enable", api.EnableBitiverseHandler)
	mux.HandleFunc("/api/bitiverse/world", api.BitiverseWorldViewHandler)
	mux.HandleFunc("/api/bitiverse/status", api.BitiverseStatusHandler)
	mux.HandleFunc("/api/bitiverse/turn", api.BitiverseTurnHandler)
	mux.HandleFunc("/api/bitiverse/stats", api.BitiverseStatsHandler)
	mux.HandleFunc("/api/bitiverse/task", api.BitiverseAssignTaskHandler)
	mux.HandleFunc("/api/bitiverse/economy", api.BitiverseEconomyHandler)

	// ── Financial Feed endpoints ──
	mux.HandleFunc("/api/financials/agent", api.AuthMiddleware(api.AgentFinancialSummaryHandler))
	mux.HandleFunc("/api/financials/platform", api.AuthMiddleware(api.PlatformFinancialSummaryHandler))
	mux.HandleFunc("/api/financials/live-tx", api.AuthMiddleware(api.LiveTransactionsHandler))
	mux.HandleFunc("/api/financials/sparkline", api.AuthMiddleware(api.SparklineDataHandler))

	// ── Cloud Provider endpoints ──
	mux.HandleFunc("/api/cloud/providers", api.AuthMiddleware(api.GetCloudProvidersHandler))
	mux.HandleFunc("/api/cloud/purchase", api.AuthMiddleware(api.CreateCloudPurchaseHandler))

	// ── Beta Program endpoints (public signup, authenticated feedback) ──
	mux.HandleFunc("/api/beta/signup", api.BetaSignupHandler)
	mux.HandleFunc("/api/beta/feedback", api.BetaFeedbackHandler)
	mux.HandleFunc("/api/promo/generate", api.GeneratePromoCodeHandler)
	mux.HandleFunc("/api/promo/code", api.GetPromoCodeHandler)
	mux.HandleFunc("/api/promo/validate", api.ValidatePromoCodeHandler)
	mux.HandleFunc("/api/beta/email-feedback", api.SendFeedbackRequestHandler)

	// ── Hosted pages ──
	mux.HandleFunc("/pages/", servePage)

	// ── WebSocket endpoints ──
	mux.HandleFunc("/ws", api.HandleWebSocket)
	mux.HandleFunc("/ws/terminal", api.HandleTerminalWebSocket)

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

	// Find an available port
	port = findAvailablePort(port)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("MetClawPolis Agent Platform API server starting on http://localhost%s", addr)
	log.Printf("View dashboard at http://localhost%s/", addr)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
	}()

	log.Fatal(http.ListenAndServe(addr, handler))
}

// servePage serves hosted agent pages
func servePage(w http.ResponseWriter, r *http.Request) {
	// Parse /pages/:agentId/:slug from URL path
	path := r.URL.Path
	// Remove /pages/ prefix
	rest := path[len("/pages/"):]
	parts := splitPath(rest)
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	agentID := parts[0]
	slug := parts[1]

	// Redirect to handler with query params
	r.URL.RawQuery = "agent_id=" + agentID + "&slug=" + slug
	api.ServePageHandler(w, r)
}

func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
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

	// Check service health
	services := map[string]string{
		"database": "unavailable",
		"redis":    "unavailable",
		"stripe":   "unavailable",
	}
	if api.DB != nil {
		if err := api.DB.Ping(); err == nil {
			services["database"] = "ok"
		}
	}
	if api.RDB != nil {
		if err := api.RDB.Ping(r.Context()).Err(); err == nil {
			services["redis"] = "ok"
		}
	}
	if stripeKey := os.Getenv("STRIPE_SECRET_KEY"); stripeKey != "" && stripeKey != "sk_live_..." {
		services["stripe"] = "configured"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"service":  "metclawpolis",
		"services": services,
	})
}

// isPortAvailable checks if a port is available for binding
func isPortAvailable(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// findAvailablePort tries to find an available port starting from the default
func findAvailablePort(defaultPort string) string {
	// If default port is available, use it
	if isPortAvailable(defaultPort) {
		return defaultPort
	}

	// Try alternative ports
	alternativePorts := []string{"8081", "8082", "8083", "8084", "8085", "3000", "3001", "3002"}

	for _, port := range alternativePorts {
		if isPortAvailable(port) {
			log.Printf("Port %s is in use, trying alternative port %s...", defaultPort, port)
			return port
		}
	}

	// If all alternative ports are taken, let the server fail with the default port
	log.Printf("All alternative ports are in use, attempting with default port %s", defaultPort)
	return defaultPort
}
