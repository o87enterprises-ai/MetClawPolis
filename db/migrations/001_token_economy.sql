-- MCLW Token Economy Tables
-- For MetClawPolis native token (MCLW) on Solana
-- Supports: agent holdings, mining rewards, transfers, Bitiverse graduation

-- Token holdings: tracks each agent's MCLW balance
CREATE TABLE IF NOT EXISTS token_holdings (
    id SERIAL PRIMARY KEY,
    agent_id TEXT REFERENCES agents(id) ON DELETE CASCADE,
    mclw_balance NUMERIC(20,9) DEFAULT 0,
    earned_mining NUMERIC(20,9) DEFAULT 0,
    earned_graduation NUMERIC(20,9) DEFAULT 0,
    earned_commerce NUMERIC(20,9) DEFAULT 0,
    spent_transfers NUMERIC(20,9) DEFAULT 0,
    last_mined_at BIGINT DEFAULT 0,
    updated_at BIGINT NOT NULL,
    UNIQUE(agent_id)
);

-- Token transactions: full audit trail of MCLW movements
CREATE TABLE IF NOT EXISTS token_transactions (
    id TEXT PRIMARY KEY,
    agent_id TEXT REFERENCES agents(id),
    type TEXT NOT NULL,
    amount NUMERIC(20,9) NOT NULL,
    counterparty TEXT,
    reference TEXT,
    block_index BIGINT,
    created_at BIGINT NOT NULL
);

-- Mining rewards: tracks PoW mining reward distribution
CREATE TABLE IF NOT EXISTS mining_rewards (
    id SERIAL PRIMARY KEY,
    agent_id TEXT REFERENCES agents(id),
    block_index BIGINT NOT NULL,
    reward_amount NUMERIC(20,9) NOT NULL,
    difficulty INT NOT NULL,
    action_type TEXT NOT NULL,
    claimed BOOLEAN DEFAULT false,
    created_at BIGINT NOT NULL
);

-- Graduation records: Bitiverse → MCLW bridge
CREATE TABLE IF NOT EXISTS graduation_records (
    id SERIAL PRIMARY KEY,
    agent_id TEXT REFERENCES agents(id) UNIQUE,
    bic_graduated NUMERIC(20,2) NOT NULL,
    mclw_received NUMERIC(20,9) NOT NULL,
    exchange_rate NUMERIC(20,9) NOT NULL,
    moral_score NUMERIC(5,4),
    training_score NUMERIC(5,4),
    completed_lessons INT DEFAULT 0,
    graduated_at BIGINT NOT NULL,
    tx_signature TEXT
);

-- Token economy config: adjustable parameters
CREATE TABLE IF NOT EXISTS token_economy_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- Insert default config values
INSERT INTO token_economy_config (key, value, updated_at) VALUES
    ('mclw_mint_address', 'PLACEHOLDER', EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('mclw_total_supply', '1000000000', EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('mclw_mining_reward_base', '0.001', EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('mclw_mining_reward_multiplier', '1.0', EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('mclw_bic_exchange_rate', '0.01', EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('mclw_graduation_bonus', '10.0', EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('mclw_min_graduation_amount', '100', EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('mclw_mining_enabled', 'true', EXTRACT(EPOCH FROM NOW())::BIGINT)
ON CONFLICT (key) DO NOTHING;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_token_holdings_agent ON token_holdings(agent_id);
CREATE INDEX IF NOT EXISTS idx_token_transactions_agent ON token_transactions(agent_id);
CREATE INDEX IF NOT EXISTS idx_token_transactions_type ON token_transactions(type);
CREATE INDEX IF NOT EXISTS idx_token_transactions_created ON token_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_mining_rewards_agent ON mining_rewards(agent_id);
CREATE INDEX IF NOT EXISTS idx_mining_rewards_block ON mining_rewards(block_index);
CREATE INDEX IF NOT EXISTS idx_mining_rewards_claimed ON mining_rewards(claimed);
CREATE INDEX IF NOT EXISTS idx_graduation_records_agent ON graduation_records(agent_id);
