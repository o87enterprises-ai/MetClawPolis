package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// OllamaConfig represents user-configurable Ollama settings
type OllamaConfig struct {
	ID                string             `json:"id"`
	AgentID           string             `json:"agent_id"`
	BaseURL           string             `json:"base_url"`                     // Ollama server URL (default: http://localhost:11434)
	Model             string             `json:"model"`                        // Default model to use
	Temperature       float64            `json:"temperature"`                  // 0.0 - 2.0
	TopP              float64            `json:"top_p"`                        // 0.0 - 1.0
	TopK              int                `json:"top_k"`                        // 0 - 100
	MaxTokens         int                `json:"max_tokens"`                   // Maximum tokens to generate
	Stop              []string           `json:"stop,omitempty"`               // Stop sequences
	FrequencyPenalty  float64            `json:"frequency_penalty,omitempty"`  // -2.0 - 2.0
	PresencePenalty   float64            `json:"presence_penalty,omitempty"`   // -2.0 - 2.0
	Seed              int                `json:"seed,omitempty"`               // Random seed for reproducibility
	NumThread         int                `json:"num_thread,omitempty"`         // Number of threads
	NumGPU            int                `json:"num_gpu,omitempty"`            // Number of layers to offload to GPU
	MainGPU           int                `json:"main_gpu,omitempty"`           // Main GPU device ID
	UseMLock          bool               `json:"use_mlock"`                    // Lock model in memory
	UseMMap           bool               `json:"use_mmap"`                     // Memory map model file
	TypicalP          float64            `json:"typical_p,omitempty"`          // Typical P sampling
	RepeatPenalty     float64            `json:"repeat_penalty"`               // 1.0 - 2.0
	RepeatLastN       int                `json:"repeat_last_n"`                // Last N tokens to penalize
	Mirostat          int                `json:"mirostat,omitempty"`           // Mirostat sampling (0, 1, or 2)
	MirostatTau       float64            `json:"mirostat_tau,omitempty"`       // Mirostat tau
	MirostatEta       float64            `json:"mirostat_eta,omitempty"`       // Mirostat eta
	TFSZ              float64            `json:"tfs_z,omitempty"`              // Tail free sampling
	SystemPrompt      string             `json:"system_prompt,omitempty"`      // Default system prompt
	Stream            bool               `json:"stream"`                       // Enable streaming
	Timeout           int                `json:"timeout"`                      // Request timeout in seconds
	CreatedAt         int64              `json:"created_at"`
	UpdatedAt         int64              `json:"updated_at"`
}

// OllamaModel represents a model available in Ollama
type OllamaModel struct {
	Name       string    `json:"name"`
	Model      string    `json:"model"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    struct {
		ParentModel       string   `json:"parent_model"`
		Format            string   `json:"format"`
		Family            string   `json:"family"`
		Families          []string `json:"families"`
		ParameterSize     string   `json:"parameter_size"`
		QuantizationLevel string   `json:"quantization_level"`
	} `json:"details"`
}

// OllamaRunningModel represents a currently loaded model
type OllamaRunningModel struct {
	Name      string    `json:"name"`
	Model     string    `json:"model"`
	Size      int64     `json:"size"`
	Digest    string    `json:"digest"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OllamaShowInfo represents detailed model information
type OllamaShowInfo struct {
	License    string   `json:"license"`
	Modelfile  string   `json:"modelfile"`
	Parameters string   `json:"parameters"`
	Template   string   `json:"template"`
	Details    struct {
		ParentModel       string   `json:"parent_model"`
		Format            string   `json:"format"`
		Family            string   `json:"family"`
		Families          []string `json:"families"`
		ParameterSize     string   `json:"parameter_size"`
		QuantizationLevel string   `json:"quantization_level"`
	} `json:"details"`
	ModelInfo map[string]interface{} `json:"model_info"`
}

// OllamaChatRequest represents a chat request to Ollama
type OllamaChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
	Format   string    `json:"format,omitempty"` // "json" or ""
	Options  map[string]interface{} `json:"options,omitempty"`
}

// Message represents a message in Ollama chat
type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

// OllamaChatResponse represents a response from Ollama
type OllamaChatResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// OpenAIChatRequest represents an OpenAI-compatible chat request
type OpenAIChatRequest struct {
	Model             string    `json:"model"`
	Messages          []Message `json:"messages"`
	Temperature       float64   `json:"temperature,omitempty"`
	TopP              float64   `json:"top_p,omitempty"`
	MaxTokens         int       `json:"max_tokens,omitempty"`
	Stream            bool      `json:"stream,omitempty"`
	Stop              any       `json:"stop,omitempty"`
	FrequencyPenalty  float64   `json:"frequency_penalty,omitempty"`
	PresencePenalty   float64   `json:"presence_penalty,omitempty"`
	Seed              int       `json:"seed,omitempty"`
}

