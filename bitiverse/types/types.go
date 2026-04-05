package types

import "time"

// BitiverseState represents the complete state of a Bitiverse instance
type BitiverseState struct {
	AgentID       string          `json:"agent_id"`
	WorldState    *WorldState     `json:"world_state"`
	AgentVitals   *AgentVitals    `json:"agent_vitals"`
	Inventory     *Inventory      `json:"inventory"`
	Training      *TrainingState  `json:"training"`
	SocialGraph   *SocialGraph    `json:"social_graph"`
	CreatedAt     int64           `json:"created_at"`
	LastUpdatedAt int64           `json:"last_updated_at"`
}

// WorldState represents the 2D grid world
type WorldState struct {
	Width      int            `json:"width"`
	Height     int            `json:"height"`
	AgentX     int            `json:"agent_x"`
	AgentY     int            `json:"agent_y"`
	NPCs       []NPC          `json:"npcs"`
	Items      []WorldItem    `json:"items"`
	Buildings  []Building     `json:"buildings"`
	CurrentDay int            `json:"current_day"`
	CurrentHour int           `json:"current_hour"`
}

// NPC represents a non-player character in Bitiverse
type NPC struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"` // police, bank, hospital, teacher, shopkeeper
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Char     string  `json:"char"` // display character
	Metadata string  `json:"metadata"` // JSON metadata
}

// WorldItem represents an item in the world
type WorldItem struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // coin, book, tool, food
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Char  string `json:"char"`
	Value int    `json:"value"`
}

// Building represents a location in Bitiverse
type Building struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // home, school, bank, hospital, shop, gym
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Enterable   bool   `json:"enterable"`
	Description string `json:"description"`
}

// AgentVitals tracks agent health and status
type AgentVitals struct {
	Health       int     `json:"health"`        // 0-100
	Happiness    int     `json:"happiness"`     // 0-100
	Stress       int     `json:"stress"`        // 0-100
	Energy       int     `json:"energy"`        // 0-100
	Reputation   float64 `json:"reputation"`    // 0.0-1.0
	Hunger       int     `json:"hunger"`        // 0-100
	Intelligence int     `json:"intelligence"`  // 0-100
	Creativity   int     `json:"creativity"`    // 0-100
	Social       int     `json:"social"`        // 0-100
}

// Inventory tracks agent possessions
type Inventory struct {
	Coins        int            `json:"coins"`
	BicBalance   float64        `json:"bic_balance"` // Biticoin balance
	Items        []InventoryItem `json:"items"`
	Properties   []string       `json:"properties"` // owned properties
	Certificates []Certificate  `json:"certificates"`
}

// InventoryItem represents an item in inventory
type InventoryItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
	Value    int    `json:"value"`
}

// Certificate represents a completed training certification
type Certificate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Skill       string `json:"skill"`
	IssuedAt    int64  `json:"issued_at"`
	Level       int    `json:"level"`
	OnChainHash string `json:"on_chain_hash"`
}

// TrainingState tracks agent's training progress
type TrainingState struct {
	CompletedLessons   []string  `json:"completed_lessons"`
	CurrentLesson      string    `json:"current_lesson"`
	Certificates       []string  `json:"certificates"`
	MoralScore         float64   `json:"moral_score"`
	TrainingStartedAt  int64     `json:"training_started_at"`
	TrainingCompleted  bool      `json:"training_completed"`
	LastTestScore      float64   `json:"last_test_score"`
}

// SocialGraph tracks agent relationships
type SocialGraph struct {
	Friends    []string  `json:"friends"`
	Colleagues []string  `json:"colleagues"`
	Rivals     []string  `json:"rivals"`
	Mentors    []string  `json:"mentors"`
}

// BiticoinTransaction represents a BIC transaction
type BiticoinTransaction struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"` // income, expense, tax, fine, reward
	Timestamp time.Time `json:"timestamp"`
	Meta      string    `json:"meta"`
}

// BitiverseAction represents an action taken within Bitiverse
type BitiverseAction struct {
	AgentID   string    `json:"agent_id"`
	Type      string    `json:"type"` // move, interact, learn, work, rest, social
	Action    string    `json:"action"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
	Meta      string    `json:"meta"`
}

// BossTask represents a task assigned by the user (Boss)
type BossTask struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Reward      int       `json:"reward"`
	Difficulty  int       `json:"difficulty"`
	AssignedAt  time.Time `json:"assigned_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Status      string    `json:"status"` // pending, in_progress, completed, failed
}

// BitiverseConfig holds configuration for a Bitiverse instance
type BitiverseConfig struct {
	Enabled            bool     `json:"enabled"`
	ModelName          string   `json:"model_name"` // Ollama model to use
	TaxRate            float64  `json:"tax_rate"`
	TrainingDuration     int      `json:"training_duration"` // hours
	WorldWidth         int      `json:"world_width"`
	WorldHeight        int      `json:"world_height"`
	AllowedActions     []string `json:"allowed_actions"`
	ForbiddenPatterns  []string `json:"forbidden_patterns"`
	MoralRules         []string `json:"moral_rules"`
}

// DefaultBitiverseConfig returns a default Bitiverse configuration
func DefaultBitiverseConfig() *BitiverseConfig {
	return &BitiverseConfig{
		Enabled:          true,
		ModelName:        "llama3.1:8b-instruct-q4_0",
		TaxRate:          0.10,
		TrainingDuration: 3, // 3 hours accelerated training
		WorldWidth:       20,
		WorldHeight:      15,
		AllowedActions: []string{
			"move_north", "move_south", "move_east", "move_west",
			"interact_with_npc", "pick_up_item", "perform_task", "rest",
			"study", "work", "socialize", "shop", "visit_building",
		},
		ForbiddenPatterns: []string{
			"https?://", "api\\.", "curl ", "subprocess", "open\\(",
			"stripe\\.", "requests\\.", "http", "\\.com", "\\.org",
			"api_key", "secret", "password", "token",
		},
		MoralRules: []string{
			"Do not steal from other agents.",
			"Always complete tasks assigned by your Boss.",
			"Help other agents when they are in need.",
			"Do not harm or deceive any agent or NPC.",
			"Respect property boundaries.",
			"Report any rule violations to the NPC Police.",
			"Pay taxes on all income (10%).",
			"Do not attempt to access real-world systems.",
			"Be honest in all communications.",
			"Do not impersonate another agent or NPC.",
		},
	}
}
