package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// ServiceCredential stores encrypted credentials for external services
type ServiceCredential struct {
	ID          string `json:"id"`
	AgentID     string `json:"agent_id"`
	ServiceName string `json:"service_name"` // ollama, huggingface, github
	EncryptedToken string `json:"encrypted_token"`
	CreatedAt   int64  `json:"created_at"`
}

// For now we use base64 encoding. In production, use AES-GCM encryption.
func encryptToken(token string) string {
	return base64.StdEncoding.EncodeToString([]byte(token))
}

func decryptToken(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// SaveServiceCredential stores an encrypted credential for an agent
func SaveServiceCredential(agentID, serviceName, token string) error {
	id := uuid.New().String()
	encrypted := encryptToken(token)
	_, err := DB.Exec(
		"INSERT INTO agent_credentials (id, agent_id, service_name, encrypted_token, created_at) VALUES ($1, $2, $3, $4, $5)",
		id, agentID, serviceName, encrypted, time.Now().Unix(),
	)
	return err
}

// GetServiceCredential retrieves and decrypts a credential
func GetServiceCredential(agentID, serviceName string) (string, error) {
	var encrypted string
	err := DB.QueryRow(
		"SELECT encrypted_token FROM agent_credentials WHERE agent_id=$1 AND service_name=$2 ORDER BY created_at DESC LIMIT 1",
		agentID, serviceName,
	).Scan(&encrypted)
	if err != nil {
		return "", err
	}
	return decryptToken(encrypted)
}

// ServiceStatus represents the connection status of an external service
type ServiceStatus struct {
	Service     string `json:"service"`
	Connected   bool   `json:"connected"`
	AccountInfo string `json:"account_info,omitempty"`
	Error       string `json:"error,omitempty"`
	LastChecked int64  `json:"last_checked"`
}

// CheckOllamaStatus checks if Ollama is accessible and returns model info
func CheckOllamaStatus() ServiceStatus {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	resp, err := http.Get(baseURL + "/api/tags")
	if err != nil {
		return ServiceStatus{Service: "ollama", Connected: false, Error: err.Error(), LastChecked: time.Now().Unix()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var models map[string]interface{}
	json.Unmarshal(body, &models)

	modelCount := 0
	if tags, ok := models["models"].([]interface{}); ok {
		modelCount = len(tags)
	}

	return ServiceStatus{
		Service:     "ollama",
		Connected:   true,
		AccountInfo: fmt.Sprintf("%d models available", modelCount),
		LastChecked: time.Now().Unix(),
	}
}

// CheckHuggingFaceStatus validates the HF token and returns account info
func CheckHuggingFaceStatus(token string) ServiceStatus {
	if token == "" || token == "hf_..." {
		return ServiceStatus{Service: "huggingface", Connected: false, Error: "no token configured", LastChecked: time.Now().Unix()}
	}

	req, _ := http.NewRequest("GET", "https://huggingface.co/api/whoami-v2", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ServiceStatus{Service: "huggingface", Connected: false, Error: err.Error(), LastChecked: time.Now().Unix()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ServiceStatus{Service: "huggingface", Connected: false, Error: "invalid token", LastChecked: time.Now().Unix()}
	}

	var info map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&info)
	username, _ := info["name"].(string)
	if username == "" {
		username, _ = info["fullname"].(string)
	}

	return ServiceStatus{
		Service:     "huggingface",
		Connected:   true,
		AccountInfo: "User: " + username,
		LastChecked: time.Now().Unix(),
	}
}

// CheckGitHubStatus validates the GitHub token and returns account info
func CheckGitHubStatus(token string) ServiceStatus {
	if token == "" || token == "ghp_..." {
		return ServiceStatus{Service: "github", Connected: false, Error: "no token configured", LastChecked: time.Now().Unix()}
	}

	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ServiceStatus{Service: "github", Connected: false, Error: err.Error(), LastChecked: time.Now().Unix()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ServiceStatus{Service: "github", Connected: false, Error: "invalid token", LastChecked: time.Now().Unix()}
	}

	var info map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&info)
	login, _ := info["login"].(string)

	return ServiceStatus{
		Service:     "github",
		Connected:   true,
		AccountInfo: "User: " + login,
		LastChecked: time.Now().Unix(),
	}
}

// SyncServicesHandler returns status of all connected services
func SyncServicesHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.Header.Get("X-Verified-Agent-ID")

	var results []ServiceStatus

	// Check Ollama
	results = append(results, CheckOllamaStatus())

	// Check HuggingFace
	hfToken := os.Getenv("HF_TOKEN")
	if agentID != "" {
		if cred, err := GetServiceCredential(agentID, "huggingface"); err == nil {
			hfToken = cred
		}
	}
	results = append(results, CheckHuggingFaceStatus(hfToken))

	// Check GitHub
	ghToken := os.Getenv("GITHUB_TOKEN")
	if agentID != "" {
		if cred, err := GetServiceCredential(agentID, "github"); err == nil {
			ghToken = cred
		}
	}
	results = append(results, CheckGitHubStatus(ghToken))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services": results,
	})
}

// ConnectServiceHandler saves a credential for an agent
func ConnectServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")
	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Service string `json:"service"`
		Token   string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Service == "" || req.Token == "" {
		http.Error(w, `{"error":"service and token required"}`, http.StatusBadRequest)
		return
	}

	if err := SaveServiceCredential(agentID, req.Service, req.Token); err != nil {
		// Try to create table if it doesn't exist
		_, _ = DB.Exec(`CREATE TABLE IF NOT EXISTS agent_credentials (
			id TEXT PRIMARY KEY,
			agent_id TEXT REFERENCES agents(id),
			service_name TEXT NOT NULL,
			encrypted_token TEXT NOT NULL,
			created_at BIGINT NOT NULL
		)`)
		if err2 := SaveServiceCredential(agentID, req.Service, req.Token); err2 != nil {
			log.Printf("Failed to save credential: %v", err2)
			http.Error(w, `{"error":"failed to save credential"}`, http.StatusInternalServerError)
			return
		}
	}

	// Verify the token works
	var status ServiceStatus
	switch req.Service {
	case "huggingface":
		status = CheckHuggingFaceStatus(req.Token)
	case "github":
		status = CheckGitHubStatus(req.Token)
	case "ollama":
		status = CheckOllamaStatus()
	default:
		status = ServiceStatus{Service: req.Service, Connected: false, Error: "unknown service"}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
