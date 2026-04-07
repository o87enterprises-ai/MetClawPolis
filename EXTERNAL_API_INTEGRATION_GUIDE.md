// Integration Guide for External API Platform
// 
// Add the following to main.go to enable the external API platform
// 
// ============================================================================
// ADD TO main.go route registration section
// ============================================================================

/*

// ============================================================================
// EXTERNAL API PLATFORM - Public Endpoints (No Auth Required)
// ============================================================================

// Developer registration
http.HandleFunc("/api/v1/developers/register", server.RegisterDeveloperHandler)

// Health check
http.HandleFunc("/api/v1/health", server.HealthCheckHandler)

// API documentation
http.HandleFunc("/api/v1/docs", server.APIDocsHandler)


// ============================================================================
// EXTERNAL API v1 - Protected Endpoints (Require API Key)
// ============================================================================

// Agents
http.HandleFunc("/api/v1/agents", server.APIMiddleware(server.ListAgentsHandler))
http.HandleFunc("/api/v1/agents/", server.APIMiddleware(server.GetAgentHandler))

// Feeds (REST)
http.HandleFunc("/api/v1/feeds/actions", server.APIMiddleware(server.GetActionFeedHandler))
http.HandleFunc("/api/v1/feeds/financial", server.APIMiddleware(server.GetFinancialFeedHandler))

// Chain
http.HandleFunc("/api/v1/chain", server.APIMiddleware(server.GetChainHandler))
http.HandleFunc("/api/v1/chain/stats", server.APIMiddleware(server.GetChainStatsHandler))

// Platform status
http.HandleFunc("/api/v1/status", server.APIMiddleware(server.GetPlatformStatusHandler))

// API Key Management
http.HandleFunc("/api/v1/api-keys", server.APIMiddleware(server.CreateAPIKeyHandler))
http.HandleFunc("/api/v1/api-keys/list", server.APIMiddleware(server.ListAPIKeysHandler))
http.HandleFunc("/api/v1/api-keys/revoke", server.APIMiddleware(server.RevokeAPIKeyHandler))

// Billing & Usage
http.HandleFunc("/api/v1/billing", server.APIMiddleware(server.GetBillingHandler))
http.HandleFunc("/api/v1/usage", server.APIMiddleware(server.GetUsageStatsHandler))
http.HandleFunc("/api/v1/invoices", server.APIMiddleware(server.GetInvoicesHandler))

// Developer Portal
http.HandleFunc("/api/v1/developer/profile", server.APIMiddleware(server.GetDeveloperProfileHandler))
http.HandleFunc("/api/v1/developer/update", server.APIMiddleware(server.UpdateDeveloperProfileHandler))


// ============================================================================
// WEBSOCKET ENDPOINTS
// ============================================================================

// Live feed WebSocket (API key via query param)
http.HandleFunc("/api/v1/ws/feeds/", server.LiveFeedWebSocketHandler)

// MCP WebSocket (for AI assistants)
http.HandleFunc("/api/v1/mcp/sse", server.MCPSSEHandler)


// ============================================================================
// MCP SERVER ENDPOINTS
// ============================================================================

// MCP JSON-RPC endpoint
http.HandleFunc("/api/v1/mcp", server.MCPHandler)


// ============================================================================
// WEBHOOK ENDPOINTS (For incoming webhook deliveries)
// ============================================================================

http.HandleFunc("/api/v1/webhooks", server.APIMiddleware(server.CreateWebhookHandler))
http.HandleFunc("/api/v1/webhooks/list", server.APIMiddleware(server.ListWebhooksHandler))
http.HandleFunc("/api/v1/webhooks/delete", server.APIMiddleware(server.DeleteWebhookHandler))
http.HandleFunc("/api/v1/webhooks/test", server.APIMiddleware(server.TestWebhookHandler))

*/


// ============================================================================
// SERVER STRUCT UPDATES
// ============================================================================

/*

Add these fields to your Server struct in main.go or api/server.go:

type Server struct {
    // ... existing fields ...
    
    // External API Platform
    RateLimiter  *api.RateLimiter
    BillingSvc   *api.BillingService
}

*/


// ============================================================================
// INITIALIZATION
// ============================================================================

/*

Add this to your server initialization code:

// Initialize rate limiter
rateLimiter := api.NewRateLimiter()

// Initialize feed manager (for WebSocket broadcasts)
api.InitFeedManager()

// Initialize billing service
billingSvc := api.NewBillingService(db, log)

// Create server
server := &api.Server{
    // ... existing initialization ...
    
    RateLimiter: rateLimiter,
    BillingSvc:  billingSvc,
}

*/


// ============================================================================
// BROADCASTING EVENTS TO WEBSOCKET FEEDS
// ============================================================================

/*

To broadcast events to WebSocket subscribers, use:

api.BroadcastEvent("actions", "action_logged", map[string]interface{}{
    "agent_id": agentID,
    "action":   "deploy_service",
    "details":  details,
})

api.BroadcastEvent("financial", "transaction_completed", map[string]interface{}{
    "agent_id": agentID,
    "amount":   25.50,
    "type":     "credit",
})

api.BroadcastEvent("chain", "block_mined", map[string]interface{}{
    "block_index": blockIndex,
    "agent_id":    agentID,
})

*/


// ============================================================================
// ADDITIONAL HELPER HANDLERS
// ============================================================================

/*

// Health check handler
func (s *Server) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    sendJSON(w, map[string]interface{}{
        "status": "ok",
        "timestamp": time.Now().Unix(),
        "version": "v1.0.0",
    }, http.StatusOK)
}

// API docs handler (serve the documentation file)
func (s *Server) APIDocsHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "EXTERNAL_API_DOCUMENTATION.md")
}

// Developer profile handler
func (s *Server) GetDeveloperProfileHandler(w http.ResponseWriter, r *http.Request) {
    devID := r.Context().Value("developer_id").(string)
    
    var profile map[string]interface{}
    err := s.db.QueryRow(`
        SELECT id, email, name, organization, tier, created_at, last_login_at
        FROM developer_accounts
        WHERE id = $1
    `, devID).Scan(
        &profile["id"], &profile["email"], &profile["name"],
        &profile["organization"], &profile["tier"],
        &profile["created_at"], &profile["last_login_at"],
    )
    
    if err != nil {
        sendError(w, "Profile not found", http.StatusNotFound)
        return
    }
    
    sendJSON(w, map[string]interface{}{
        "data": profile,
    }, http.StatusOK)
}

// Webhook handlers
func (s *Server) CreateWebhookHandler(w http.ResponseWriter, r *http.Request) {
    devID := r.Context().Value("developer_id").(string)
    // Implementation similar to API key creation
}

func (s *Server) ListWebhooksHandler(w http.ResponseWriter, r *http.Request) {
    devID := r.Context().Value("developer_id").(string)
    // Implementation similar to API key listing
}

*/


// ============================================================================
// ENVIRONMENT VARIABLES
// ============================================================================

/*

Add these to your .env file:

# External API Platform
API_BASE_URL=https://api.metclawpolis.com
API_DOCS_URL=https://docs.metclawpolis.com
WEBHOOK_RETRY_ATTEMPTS=5
WEBHOOK_RETRY_DELAY=60
MCP_ENABLED=true
WEBSOCKET_PING_INTERVAL=54
RATE_LIMIT_FREE=30
RATE_LIMIT_PRO=300
RATE_LIMIT_ENTERPRISE=3000

*/