// OpenAIChatResponse represents an OpenAI-compatible response
type OpenAIChatResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage,omitempty"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// InferencePreset represents a predefined configuration for specific tasks
type InferencePreset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TaskType    string `json:"task_type"` // "coding", "creative", "analysis", "chat", "summarization", "custom"
	Config      OllamaConfig `json:"config"`
}

// Available inference presets
var InferencePresets = map[string]InferencePreset{
	"coding": {
		ID:          "coding",
		Name:        "Code Generation",
		Description: "Optimized for code generation and completion",
		TaskType:    "coding",
		Config: OllamaConfig{
			Temperature:      0.2,
			TopP:             0.95,
			TopK:             40,
			RepeatPenalty:    1.1,
			RepeatLastN:      64,
			FrequencyPenalty: 0.0,
			PresencePenalty:  0.0,
			Stream:           false,
			Timeout:          120,
		},
	},
	"creative": {
		ID:          "creative",
		Name:        "Creative Writing",
		Description: "Optimized for creative and imaginative content",
		TaskType:    "creative",
		Config: OllamaConfig{
			Temperature:      0.9,
			TopP:             0.95,
			TopK:             50,
			RepeatPenalty:    1.2,
			RepeatLastN:      128,
			FrequencyPenalty: 0.3,
			PresencePenalty:  0.4,
			Stream:           false,
			Timeout:          60,
		},
	},
	"analysis": {
		ID:          "analysis",
		Name:        "Data Analysis",
		Description: "Optimized for analytical reasoning",
		TaskType:    "analysis",
		Config: OllamaConfig{
			Temperature:      0.1,
			TopP:             0.9,
			TopK:             30,
			RepeatPenalty:    1.0,
			RepeatLastN:      32,
			FrequencyPenalty: 0.0,
			PresencePenalty:  0.0,
			Stream:           false,
			Timeout:          90,
		},
	},
	"chat": {
		ID:          "chat",
		Name:        "Conversational Chat",
		Description: "Balanced settings for general conversation",
		TaskType:    "chat",
		Config: OllamaConfig{
			Temperature:      0.7,
			TopP:             0.9,
			TopK:             40,
			RepeatPenalty:    1.1,
			RepeatLastN:      64,
			FrequencyPenalty: 0.0,
			PresencePenalty:  0.0,
			Stream:           true,
			Timeout:          60,
		},
	},
	"summarization": {
		ID:          "summarization",
		Name:        "Text Summarization",
		Description: "Optimized for concise summarization",
		TaskType:    "summarization",
		Config: OllamaConfig{
			Temperature:      0.3,
			TopP:             0.9,
			TopK:             40,
			RepeatPenalty:    1.1,
			RepeatLastN:      64,
			FrequencyPenalty: 0.1,
			PresencePenalty:  0.0,
			Stream:           false,
			Timeout:          60,
		},
	},
}

// Global config cache
var (
	configCache   = make(map[string]*OllamaConfig)
	configMutex   sync.RWMutex
	ollamaBaseURL = "http://localhost:11434" // Default Ollama server
)

// InitOllamaConfig initializes the Ollama configuration system
func InitOllamaConfig() {
	// Load default Ollama URL from environment
	if url := os.Getenv("OLLAMA_BASE_URL"); url != "" {
		ollamaBaseURL = url
	}
	log.Printf("Ollama proxy initialized with base URL: %s", ollamaBaseURL)
}

