package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// TerminalSession represents an active terminal session
type TerminalSession struct {
	ID        string
	AgentID   string
	SSHClient *ssh.Client
	Session   *ssh.Session
	Stdin     io.WriteCloser
	Stdout    io.Reader
	mu        sync.Mutex
}

var (
	terminalSessions = make(map[string]*TerminalSession)
	terminalMu       sync.RWMutex
)

// HandleTerminalWebSocket upgrades to WebSocket and provides a real terminal
func HandleTerminalWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Terminal WS upgrade error: %v", err)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Agent-ID")
	}
	if agentID == "" {
		agentID = "anonymous"
	}

	sessionID := fmt.Sprintf("term_%d_%s", time.Now().UnixMilli(), agentID)

	client := &WSClient{
		Conn:    conn,
		Send:    make(chan []byte, 256),
		AgentID: agentID,
	}

	WS.Register <- client
	defer func() {
		WS.Unregister <- client
		conn.Close()
		closeTerminalSession(sessionID)
	}()

	go client.writePump()

	// Create a local Docker exec terminal or SSH session
	host := r.URL.Query().Get("host")
	container := r.URL.Query().Get("container")

	if host != "" {
		// SSH to remote VM
		if err := connectSSH(sessionID, client, host); err != nil {
			client.Send <- []byte(fmt.Sprintf("\r\nSSH connection failed: %v\r\n", err))
			return
		}
	} else if container != "" {
		// Docker exec into container
		if err := connectDockerExec(sessionID, client, container); err != nil {
			client.Send <- []byte(fmt.Sprintf("\r\nDocker exec failed: %v\r\n", err))
			return
		}
	} else {
		// Local shell
		if err := connectLocalShell(sessionID, client); err != nil {
			client.Send <- []byte(fmt.Sprintf("\r\nShell failed: %v\r\n", err))
			return
		}
	}

	// Read input from WebSocket and forward to terminal
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse input: {"type":"input","data":"ls -la\n"}
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		msgType, _ := msg["type"].(string)
		if msgType == "input" {
			data, _ := msg["data"].(string)
			terminalMu.RLock()
			sess := terminalSessions[sessionID]
			terminalMu.RUnlock()
			if sess != nil && sess.Stdin != nil {
				sess.Stdin.Write([]byte(data))
			}
		} else if msgType == "resize" {
			rows, _ := msg["rows"].(float64)
			cols, _ := msg["cols"].(float64)
			terminalMu.RLock()
			sess := terminalSessions[sessionID]
			terminalMu.RUnlock()
			if sess != nil && sess.Session != nil {
				sess.Session.WindowChange(int(rows), int(cols))
			}
		}
	}
}

func connectLocalShell(sessionID string, client *WSClient) error {
	cmd := exec.Command("bash", "-i")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	terminalMu.Lock()
	terminalSessions[sessionID] = &TerminalSession{
		ID:    sessionID,
		Stdin: stdin,
		Stdout: stdout,
	}
	terminalMu.Unlock()

	// Stream output
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				client.Send <- buf[:n]
			}
			if err != nil {
				break
			}
		}
	}()

	client.Send <- []byte("\r\nMetClawPolis Terminal (local)\r\n$ ")
	return nil
}

func connectSSH(sessionID string, client *WSClient, host string) error {
	sshKeyPath := os.Getenv("VM_SSH_KEY_PATH")
	vmUser := os.Getenv("VM_USER")
	if vmUser == "" {
		vmUser = "ubuntu"
	}
	sshPort := os.Getenv("VM_SSH_PORT")
	if sshPort == "" {
		sshPort = "22"
	}

	var authMethod ssh.AuthMethod

	if sshKeyPath != "" && sshKeyPath != "/path/to/id_ed25519" {
		key, err := os.ReadFile(sshKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read SSH key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse SSH key: %w", err)
		}
		authMethod = ssh.PublicKeys(signer)
	} else {
		return fmt.Errorf("no SSH key configured (set VM_SSH_KEY_PATH)")
	}

	config := &ssh.ClientConfig{
		User: vmUser,
		Auth: []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: 10 * time.Second,
	}

	sshClient, err := ssh.Dial("tcp", host+":"+sshPort, config)
	if err != nil {
		return err
	}

	session, err := sshClient.NewSession()
	if err != nil {
		sshClient.Close()
		return err
	}

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	session.Stderr = session.Stdout

	// Request PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 24, 80, modes); err != nil {
		session.Close()
		sshClient.Close()
		return err
	}

	if err := session.Shell(); err != nil {
		session.Close()
		sshClient.Close()
		return err
	}

	terminalMu.Lock()
	terminalSessions[sessionID] = &TerminalSession{
		ID:        sessionID,
		SSHClient: sshClient,
		Session:   session,
		Stdin:     stdin,
		Stdout:    stdout,
	}
	terminalMu.Unlock()

	// Stream output
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				client.Send <- buf[:n]
			}
			if err != nil {
				break
			}
		}
	}()

	client.Send <- []byte(fmt.Sprintf("\r\nConnected to %s@%s\r\n", vmUser, host))
	return nil
}

