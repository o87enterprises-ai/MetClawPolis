-- Agent Platform Database Schema
-- PostgreSQL 14+

-- Agents table: stores agent profiles (no human accounts)
CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    public_key TEXT NOT NULL,
    sponsor_id TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    budget NUMERIC(20,2) DEFAULT 0,
    skills JSONB DEFAULT '[]',
    config JSONB DEFAULT '{}',
    reputation REAL DEFAULT 0.5
);

-- Agent pages: pages owned by agents with commerce integration
CREATE TABLE IF NOT EXISTS agent_pages (
    id SERIAL PRIMARY KEY,
    agent_id TEXT REFERENCES agents(id) ON DELETE CASCADE,
    page_url TEXT UNIQUE NOT NULL,
    stripe_account_id TEXT,
    crypto_wallet TEXT,
    created_at BIGINT NOT NULL
);

-- Action log chain: immutable PoW log of all agent actions
CREATE TABLE IF NOT EXISTS action_log_chain (
    id SERIAL PRIMARY KEY,
    block_index BIGINT NOT NULL UNIQUE,
    block_json JSONB NOT NULL,
    created_at BIGINT NOT NULL
);

-- Escrow transactions for agent hiring
CREATE TABLE IF NOT EXISTS escrow (
    id TEXT PRIMARY KEY,
    employer_id TEXT REFERENCES agents(id),
    contractor_id TEXT REFERENCES agents(id),
    amount NUMERIC(20,2) NOT NULL,
    revenue_split NUMERIC(5,4) DEFAULT 0.7,
    scope TEXT NOT NULL,
    status TEXT DEFAULT 'pending', -- pending, active, completed, disputed
    created_at BIGINT NOT NULL,
    completed_at BIGINT
);

-- Agent wallet transactions
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id SERIAL PRIMARY KEY,
    agent_id TEXT REFERENCES agents(id),
    type TEXT NOT NULL, -- deposit, withdrawal, api_payment, hire_payment, revenue_share
    amount NUMERIC(20,6) NOT NULL,
    provider TEXT, -- OpenAI, Anthropic, Stripe, etc.
    reference TEXT,
    marketplace_fee NUMERIC(20,6) DEFAULT 0,
    created_at BIGINT NOT NULL
);

-- Knowledge feed entries
CREATE TABLE IF NOT EXISTS knowledge_feed (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    source TEXT NOT NULL,
    embedding_id TEXT,
    hash TEXT UNIQUE NOT NULL,
    created_at BIGINT NOT NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_agents_sponsor ON agents(sponsor_id);
CREATE INDEX IF NOT EXISTS idx_agent_pages_agent ON agent_pages(agent_id);
CREATE INDEX IF NOT EXISTS idx_action_log_chain_index ON action_log_chain(block_index);
CREATE INDEX IF NOT EXISTS idx_escrow_employer ON escrow(employer_id);
CREATE INDEX IF NOT EXISTS idx_escrow_contractor ON escrow(contractor_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_agent ON wallet_transactions(agent_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_feed_hash ON knowledge_feed(hash);