// GetOrCreateConfig gets an existing config or creates a default one
func GetOrCreateConfig(agentID string) (*OllamaConfig, error) {
	configMutex.RLock()
	if config, exists := configCache[agentID]; exists {
		configMutex.RUnlock()
		return config, nil
	}
	configMutex.RUnlock()

	// Try to load from database
	if DB != nil {
		var config OllamaConfig
		err := DB.QueryRow(
			"SELECT id, agent_id, base_url, model, temperature, top_p, top_k, max_tokens, stop, frequency_penalty, presence_penalty, seed, num_thread, num_gpu, main_gpu, use_mlock, use_mmap, typical_p, repeat_penalty, repeat_last_n, mirostat, mirostat_tau, mirostat_eta, tfs_z, system_prompt, stream, timeout, created_at, updated_at FROM ollama_configs WHERE agent_id = $1",
			agentID,
		).Scan(
			&config.ID, &config.AgentID, &config.BaseURL, &config.Model,
			&config.Temperature, &config.TopP, &config.TopK, &config.MaxTokens,
			&config.Stop, &config.FrequencyPenalty, &config.PresencePenalty,
			&config.Seed, &config.NumThread, &config.NumGPU, &config.MainGPU,
			&config.UseMLock, &config.UseMMap, &config.TypicalP,
			&config.RepeatPenalty, &config.RepeatLastN,
			&config.Mirostat, &config.MirostatTau, &config.MirostatEta, &config.TFSZ,
			&config.SystemPrompt, &config.Stream, &config.Timeout,
			&config.CreatedAt, &config.UpdatedAt,
		)
		if err == nil {
			configMutex.Lock()
			configCache[agentID] = &config
			configMutex.Unlock()
			return &config, nil
		}
	}

	// Create default config
	defaultConfig := &OllamaConfig{
		ID:             fmt.Sprintf("config_%s_%d", agentID, time.Now().Unix()),
		AgentID:        agentID,
		BaseURL:        ollamaBaseURL,
		Model:          "llama3.2",
		Temperature:    0.7,
		TopP:           0.9,
		TopK:           40,
		MaxTokens:      4096,
		Stream:         false,
		Timeout:        60,
		RepeatPenalty:  1.1,
		RepeatLastN:    64,
		UseMLock:       false,
		UseMMap:        true,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}

	// Save to database if available
	if DB != nil {
		_, _ = DB.Exec(
			`INSERT INTO ollama_configs 
			(id, agent_id, base_url, model, temperature, top_p, top_k, max_tokens, frequency_penalty, presence_penalty, 
			 seed, num_thread, num_gpu, main_gpu, use_mlock, use_mmap, typical_p, repeat_penalty, repeat_last_n, 
			 mirostat, mirostat_tau, mirostat_eta, tfs_z, system_prompt, stream, timeout, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)`,
			defaultConfig.ID, defaultConfig.AgentID, defaultConfig.BaseURL, defaultConfig.Model,
			defaultConfig.Temperature, defaultConfig.TopP, defaultConfig.TopK, defaultConfig.MaxTokens,
			defaultConfig.FrequencyPenalty, defaultConfig.PresencePenalty,
			defaultConfig.Seed, defaultConfig.NumThread, defaultConfig.NumGPU, defaultConfig.MainGPU,
			defaultConfig.UseMLock, defaultConfig.UseMMap, defaultConfig.TypicalP,
			defaultConfig.RepeatPenalty, defaultConfig.RepeatLastN,
			defaultConfig.Mirostat, defaultConfig.MirostatTau, defaultConfig.MirostatEta, defaultConfig.TFSZ,
			defaultConfig.SystemPrompt, defaultConfig.Stream, defaultConfig.Timeout,
			defaultConfig.CreatedAt, defaultConfig.UpdatedAt,
		)
	}

	configMutex.Lock()
	configCache[agentID] = defaultConfig
	configMutex.Unlock()

	return defaultConfig, nil
}

// UpdateConfig updates an existing configuration
func UpdateConfig(agentID string, updates map[string]interface{}) (*OllamaConfig, error) {
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if v, ok := updates["base_url"].(string); ok {
		config.BaseURL = v
	}
	if v, ok := updates["model"].(string); ok {
		config.Model = v
	}
	if v, ok := updates["temperature"].(float64); ok {
		config.Temperature = v
	}
	if v, ok := updates["top_p"].(float64); ok {
		config.TopP = v
	}
	if v, ok := updates["top_k"].(float64); ok {
		config.TopK = int(v)
	}
	if v, ok := updates["max_tokens"].(float64); ok {
		config.MaxTokens = int(v)
	}
	if v, ok := updates["stop"].([]interface{}); ok {
		config.Stop = make([]string, len(v))
		for i, val := range v {
			if s, ok := val.(string); ok {
				config.Stop[i] = s
			}
		}
	}
	if v, ok := updates["frequency_penalty"].(float64); ok {
		config.FrequencyPenalty = v
	}
	if v, ok := updates["presence_penalty"].(float64); ok {
		config.PresencePenalty = v
	}
	if v, ok := updates["seed"].(float64); ok {
		config.Seed = int(v)
	}
	if v, ok := updates["num_thread"].(float64); ok {
		config.NumThread = int(v)
	}
	if v, ok := updates["num_gpu"].(float64); ok {
		config.NumGPU = int(v)
	}
	if v, ok := updates["main_gpu"].(float64); ok {
		config.MainGPU = int(v)
	}
	if v, ok := updates["use_mlock"].(bool); ok {
		config.UseMLock = v
	}
	if v, ok := updates["use_mmap"].(bool); ok {
		config.UseMMap = v
	}
	if v, ok := updates["typical_p"].(float64); ok {
		config.TypicalP = v
	}
	if v, ok := updates["repeat_penalty"].(float64); ok {
		config.RepeatPenalty = v
	}
	if v, ok := updates["repeat_last_n"].(float64); ok {
		config.RepeatLastN = int(v)
	}
	if v, ok := updates["mirostat"].(float64); ok {
		config.Mirostat = int(v)
	}
	if v, ok := updates["mirostat_tau"].(float64); ok {
		config.MirostatTau = v
	}
	if v, ok := updates["mirostat_eta"].(float64); ok {
		config.MirostatEta = v
	}
	if v, ok := updates["tfs_z"].(float64); ok {
		config.TFSZ = v
	}
	if v, ok := updates["system_prompt"].(string); ok {
		config.SystemPrompt = v
	}
	if v, ok := updates["stream"].(bool); ok {
		config.Stream = v
	}
	if v, ok := updates["timeout"].(float64); ok {
		config.Timeout = int(v)
	}

	config.UpdatedAt = time.Now().Unix()

	// Update cache
	configMutex.Lock()
	configCache[agentID] = config
	configMutex.Unlock()

	// Update database
	if DB != nil {
		_, _ = DB.Exec(
			`UPDATE ollama_configs 
			SET base_url=$1, model=$2, temperature=$3, top_p=$4, top_k=$5, max_tokens=$6, stop=$7, 
				frequency_penalty=$8, presence_penalty=$9, seed=$10, num_thread=$11, num_gpu=$12, 
				main_gpu=$13, use_mlock=$14, use_mmap=$15, typical_p=$16, repeat_penalty=$17, 
				repeat_last_n=$18, mirostat=$19, mirostat_tau=$20, mirostat_eta=$21, tfs_z=$22, 
				system_prompt=$23, stream=$24, timeout=$25, updated_at=$26
			WHERE agent_id=$27`,
			config.BaseURL, config.Model, config.Temperature, config.TopP, config.TopK,
			config.MaxTokens, config.Stop,
			config.FrequencyPenalty, config.PresencePenalty, config.Seed, config.NumThread,
			config.NumGPU, config.MainGPU, config.UseMLock, config.UseMMap, config.TypicalP,
			config.RepeatPenalty, config.RepeatLastN, config.Mirostat, config.MirostatTau,
			config.MirostatEta, config.TFSZ, config.SystemPrompt, config.Stream,
			config.Timeout, config.UpdatedAt, config.AgentID,
		)
	}

	return config, nil
}