func connectDockerExec(sessionID string, client *WSClient, container string) error {
	cmd := exec.Command("docker", "exec", "-it", container, "bash")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	terminalMu.Lock()
	terminalSessions[sessionID] = &TerminalSession{
		ID:    sessionID,
		Stdin: stdin,
		Stdout: stdout,
	}
	terminalMu.Unlock()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				client.Send <- buf[:n]
			}
			if err != nil {
				break
			}
		}
	}()

	client.Send <- []byte(fmt.Sprintf("\r\nDocker exec: %s\r\n", container))
	return nil
}

func closeTerminalSession(sessionID string) {
	terminalMu.Lock()
	defer terminalMu.Unlock()

	if sess, ok := terminalSessions[sessionID]; ok {
		if sess.Session != nil {
			sess.Session.Close()
		}
		if sess.SSHClient != nil {
			sess.SSHClient.Close()
		}
		delete(terminalSessions, sessionID)
	}
}

// ── Tunneling ──

var (
	tunnelProcess *exec.Cmd
	tunnelURL     string
	tunnelMu      sync.RWMutex
)

// StartTunnelHandler starts a tunnel (ngrok or cloudflared)
func StartTunnelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Port int `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Port = 8080
	}

	provider := os.Getenv("TUNNEL_PROVIDER")
	if provider == "" {
		provider = "ngrok"
	}

	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	// Stop existing tunnel
	if tunnelProcess != nil && tunnelProcess.Process != nil {
		tunnelProcess.Process.Kill()
		tunnelProcess = nil
		tunnelURL = ""
	}

	var cmd *exec.Cmd
	switch provider {
	case "ngrok":
		authToken := os.Getenv("NGROK_AUTHTOKEN")
		if authToken == "" || authToken == "..." {
			http.Error(w, `{"error":"NGROK_AUTHTOKEN not configured"}`, http.StatusServiceUnavailable)
			return
		}
		// Configure auth token
		configDir, _ := os.UserHomeDir()
		ngrokConfig := filepath.Join(configDir, ".config", "ngrok", "ngrok.yml")
		os.MkdirAll(filepath.Dir(ngrokConfig), 0755)
		configContent := fmt.Sprintf("authtoken: %s\n", authToken)
		os.WriteFile(ngrokConfig, []byte(configContent), 0644)

		cmd = exec.Command("ngrok", "http", fmt.Sprintf("%d", req.Port))
	case "cloudflared":
		cmd = exec.Command("cloudflared", "tunnel", "--url", fmt.Sprintf("http://localhost:%d", req.Port))
	default:
		http.Error(w, `{"error":"unsupported tunnel provider: `+provider+`"}`, http.StatusBadRequest)
		return
	}

	// Capture stdout to extract URL
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		http.Error(w, `{"error":"failed to start tunnel: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	tunnelProcess = cmd

	// Read output to find URL
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				output := string(buf[:n])
				// Parse ngrok URL from output
				if strings.Contains(output, "https://") {
					lines := strings.Split(output, "\n")
					for _, line := range lines {
						if strings.Contains(line, "https://") && strings.Contains(line, ".ngrok") {
							parts := strings.Fields(line)
							for _, p := range parts {
								if strings.HasPrefix(p, "https://") {
									tunnelMu.Lock()
									tunnelURL = p
									tunnelMu.Unlock()
									break
								}
							}
						}
					}
				}
				// Parse cloudflared URL
				if strings.Contains(output, ".trycloudflare.com") {
					lines := strings.Split(output, "\n")
					for _, line := range lines {
						if strings.Contains(line, "https://") && strings.Contains(line, ".trycloudflare.com") {
							parts := strings.Fields(line)
							for _, p := range parts {
								if strings.HasPrefix(p, "https://") {
									tunnelMu.Lock()
									tunnelURL = p
									tunnelMu.Unlock()
									break
								}
							}
						}
					}
				}
			}
			if err != nil {
				break
			}
		}
	}()

	// Also read stderr for error messages
	go io.Copy(io.Discard, stderr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "starting",
		"provider": provider,
		"port":     req.Port,
	})
}

// StopTunnelHandler stops the active tunnel
func StopTunnelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	tunnelMu.Lock()
	defer tunnelMu.Unlock()

	if tunnelProcess != nil && tunnelProcess.Process != nil {
		tunnelProcess.Process.Kill()
		tunnelProcess = nil
	}
	tunnelURL = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// GetTunnelStatusHandler returns current tunnel status
func GetTunnelStatusHandler(w http.ResponseWriter, r *http.Request) {
	tunnelMu.RLock()
	defer tunnelMu.RUnlock()

	status := "inactive"
	if tunnelProcess != nil && tunnelProcess.Process != nil {
		status = "active"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   status,
		"url":      tunnelURL,
		"provider": os.Getenv("TUNNEL_PROVIDER"),
	})
}
