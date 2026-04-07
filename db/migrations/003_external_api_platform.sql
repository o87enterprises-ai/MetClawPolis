-- External API Platform Migration
-- Enables third-party API access, developer accounts, API keys, usage tracking, and billing

-- ============================================================================
-- DEVELOPER ACCOUNTS
-- Human developers who register for API access
-- ============================================================================
CREATE TABLE IF NOT EXISTS developer_accounts (
    id TEXT PRIMARY KEY,                          -- UUID
    email TEXT UNIQUE NOT NULL,                   -- Developer email
    name TEXT NOT NULL,                           -- Display name
    organization TEXT,                            -- Company/org name (optional)
    password_hash TEXT NOT NULL,                  -- Bcrypt hashed password
    status TEXT DEFAULT 'active',                 -- active, suspended, deleted
    tier TEXT DEFAULT 'free',                     -- free, pro, enterprise
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    last_login_at BIGINT,
    metadata JSONB DEFAULT '{}'                   -- Additional developer info
);

CREATE INDEX IF NOT EXISTS idx_developer_accounts_email ON developer_accounts(email);
CREATE INDEX IF NOT EXISTS idx_developer_accounts_status ON developer_accounts(status);
CREATE INDEX IF NOT EXISTS idx_developer_accounts_tier ON developer_accounts(tier);

-- ============================================================================
-- API KEYS
-- Multiple keys per developer for different integrations
-- ============================================================================
CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,                          -- UUID
    developer_id TEXT NOT NULL REFERENCES developer_accounts(id) ON DELETE CASCADE,
    key_prefix TEXT NOT NULL,                     -- First 8 chars for identification (e.g., "mclw_live_")
    key_hash TEXT UNIQUE NOT NULL,                -- SHA-256 hash of the full key
    name TEXT NOT NULL,                           -- Key name (e.g., "Production", "Testing")
    permissions JSONB NOT NULL DEFAULT '[]',      -- Array of permission scopes
    rate_limit_override INTEGER,                  -- Custom rate limit (requests/min), NULL uses tier default
    quota_override BIGINT,                        -- Custom monthly quota, NULL uses tier default
    ip_whitelist JSONB DEFAULT '[]',              -- Array of allowed IPs/CIDRs (empty = all allowed)
    last_used_at BIGINT,                          -- Last successful request
    expires_at BIGINT,                            -- Key expiration (NULL = never)
    revoked_at BIGINT,                            -- When key was revoked (NULL = active)
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    metadata JSONB DEFAULT '{}'                   -- Usage notes, integration type, etc.
);