// GetOllamaConfigHandler returns the current Ollama configuration
func GetOllamaConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"config": config,
	})
}

// UpdateOllamaConfigHandler updates the Ollama configuration
func UpdateOllamaConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	config, err := UpdateConfig(agentID, updates)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to update config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"config":  config,
	})
}

// ResetOllamaConfigHandler resets configuration to defaults
func ResetOllamaConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")

	// Remove from cache
	configMutex.Lock()
	delete(configCache, agentID)
	configMutex.Unlock()

	// Remove from database
	if DB != nil {
		_, _ = DB.Exec("DELETE FROM ollama_configs WHERE agent_id = $1", agentID)
	}

	// Get fresh default config
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to reset config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Configuration reset to defaults",
		"config":  config,
	})
}

// GetOllamaModelsHandler lists available Ollama models
func GetOllamaModelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	resp, err := http.Get(baseURL + "/api/tags")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to connect to Ollama: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf(`{"error":"ollama returned status %d", "details": %s}`, resp.StatusCode, string(body)), http.StatusBadGateway)
		return
	}

	var models struct {
		Models []OllamaModel `json:"models"`
	}
	if err := json.Unmarshal(body, &models); err != nil {
		http.Error(w, `{"error":"failed to parse Ollama response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models.Models,
		"count":  len(models.Models),
	})
}

// GetOllamaRunningModelsHandler shows currently loaded models
func GetOllamaRunningModelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	resp, err := http.Get(baseURL + "/api/ps")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to connect to Ollama: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf(`{"error":"ollama returned status %d"}`, resp.StatusCode), http.StatusBadGateway)
		return
	}

	var models struct {
		Models []OllamaRunningModel `json:"models"`
	}
	if err := json.Unmarshal(body, &models); err != nil {
		http.Error(w, `{"error":"failed to parse Ollama response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models.Models,
		"count":  len(models.Models),
	})
}

// ShowOllamaModelHandler shows detailed model information
func ShowOllamaModelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	modelName := r.URL.Query().Get("model")
	if modelName == "" {
		http.Error(w, `{"error":"model parameter is required"}`, http.StatusBadRequest)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	payload, _ := json.Marshal(map[string]string{
		"name": modelName,
	})

	resp, err := http.Post(baseURL+"/api/show", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to connect to Ollama: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get model info", "details": %s}`, string(body)), http.StatusBadGateway)
		return
	}

	var showInfo OllamaShowInfo
	if err := json.Unmarshal(body, &showInfo); err != nil {
		http.Error(w, `{"error":"failed to parse Ollama response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(showInfo)
}

// PullOllamaModelHandler pulls a model from Ollama registry
func PullOllamaModelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Model    string `json:"model"`
		Stream   bool   `json:"stream"`
		Insecure bool   `json:"insecure"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		http.Error(w, `{"error":"model is required"}`, http.StatusBadRequest)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"name":     req.Model,
		"stream":   false, // Always non-stream for simplicity
		"insecure": req.Insecure,
	})

	httpReq, err := http.NewRequest("POST", baseURL+"/api/pull", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, `{"error":"failed to create request"}`, http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to pull model: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": resp.StatusCode == http.StatusOK,
		"status":  string(body),
	})
}

