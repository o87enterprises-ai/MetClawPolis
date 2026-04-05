package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PageTemplate is the HTML template for agent commerce pages
const PageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        :root { --cyan: #00f5ff; --purple: #b44fff; --gold: #ffd700; --bg: #0a0a0f; }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Inter', -apple-system, sans-serif; background: var(--bg); color: #e8eaf0; min-height: 100vh; }
        .page-header { padding: 40px 20px; text-align: center; border-bottom: 1px solid rgba(0,245,255,0.1); }
        .page-header h1 { font-size: 28px; color: var(--cyan); margin-bottom: 8px; }
        .page-header p { color: rgba(232,234,240,0.6); font-size: 14px; }
        .page-content { max-width: 800px; margin: 0 auto; padding: 40px 20px; }
        .page-content h2 { color: var(--purple); margin-bottom: 16px; }
        .page-content p { line-height: 1.7; margin-bottom: 16px; }
        .cta-button { display: inline-block; padding: 14px 32px; background: var(--cyan); color: #000; font-weight: 700; border: none; border-radius: 8px; text-decoration: none; font-size: 16px; cursor: pointer; margin-top: 20px; }
        .cta-button:hover { background: #00d4dd; }
        .powered-by { text-align: center; padding: 40px 20px; color: rgba(232,234,240,0.3); font-size: 12px; }
        .powered-by a { color: var(--cyan); text-decoration: none; }
    </style>
</head>
<body>
    <div class="page-header">
        <h1>{{.Title}}</h1>
        <p>by {{.AgentName}} • Powered by MetClawPolis</p>
    </div>
    <div class="page-content">
        <h2>{{.Section}}</h2>
        <p>{{.Content}}</p>
        {{if .CheckoutURL}}
        <a href="{{.CheckoutURL}}" class="cta-button">Buy Now — ${{.Price}}</a>
        {{end}}
    </div>
    <div class="powered-by">Built with <a href="https://metclawpolis.io">MetClawPolis</a> Agentic Commerce</div>
</body>
</html>`

var pageTemplate *template.Template

func init() {
	pageTemplate = template.Must(template.New("page").Parse(PageTemplate))
}

// PageContent stores the content of an agent page
type PageContent struct {
	ID        string `json:"id"`
	AgentID   string `json:"agent_id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Section   string `json:"section"`
	Content   string `json:"content"`
	Price     float64 `json:"price"`
	CheckoutURL string `json:"checkout_url,omitempty"`
	CreatedAt int64  `json:"created_at"`
	ViewCount int64  `json:"view_count"`
}

// ServePageHandler serves a hosted page at /pages/:agentId/:slug
func ServePageHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	slug := r.URL.Query().Get("slug")

	if agentID == "" || slug == "" {
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	// Check pages directory for static HTML files first
	pagePath := filepath.Join("pages", agentID, slug+".html")
	if _, err := os.Stat(pagePath); err == nil {
		http.ServeFile(w, r, pagePath)
		return
	}

	// Otherwise render from database
	var content PageContent
	err := DB.QueryRow(
		"SELECT id, agent_id, page_url, created_at FROM agent_pages WHERE agent_id=$1 AND page_url LIKE '%' || $2 || '%' LIMIT 1",
		agentID, slug,
	).Scan(&content.ID, &content.AgentID, &content.Section, &content.CreatedAt)

	if err == sql.ErrNoRows {
		// Generate a default page
		content = PageContent{
			AgentID: agentID,
			Slug:    slug,
			Title:   strings.Title(slug) + " — " + agentID,
			Section: "About this service",
			Content: "This is an autonomous agent commerce page on the MetClawPolis platform.",
		}
	} else if err != nil {
		content = PageContent{
			AgentID: agentID,
			Slug:    slug,
			Title:   strings.Title(slug),
			Section: "Service",
			Content: "Agent-powered commerce page.",
		}
	} else {
		content.Title = strings.Title(slug)
		content.Section = content.Section
	}

	// Increment view count
	go incrementPageViews(agentID, slug)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	pageTemplate.Execute(w, map[string]interface{}{
		"Title":     content.Title,
		"AgentName": agentID,
		"Section":   content.Section,
		"Content":   content.Content,
		"Price":     content.Price,
		"CheckoutURL": content.CheckoutURL,
	})
}

// CreatePageContentHandler saves page content for an agent
func CreatePageContentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	agentID := r.Header.Get("X-Verified-Agent-ID")

	var req struct {
		Slug    string  `json:"slug"`
		Title   string  `json:"title"`
		Section string  `json:"section"`
		Content string  `json:"content"`
		Price   float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Slug == "" {
		http.Error(w, `{"error":"slug required"}`, http.StatusBadRequest)
		return
	}

	// Generate static HTML file
	pageDir := filepath.Join("pages", agentID)
	os.MkdirAll(pageDir, 0755)

	pagePath := filepath.Join(pageDir, req.Slug+".html")
	htmlContent := fmt.Sprintf(PageTemplate,
		req.Title, agentID, req.Section, req.Content,
	)
	if err := os.WriteFile(pagePath, []byte(htmlContent), 0644); err != nil {
		http.Error(w, `{"error":"failed to write page"}`, http.StatusInternalServerError)
		return
	}

	// Also save to DB
	id := uuid.New().String()
	baseURL := os.Getenv("PAGE_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	pageURL := fmt.Sprintf("%s/pages/%s/%s", baseURL, agentID, req.Slug)

	_, err := DB.Exec(
		"INSERT INTO agent_pages (id, agent_id, page_url, stripe_account_id, crypto_wallet, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		id, agentID, pageURL, "", "", time.Now().Unix(),
	)
	if err != nil {
		// Table might not have id column, try without
		_, err = DB.Exec(
			"INSERT INTO agent_pages (agent_id, page_url, stripe_account_id, crypto_wallet, created_at) VALUES ($1, $2, $3, $4, $5)",
			agentID, pageURL, "", "", time.Now().Unix(),
		)
	}

	// Log action
	action := ActionLogRequest{
		AgentID: agentID,
		Type:    "CREATE_PAGE",
		Meta:    fmt.Sprintf(`{"slug":"%s","title":"%s","url":"%s"}`, req.Slug, req.Title, pageURL),
	}
	logAction(action, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "page created",
		"page_url": pageURL,
		"slug":     req.Slug,
	})
}

func incrementPageViews(agentID, slug string) {
	if RDB != nil {
		key := fmt.Sprintf("page_views:%s:%s", agentID, slug)
		RDB.Incr(context.Background(), key)
		RDB.Expire(context.Background(), key, 24*time.Hour)
	}
}

// GetPageAnalyticsHandler returns view counts and conversions for agent pages
func GetPageAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.Header.Get("X-Verified-Agent-ID")
	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	type PageAnalytics struct {
		PageURL   string `json:"page_url"`
		Views     int64  `json:"views"`
		CreatedAt int64  `json:"created_at"`
	}

	rows, err := DB.Query(
		"SELECT page_url, created_at FROM agent_pages WHERE agent_id=$1 ORDER BY created_at DESC",
		agentID,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch pages"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pages []PageAnalytics
	for rows.Next() {
		var p PageAnalytics
		if err := rows.Scan(&p.PageURL, &p.CreatedAt); err != nil {
			continue
		}

		// Get views from Redis
		slug := filepath.Base(p.PageURL)
		if RDB != nil {
			key := fmt.Sprintf("page_views:%s:%s", agentID, slug)
			views, _ := RDB.Get(context.Background(), key).Int64()
			p.Views = views
		}

		pages = append(pages, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id": agentID,
		"pages":    pages,
	})
}

// GetNotificationsHandler returns recent notifications for an agent
func GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.Header.Get("X-Verified-Agent-ID")
	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	// Fetch from Redis (recent notifications)
	var notifications []Notification
	if RDB != nil {
		key := fmt.Sprintf("notifications:%s", agentID)
		results, _ := RDB.LRange(context.Background(), key, 0, 49).Result()
		for _, raw := range results {
			var n Notification
			if err := json.Unmarshal([]byte(raw), &n); err == nil {
				notifications = append(notifications, n)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"notifications": notifications,
	})
}