CREATE INDEX IF NOT EXISTS idx_api_keys_developer_id ON api_keys(developer_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON api_keys(key_prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_revoked_at ON api_keys(revoked_at);

-- ============================================================================
-- USAGE TRACKING
-- Per-request tracking for billing and analytics
-- ============================================================================
CREATE TABLE IF NOT EXISTS api_usage_logs (
    id BIGSERIAL PRIMARY KEY,
    api_key_id TEXT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    developer_id TEXT NOT NULL REFERENCES developer_accounts(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,                       -- API endpoint called
    method TEXT NOT NULL,                         -- HTTP method
    status_code INTEGER NOT NULL,                 -- Response status code
    request_size_bytes INTEGER,                   -- Request payload size
    response_size_bytes INTEGER,                  -- Response payload size
    response_time_ms INTEGER,                     -- Response time in milliseconds
    tokens_consumed INTEGER DEFAULT 0,            -- For token-based billing (LLM calls)
    ip_address TEXT,                              -- Client IP
    user_agent TEXT,                              -- Client user agent
    created_at BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_api_usage_logs_api_key_id ON api_usage_logs(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_usage_logs_developer_id ON api_usage_logs(developer_id);
CREATE INDEX IF NOT EXISTS idx_api_usage_logs_created_at ON api_usage_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_api_usage_logs_endpoint ON api_usage_logs(endpoint);

-- ============================================================================
-- USAGE QUOTAS
-- Monthly rolling quotas per developer
-- ============================================================================
CREATE TABLE IF NOT EXISTS api_quotas (
    id TEXT PRIMARY KEY,                          -- UUID
    developer_id TEXT UNIQUE NOT NULL REFERENCES developer_accounts(id) ON DELETE CASCADE,
    period_start BIGINT NOT NULL,                 -- Start of current billing period (epoch)
    period_end BIGINT NOT NULL,                   -- End of current billing period (epoch)
    requests_made BIGINT DEFAULT 0,               -- Total requests this period
    requests_limit BIGINT NOT NULL,               -- Max requests allowed
    tokens_consumed BIGINT DEFAULT 0,             -- Total tokens consumed (for LLM calls)
    tokens_limit BIGINT,                          -- Token limit (NULL = unlimited)
    compute_seconds_used BIGINT DEFAULT 0,        -- Total compute seconds (for agent runs)
    compute_seconds_limit BIGINT,                 -- Compute limit (NULL = unlimited)
    overage_charges NUMERIC(10,4) DEFAULT 0,      -- Charges for exceeding quota
    last_reset_at BIGINT NOT NULL,                -- When quota was last reset
    updated_at BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_api_quotas_developer_id ON api_quotas(developer_id);
CREATE INDEX IF NOT EXISTS idx_api_quotas_period ON api_quotas(period_start, period_end);

-- ============================================================================
-- WEBSOCKET SUBSCRIPTIONS
-- For live agent feed access
-- ============================================================================
CREATE TABLE IF NOT EXISTS ws_subscriptions (
    id TEXT PRIMARY KEY,                          -- UUID
    developer_id TEXT NOT NULL REFERENCES developer_accounts(id) ON DELETE CASCADE,
    api_key_id TEXT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    feed_type TEXT NOT NULL,                      -- agent_actions, financial, chain, all
    agent_filter JSONB DEFAULT '[]',              -- Array of agent IDs to filter (empty = all)
    status TEXT DEFAULT 'active',                 -- active, paused, revoked
    max_connections INTEGER DEFAULT 3,            -- Max concurrent WebSocket connections
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ws_subscriptions_developer_id ON ws_subscriptions(developer_id);
CREATE INDEX IF NOT EXISTS idx_ws_subscriptions_status ON ws_subscriptions(status);

-- ============================================================================
-- MCP SERVER CONFIGURATION
-- MCP server instances and their settings
-- ============================================================================
CREATE TABLE IF NOT EXISTS mcp_servers (
    id TEXT PRIMARY KEY,                          -- UUID
    developer_id TEXT NOT NULL REFERENCES developer_accounts(id) ON DELETE CASCADE,
    name TEXT NOT NULL,                           -- Server display name
    description TEXT,                             -- Server description
    endpoint_url TEXT UNIQUE NOT NULL,            -- Public endpoint URL
    auth_token_hash TEXT NOT NULL,                -- MCP auth token hash
    allowed_tools JSONB DEFAULT '[]',             -- Array of allowed MCP tools
    allowed_resources JSONB DEFAULT '[]',         -- Array of allowed MCP resources
    rate_limit INTEGER DEFAULT 100,               -- Requests per minute
    status TEXT DEFAULT 'active',                 -- active, paused, disabled
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_mcp_servers_developer_id ON mcp_servers(developer_id);
CREATE INDEX IF NOT EXISTS idx_mcp_servers_status ON mcp_servers(status);

-- ============================================================================
-- BILLING & INVOICES
-- For paid tiers and overage charges
-- ============================================================================
CREATE TABLE IF NOT EXISTS api_invoices (
    id TEXT PRIMARY KEY,                          -- UUID
    developer_id TEXT NOT NULL REFERENCES developer_accounts(id) ON DELETE CASCADE,
    amount NUMERIC(10,2) NOT NULL,                -- Total amount charged
    currency TEXT DEFAULT 'USD',
    status TEXT DEFAULT 'pending',                -- pending, paid, failed, refunded
    line_items JSONB NOT NULL,                    -- Array of line items
    stripe_invoice_id TEXT,                       -- Stripe invoice ID (if applicable)
    stripe_payment_intent_id TEXT,                -- Stripe payment intent
    paid_at BIGINT,                               -- When payment was received
    due_date BIGINT NOT NULL,                     -- Payment due date
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_api_invoices_developer_id ON api_invoices(developer_id);
CREATE INDEX IF NOT EXISTS idx_api_invoices_status ON api_invoices(status);

-- ============================================================================
-- API ANALYTICS AGGREGATES
-- Pre-computed analytics for dashboard
-- ============================================================================
CREATE TABLE IF NOT EXISTS api_analytics (
    id TEXT PRIMARY KEY,                          -- UUID
    developer_id TEXT NOT NULL REFERENCES developer_accounts(id) ON DELETE CASCADE,
    period TEXT NOT NULL,                         -- hour, day, week, month
    period_start BIGINT NOT NULL,                 -- Start of aggregation period
    period_end BIGINT NOT NULL,                   -- End of aggregation period
    total_requests BIGINT DEFAULT 0,
    successful_requests BIGINT DEFAULT 0,
    failed_requests BIGINT DEFAULT 0,
    avg_response_time_ms INTEGER DEFAULT 0,
    total_tokens_consumed BIGINT DEFAULT 0,
    total_bandwidth_bytes BIGINT DEFAULT 0,
    unique_endpoints INTEGER DEFAULT 0,           -- Number of unique endpoints called
    top_endpoints JSONB DEFAULT '[]',             -- Top 10 endpoints by usage
    error_breakdown JSONB DEFAULT '{}',           -- Error counts by status code
    created_at BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_api_analytics_developer_id ON api_analytics(developer_id);
CREATE INDEX IF NOT EXISTS idx_api_analytics_period ON api_analytics(period, period_start);

-- ============================================================================
-- WEBHOOK ENDPOINTS
-- For developers to receive real-time events
-- ============================================================================
CREATE TABLE IF NOT EXISTS webhooks (
    id TEXT PRIMARY KEY,                          -- UUID
    developer_id TEXT NOT NULL REFERENCES developer_accounts(id) ON DELETE CASCADE,
    url TEXT NOT NULL,                            -- Webhook endpoint URL
    secret TEXT NOT NULL,                         -- HMAC signing secret
    events JSONB NOT NULL DEFAULT '[]',           -- Array of event types to send
    status TEXT DEFAULT 'active',                 -- active, disabled, failed
    failure_count INTEGER DEFAULT 0,              -- Consecutive delivery failures
    last_delivery_at BIGINT,                      -- Last successful delivery
    last_delivery_status TEXT,                    -- Status of last delivery
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_webhooks_developer_id ON webhooks(developer_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_status ON webhooks(status);

-- ============================================================================
-- WEBHOOK DELIVERY LOGS
-- Track webhook delivery attempts
-- ============================================================================
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id BIGSERIAL PRIMARY KEY,
    webhook_id TEXT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,                     -- Type of event
    payload_hash TEXT NOT NULL,                   -- Hash of payload for verification
    response_status INTEGER,                      -- HTTP response status from endpoint
    response_time_ms INTEGER,                     -- Delivery attempt time
    attempt_number INTEGER DEFAULT 1,             -- Retry attempt number
    delivered BOOLEAN DEFAULT FALSE,              -- Whether delivery was successful
    error_message TEXT,                           -- Error if delivery failed
    created_at BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_created_at ON webhook_deliveries(created_at);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_delivered ON webhook_deliveries(delivered);

-- ============================================================================
-- SEED DATA: Tier configurations
-- ============================================================================
-- Note: Tier configs are managed in code, but we store the current config here
CREATE TABLE IF NOT EXISTS api_tier_configs (
    tier TEXT PRIMARY KEY,                        -- free, pro, enterprise
    config JSONB NOT NULL,                        -- Full tier configuration
    updated_at BIGINT NOT NULL
);

INSERT INTO api_tier_configs (tier, config, updated_at) VALUES
('free', '{
    "name": "Free",
    "price_monthly": 0,
    "rate_limit_per_minute": 30,
    "monthly_request_quota": 10000,
    "token_limit": null,
    "compute_seconds_limit": 3600,
    "max_api_keys": 3,
    "max_webhook_endpoints": 2,
    "max_ws_subscriptions": 2,
    "features": ["basic_api_access", "agent_feeds_read", "chain_read", "websocket_feeds"],
    "overage_rate_per_1k_requests": 0.50,
    "overage_rate_per_1k_tokens": 0.01
}'::jsonb, EXTRACT(EPOCH FROM NOW())::BIGINT),

('pro', '{
    "name": "Pro",
    "price_monthly": 49.00,
    "rate_limit_per_minute": 300,
    "monthly_request_quota": 500000,
    "token_limit": 10000000,
    "compute_seconds_limit": 86400,
    "max_api_keys": 20,
    "max_webhook_endpoints": 10,
    "max_ws_subscriptions": 10,
    "features": ["full_api_access", "agent_feeds_read", "agent_feeds_write", "chain_read", "chain_write", "websocket_feeds", "mcp_server", "webhooks", "analytics_dashboard", "priority_support"],
    "overage_rate_per_1k_requests": 0.25,
    "overage_rate_per_1k_tokens": 0.005
}'::jsonb, EXTRACT(EPOCH FROM NOW())::BIGINT),

('enterprise', '{
    "name": "Enterprise",
    "price_monthly": 299.00,
    "rate_limit_per_minute": 3000,
    "monthly_request_quota": null,
    "token_limit": null,
    "compute_seconds_limit": null,
    "max_api_keys": 100,
    "max_webhook_endpoints": 50,
    "max_ws_subscriptions": 50,
    "features": ["full_api_access", "agent_feeds_read", "agent_feeds_write", "chain_read", "chain_write", "websocket_feeds", "mcp_server", "webhooks", "analytics_dashboard", "dedicated_support", "custom_integrations", "sla_guarantee", "white_label"],
    "overage_rate_per_1k_requests": 0.10,
    "overage_rate_per_1k_tokens": 0.002
}'::jsonb, EXTRACT(EPOCH FROM NOW())::BIGINT)
ON CONFLICT (tier) DO NOTHING;

-- ============================================================================
-- COMMENTS
-- ============================================================================
COMMENT ON TABLE developer_accounts IS 'Developer accounts for external API access';
COMMENT ON TABLE api_keys IS 'API keys for authenticating external requests';
COMMENT ON TABLE api_usage_logs IS 'Per-request usage tracking for billing and analytics';
COMMENT ON TABLE api_quotas IS 'Monthly rolling quotas per developer';
COMMENT ON TABLE ws_subscriptions IS 'WebSocket subscriptions for live agent feeds';
COMMENT ON TABLE mcp_servers IS 'MCP server configurations for third-party integrations';
COMMENT ON TABLE api_invoices IS 'Billing invoices for paid tiers and overages';
COMMENT ON TABLE api_analytics IS 'Pre-computed analytics aggregates for dashboard';
COMMENT ON TABLE webhooks IS 'Webhook endpoints for real-time event delivery';
COMMENT ON TABLE webhook_deliveries IS 'Webhook delivery attempt logs';
COMMENT ON TABLE api_tier_configs IS 'API tier pricing and feature configurations';