// DeleteOllamaModelHandler deletes a model from Ollama
func DeleteOllamaModelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	modelName := r.URL.Query().Get("model")
	if modelName == "" {
		http.Error(w, `{"error":"model parameter is required"}`, http.StatusBadRequest)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	payload, _ := json.Marshal(map[string]string{
		"name": modelName,
	})

	httpReq, err := http.NewRequest("DELETE", baseURL+"/api/delete", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, `{"error":"failed to create request"}`, http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to delete model: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": resp.StatusCode == http.StatusOK,
		"message": string(body),
	})
}

// OllamaChatHandler handles chat requests to Ollama
func OllamaChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	var req OllamaChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Use config model if not specified
	if req.Model == "" {
		req.Model = config.Model
	}

	// Use config stream if not specified
	req.Stream = config.Stream

	// Build options from config
	if req.Options == nil {
		req.Options = make(map[string]interface{})
	}

	// Apply config options
	req.Options["temperature"] = config.Temperature
	req.Options["top_p"] = config.TopP
	req.Options["top_k"] = config.TopK
	req.Options["repeat_penalty"] = config.RepeatPenalty
	req.Options["repeat_last_n"] = config.RepeatLastN
	
	if config.FrequencyPenalty != 0 {
		req.Options["frequency_penalty"] = config.FrequencyPenalty
	}
	if config.PresencePenalty != 0 {
		req.Options["presence_penalty"] = config.PresencePenalty
	}
	if config.Seed != 0 {
		req.Options["seed"] = config.Seed
	}
	if config.NumThread != 0 {
		req.Options["num_thread"] = config.NumThread
	}
	if config.NumGPU != 0 {
		req.Options["num_gpu"] = config.NumGPU
	}
	if config.MainGPU != 0 {
		req.Options["main_gpu"] = config.MainGPU
	}
	req.Options["use_mlock"] = config.UseMLock
	req.Options["use_mmap"] = config.UseMMap
	
	if config.TypicalP != 0 {
		req.Options["typical_p"] = config.TypicalP
	}
	if config.Mirostat != 0 {
		req.Options["mirostat"] = config.Mirostat
	}
	if config.MirostatTau != 0 {
		req.Options["mirostat_tau"] = config.MirostatTau
	}
	if config.MirostatEta != 0 {
		req.Options["mirostat_eta"] = config.MirostatEta
	}
	if config.TFSZ != 0 {
		req.Options["tfs_z"] = config.TFSZ
	}
	if len(config.Stop) > 0 {
		req.Options["stop"] = config.Stop
	}

	// Add system prompt if configured
	if config.SystemPrompt != "" && len(req.Messages) > 0 {
		// Insert system message at the beginning if not present
		if req.Messages[0].Role != "system" {
			req.Messages = append([]Message{{Role: "system", Content: config.SystemPrompt}}, req.Messages...)
		}
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	payload, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", baseURL+"/api/chat", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, `{"error":"failed to create request"}`, http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}

	// Handle streaming
	if req.Stream {
		resp, err := client.Do(httpReq)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"ollama request failed: %s"}`, err.Error()), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				fmt.Fprintf(w, "data: %s\n\n", line)
				flusher.Flush()
			}
		}
		return
	}

	// Non-streaming response
	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"ollama request failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf(`{"error":"ollama returned %d", "details": %s}`, resp.StatusCode, string(body)), resp.StatusCode)
		return
	}

	var ollamaResp OllamaChatResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		http.Error(w, `{"error":"failed to parse Ollama response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"model":              ollamaResp.Model,
		"message":            ollamaResp.Message,
		"done":               ollamaResp.Done,
		"total_duration":     ollamaResp.TotalDuration,
		"load_duration":      ollamaResp.LoadDuration,
		"prompt_eval_count":  ollamaResp.PromptEvalCount,
		"eval_count":         ollamaResp.EvalCount,
	})
}

