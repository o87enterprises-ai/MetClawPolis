-- Beta Program Tables
-- For MetClawPolis beta testing program, feedback collection, and promo codes

-- Beta signups: stores beta tester registrations
CREATE TABLE IF NOT EXISTS beta_signups (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    interests TEXT DEFAULT '',
    signed_up_at BIGINT NOT NULL,
    status TEXT DEFAULT 'active', -- active, completed, churned
    referred_by TEXT DEFAULT ''
);

-- Beta feedback: stores feedback from beta testers
CREATE TABLE IF NOT EXISTS beta_feedback (
    id TEXT PRIMARY KEY,
    beta_signup_id TEXT REFERENCES beta_signups(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comments TEXT DEFAULT '',
    feature_requests TEXT DEFAULT '',
    bugs_reported TEXT DEFAULT '',
    submitted_at BIGINT NOT NULL
);

-- Promo codes: stores discount codes for beta testers
CREATE TABLE IF NOT EXISTS promo_codes (
    id TEXT PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    beta_signup_id TEXT REFERENCES beta_signups(id) ON DELETE CASCADE,
    discount_tier TEXT NOT NULL DEFAULT '3percent', -- 3percent, 1percent
    created_at BIGINT NOT NULL,
    used_count INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_beta_signups_email ON beta_signups(email);
CREATE INDEX IF NOT EXISTS idx_beta_feedback_signup ON beta_feedback(beta_signup_id);
CREATE INDEX IF NOT EXISTS idx_promo_codes_code ON promo_codes(code);
CREATE INDEX IF NOT EXISTS idx_promo_codes_signup ON promo_codes(beta_signup_id);
