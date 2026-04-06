package api

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/agent_platform?sslmode=disable"
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

// InitOllamaConfigSchema creates the ollama_configs table if it doesn't exist
func InitOllamaConfigSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS ollama_configs (
		id VARCHAR(255) PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL,
		base_url VARCHAR(500) DEFAULT 'http://localhost:11434',
		model VARCHAR(255) DEFAULT 'llama3.2',
		temperature DOUBLE PRECISION DEFAULT 0.7,
		top_p DOUBLE PRECISION DEFAULT 0.9,
		top_k INTEGER DEFAULT 40,
		max_tokens INTEGER DEFAULT 4096,
		stop TEXT[],
		frequency_penalty DOUBLE PRECISION DEFAULT 0.0,
		presence_penalty DOUBLE PRECISION DEFAULT 0.0,
		seed INTEGER DEFAULT 0,
		num_thread INTEGER DEFAULT 0,
		num_gpu INTEGER DEFAULT 0,
		main_gpu INTEGER DEFAULT 0,
		use_mlock BOOLEAN DEFAULT false,
		use_mmap BOOLEAN DEFAULT true,
		typical_p DOUBLE PRECISION DEFAULT 0.0,
		repeat_penalty DOUBLE PRECISION DEFAULT 1.1,
		repeat_last_n INTEGER DEFAULT 64,
		mirostat INTEGER DEFAULT 0,
		mirostat_tau DOUBLE PRECISION DEFAULT 0.0,
		mirostat_eta DOUBLE PRECISION DEFAULT 0.0,
		tfs_z DOUBLE PRECISION DEFAULT 0.0,
		system_prompt TEXT DEFAULT '',
		stream BOOLEAN DEFAULT false,
		timeout INTEGER DEFAULT 60,
		created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW()),
		updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW()),
		UNIQUE(agent_id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_ollama_configs_agent_id ON ollama_configs(agent_id);
	`
	
	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create ollama_configs table: %w", err)
	}
	
	return nil
}