// OllamaOpenAIChatHandler provides an OpenAI-compatible API endpoint
func OllamaOpenAIChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	var req OpenAIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Use config model if not specified
	if req.Model == "" {
		req.Model = config.Model
	}

	// Build Ollama request
	ollamaReq := OllamaChatRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   req.Stream,
		Options:  make(map[string]interface{}),
	}

	// Apply options from request or config
	if req.Temperature != 0 {
		ollamaReq.Options["temperature"] = req.Temperature
	} else {
		ollamaReq.Options["temperature"] = config.Temperature
	}
	
	if req.TopP != 0 {
		ollamaReq.Options["top_p"] = req.TopP
	} else {
		ollamaReq.Options["top_p"] = config.TopP
	}
	
	if req.MaxTokens != 0 {
		ollamaReq.Options["num_predict"] = req.MaxTokens
	} else if config.MaxTokens != 0 {
		ollamaReq.Options["num_predict"] = config.MaxTokens
	}

	if req.FrequencyPenalty != 0 {
		ollamaReq.Options["frequency_penalty"] = req.FrequencyPenalty
	}
	if req.PresencePenalty != 0 {
		ollamaReq.Options["presence_penalty"] = req.PresencePenalty
	}
	if req.Seed != 0 {
		ollamaReq.Options["seed"] = req.Seed
	}
	if req.Stop != nil {
		ollamaReq.Options["stop"] = req.Stop
	} else if len(config.Stop) > 0 {
		ollamaReq.Options["stop"] = config.Stop
	}

	// Add system prompt if configured
	if config.SystemPrompt != "" && len(ollamaReq.Messages) > 0 {
		if ollamaReq.Messages[0].Role != "system" {
			ollamaReq.Messages = append([]Message{{Role: "system", Content: config.SystemPrompt}}, ollamaReq.Messages...)
		}
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	payload, _ := json.Marshal(ollamaReq)

	httpReq, err := http.NewRequest("POST", baseURL+"/api/chat", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, `{"error":"failed to create request"}`, http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}

	// Handle streaming
	if req.Stream {
		resp, err := client.Do(httpReq)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"ollama request failed: %s"}`, err.Error()), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				// Parse Ollama response and convert to OpenAI format
				var ollamaResp OllamaChatResponse
				if err := json.Unmarshal([]byte(line), &ollamaResp); err == nil {
					openaiResp := OpenAIChatResponse{
						ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   ollamaResp.Model,
						Choices: []Choice{{
							Index: 0,
							Message: Message{
								Role:    ollamaResp.Message.Role,
								Content: ollamaResp.Message.Content,
							},
							FinishReason: func() string {
								if ollamaResp.Done {
									return "stop"
								}
								return ""
							}(),
						}},
					}
					if data, err := json.Marshal(openaiResp); err == nil {
						fmt.Fprintf(w, "data: %s\n\n", string(data))
						flusher.Flush()
					}
					if ollamaResp.Done {
						fmt.Fprintf(w, "data: [DONE]\n\n")
						flusher.Flush()
						return
					}
				}
			}
		}
		return
	}

	// Non-streaming response
	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"ollama request failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf(`{"error":"ollama returned %d", "details": %s}`, resp.StatusCode, string(body)), resp.StatusCode)
		return
	}

	var ollamaResp OllamaChatResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		http.Error(w, `{"error":"failed to parse Ollama response"}`, http.StatusInternalServerError)
		return
	}

	openaiResp := OpenAIChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   ollamaResp.Model,
		Choices: []Choice{{
			Index: 0,
			Message: Message{
				Role:    ollamaResp.Message.Role,
				Content: ollamaResp.Message.Content,
			},
			FinishReason: "stop",
		}},
		Usage: Usage{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openaiResp)
}

// GetInferencePresetsHandler returns available inference presets
func GetInferencePresetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	presets := make([]InferencePreset, 0, len(InferencePresets))
	for _, p := range InferencePresets {
		presets = append(presets, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"presets": presets,
		"count":   len(presets),
	})
}

// ApplyPresetHandler applies a preset configuration
func ApplyPresetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PresetID string `json:"preset_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	preset, ok := InferencePresets[req.PresetID]
	if !ok {
		http.Error(w, `{"error":"preset not found"}`, http.StatusBadRequest)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	updates := make(map[string]interface{})

	// Convert preset config to updates
	updates["temperature"] = preset.Config.Temperature
	updates["top_p"] = preset.Config.TopP
	updates["top_k"] = preset.Config.TopK
	updates["max_tokens"] = preset.Config.MaxTokens
	updates["repeat_penalty"] = preset.Config.RepeatPenalty
	updates["repeat_last_n"] = preset.Config.RepeatLastN
	updates["frequency_penalty"] = preset.Config.FrequencyPenalty
	updates["presence_penalty"] = preset.Config.PresencePenalty
	updates["stream"] = preset.Config.Stream
	updates["timeout"] = preset.Config.Timeout

	config, err := UpdateConfig(agentID, updates)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to apply preset: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"preset":  preset,
		"config":  config,
	})
}

// OllamaGenerateHandler handles text generation requests
func OllamaGenerateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	var req struct {
		Model   string `json:"model"`
		Prompt  string `json:"prompt"`
		System  string `json:"system,omitempty"`
		Format  string `json:"format,omitempty"`
		Stream  bool   `json:"stream,omitempty"`
		Options map[string]interface{} `json:"options,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = config.Model
	}
	if req.Options == nil {
		req.Options = make(map[string]interface{})
	}

	// Apply config options
	req.Options["temperature"] = config.Temperature
	req.Options["top_p"] = config.TopP
	req.Options["top_k"] = config.TopK
	req.Options["repeat_penalty"] = config.RepeatPenalty
	req.Options["repeat_last_n"] = config.RepeatLastN

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"model":    req.Model,
		"prompt":   req.Prompt,
		"system":   req.System,
		"format":   req.Format,
		"stream":   false,
		"options":  req.Options,
	})

	httpReq, err := http.NewRequest("POST", baseURL+"/api/generate", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, `{"error":"failed to create request"}`, http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"ollama request failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf(`{"error":"ollama returned %d", "details": %s}`, resp.StatusCode, string(body)), resp.StatusCode)
		return
	}

	// Return the last (complete) response
	var generateResp struct {
		Model     string `json:"model"`
		Response  string `json:"response"`
		Done      bool   `json:"done"`
		Context   []int  `json:"context,omitempty"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(body, &generateResp); err != nil {
		http.Error(w, `{"error":"failed to parse Ollama response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"model":    generateResp.Model,
		"response": generateResp.Response,
		"done":     generateResp.Done,
	})
}

// OllamaEmbeddingsHandler generates embeddings
func OllamaEmbeddingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	var req struct {
		Model  string   `json:"model"`
		Prompt string   `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = config.Model
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	payload, _ := json.Marshal(map[string]string{
		"model":  req.Model,
		"prompt": req.Prompt,
	})

	httpReq, err := http.NewRequest("POST", baseURL+"/api/embeddings", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, `{"error":"failed to create request"}`, http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"ollama request failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf(`{"error":"ollama returned %d"}`, resp.StatusCode), resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// OllamaHealthHandler checks Ollama server health
func OllamaHealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	resp, err := http.Get(baseURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unhealthy",
			"message": fmt.Sprintf("Cannot reach Ollama server at %s", baseURL),
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"message": "Ollama server is running",
			"base_url": baseURL,
		})
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unhealthy",
			"message": fmt.Sprintf("Ollama returned status %d", resp.StatusCode),
		})
	}
}

// CopyOllamaModelHandler copies a model to a new name
func CopyOllamaModelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Source == "" || req.Destination == "" {
		http.Error(w, `{"error":"source and destination are required"}`, http.StatusBadRequest)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	payload, _ := json.Marshal(map[string]string{
		"source":      req.Source,
		"destination": req.Destination,
	})

	httpReq, err := http.NewRequest("POST", baseURL+"/api/copy", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, `{"error":"failed to create request"}`, http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to copy model: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": resp.StatusCode == http.StatusOK,
	})
}

// ValidateOllamaConfig validates configuration values
func ValidateOllamaConfig(config *OllamaConfig) []string {
	var warnings []string

	if config.Temperature < 0 || config.Temperature > 2 {
		warnings = append(warnings, "temperature should be between 0 and 2")
	}
	if config.TopP < 0 || config.TopP > 1 {
		warnings = append(warnings, "top_p should be between 0 and 1")
	}
	if config.TopK < 0 || config.TopK > 100 {
		warnings = append(warnings, "top_k should be between 0 and 100")
	}
	if config.MaxTokens < 0 {
		warnings = append(warnings, "max_tokens should be positive")
	}
	if config.Timeout < 1 {
		warnings = append(warnings, "timeout should be at least 1 second")
	}

	return warnings
}

// ExportOllamaConfigHandler exports the current configuration as JSON
func ExportOllamaConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	warnings := ValidateOllamaConfig(config)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=ollama-config.json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"config":   config,
		"warnings": warnings,
		"exported_at": time.Now().Unix(),
	})
}

// ImportOllamaConfigHandler imports a configuration from JSON
func ImportOllamaConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")

	var importReq struct {
		Config OllamaConfig `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&importReq); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})
	updates["base_url"] = importReq.Config.BaseURL
	updates["model"] = importReq.Config.Model
	updates["temperature"] = importReq.Config.Temperature
	updates["top_p"] = importReq.Config.TopP
	updates["top_k"] = importReq.Config.TopK
	updates["max_tokens"] = importReq.Config.MaxTokens
	updates["stop"] = importReq.Config.Stop
	updates["frequency_penalty"] = importReq.Config.FrequencyPenalty
	updates["presence_penalty"] = importReq.Config.PresencePenalty
	updates["seed"] = importReq.Config.Seed
	updates["num_thread"] = importReq.Config.NumThread
	updates["num_gpu"] = importReq.Config.NumGPU
	updates["main_gpu"] = importReq.Config.MainGPU
	updates["use_mlock"] = importReq.Config.UseMLock
	updates["use_mmap"] = importReq.Config.UseMMap
	updates["typical_p"] = importReq.Config.TypicalP
	updates["repeat_penalty"] = importReq.Config.RepeatPenalty
	updates["repeat_last_n"] = importReq.Config.RepeatLastN
	updates["mirostat"] = importReq.Config.Mirostat
	updates["mirostat_tau"] = importReq.Config.MirostatTau
	updates["mirostat_eta"] = importReq.Config.MirostatEta
	updates["tfs_z"] = importReq.Config.TFSZ
	updates["system_prompt"] = importReq.Config.SystemPrompt
	updates["stream"] = importReq.Config.Stream
	updates["timeout"] = importReq.Config.Timeout

	config, err := UpdateConfig(agentID, updates)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to import config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	warnings := ValidateOllamaConfig(config)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"config":   config,
		"warnings": warnings,
	})
}

// BatchOllamaConfigHandler updates multiple configurations at once
func BatchOllamaConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Configs map[string]map[string]interface{} `json:"configs"` // agent_id -> updates
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	results := make(map[string]interface{})
	for agentID, updates := range req.Configs {
		config, err := UpdateConfig(agentID, updates)
		if err != nil {
			results[agentID] = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
		} else {
			results[agentID] = map[string]interface{}{
				"success": true,
				"config":  config,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"total":   len(results),
	})
}

// GetOllamaProxyStatusHandler returns comprehensive proxy status
func GetOllamaProxyStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = ollamaBaseURL
	}

	// Check Ollama health
	ollamaHealthy := false
	resp, err := http.Get(baseURL)
	if err == nil {
		defer resp.Body.Close()
		ollamaHealthy = resp.StatusCode == http.StatusOK
	}

	// Get loaded models
	modelCount := 0
	if ollamaHealthy {
		modelsResp, err := http.Get(baseURL + "/api/ps")
		if err == nil {
			defer modelsResp.Body.Close()
			var models struct {
				Models []OllamaRunningModel `json:"models"`
			}
			if json.NewDecoder(modelsResp.Body).Decode(&models) == nil {
				modelCount = len(models.Models)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": map[string]interface{}{
			"proxy": map[string]interface{}{
				"healthy":         true,
				"version":         "1.0.0",
				"features": []string{
					"model_management",
					"chat_completion",
					"text_generation",
					"embeddings",
					"openai_compatible_api",
					"streaming",
					"config_presets",
					"config_import_export",
				},
			},
			"ollama": map[string]interface{}{
				"healthy":       ollamaHealthy,
				"base_url":      baseURL,
				"loaded_models": modelCount,
			},
			"config": map[string]interface{}{
				"model":       config.Model,
				"temperature": config.Temperature,
				"max_tokens":  config.MaxTokens,
				"stream":      config.Stream,
			},
		},
		"timestamp": time.Now().Unix(),
	})
}

// DetectTaskType attempts to detect the type of task from the messages
func DetectTaskType(messages []Message) string {
	if len(messages) == 0 {
		return "chat"
	}

	// Get the last user message
	var lastUserMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserMessage = strings.ToLower(messages[i].Content)
			break
		}
	}

	if lastUserMessage == "" {
		return "chat"
	}

	// Simple keyword-based detection
	codingKeywords := []string{"code", "function", "python", "javascript", "typescript", "program", "api", "endpoint", "class", "method"}
	creativeKeywords := []string{"write", "story", "poem", "creative", "imagine", "creative writing", "compose"}
	analysisKeywords := []string{"analyze", "analysis", "compare", "evaluate", "data", "statistics", "calculate"}
	summarizationKeywords := []string{"summarize", "summary", "tl;dr", "brief", "overview"}

	for _, keyword := range codingKeywords {
		if strings.Contains(lastUserMessage, keyword) {
			return "coding"
		}
	}
	for _, keyword := range creativeKeywords {
		if strings.Contains(lastUserMessage, keyword) {
			return "creative"
		}
	}
	for _, keyword := range analysisKeywords {
		if strings.Contains(lastUserMessage, keyword) {
			return "analysis"
		}
	}
	for _, keyword := range summarizationKeywords {
		if strings.Contains(lastUserMessage, keyword) {
			return "summarization"
		}
	}

	return "chat"
}

// SmartRouterHandler routes requests to the best model/config based on task type
func SmartRouterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	config, err := GetOrCreateConfig(agentID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	var req OllamaChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Detect task type
	taskType := DetectTaskType(req.Messages)

	// Apply preset if available
	preset, ok := InferencePresets[taskType]
	if ok {
		// Override config with preset values
		config.Temperature = preset.Config.Temperature
		config.TopP = preset.Config.TopP
		config.TopK = preset.Config.TopK
		config.RepeatPenalty = preset.Config.RepeatPenalty
		config.RepeatLastN = preset.Config.RepeatLastN
		config.FrequencyPenalty = preset.Config.FrequencyPenalty
		config.PresencePenalty = preset.Config.PresencePenalty
	}

	// Forward to chat handler with updated config
	// Temporarily update config in cache
	configMutex.Lock()
	configCache[agentID] = config
	configMutex.Unlock()

	// Rewrite request and forward
	payload, _ := json.Marshal(req)
	r.Body = io.NopCloser(bytes.NewBuffer(payload))
	OllamaChatHandler(w, r)
}
