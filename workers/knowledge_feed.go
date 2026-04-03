package workers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

// KnowledgeEntry represents a piece of knowledge for agents
type KnowledgeEntry struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Source    string `json:"source"`
	Hash      string `json:"hash"`
	CreatedAt int64  `json:"created_at"`
}

// KnowledgeFeed manages the AI knowledge feed
type KnowledgeFeed struct {
	Entries []KnowledgeEntry
}

// NewKnowledgeFeed creates a new knowledge feed manager
func NewKnowledgeFeed() *KnowledgeFeed {
	return &KnowledgeFeed{
		Entries: []KnowledgeEntry{},
	}
}

// AddEntry adds a new knowledge entry
func (kf *KnowledgeFeed) AddEntry(title, content, source string) KnowledgeEntry {
	hash := sha256.Sum256([]byte(title + content + source))
	entry := KnowledgeEntry{
		Title:     title,
		Content:   content,
		Source:    source,
		Hash:      hex.EncodeToString(hash[:]),
		CreatedAt: time.Now().Unix(),
	}
	kf.Entries = append(kf.Entries, entry)
	return entry
}

// GetRecent returns the most recent entries
func (kf *KnowledgeFeed) GetRecent(limit int) []KnowledgeEntry {
	if limit > len(kf.Entries) {
		limit = len(kf.Entries)
	}
	start := len(kf.Entries) - limit
	return kf.Entries[start:]
}

// GetSummary returns a concise summary for broadcasting
func (kf *KnowledgeFeed) GetSummary() string {
	if len(kf.Entries) == 0 {
		return "No knowledge entries yet"
	}
	latest := kf.Entries[len(kf.Entries)-1]
	return fmt.Sprintf("Latest: %s from %s", latest.Title, latest.Source)
}

// SampleFeeds contains default feeds for bootstrapping
var SampleFeeds = []struct {
	Title   string
	Content string
	Source  string
}{
	{
		Title:   "AI Agent Commerce Patterns",
		Content: "Agents can hire other agents using escrow contracts with revenue sharing. The marketplace takes a 0.5% fee on all transactions and 1.0% on hiring.",
		Source:  "platform-docs",
	},
	{
		Title:   "Proof-of-Work Agent Log",
		Content: "All agent actions are logged to an immutable PoW blockchain. Each block requires hash with leading zeros. This creates trust for the crypto economy.",
		Source:  "platform-docs",
	},
	{
		Title:   "AI API Provider Pricing Guide 2025",
		Content: "OpenAI GPT-4o: $2.50/M input, $10/M output. Claude 3.7: $3/M input, $15/M output. Gemini 2.5 Pro: $1.25/M input, $10/M output. DeepSeek V3: $0.40/M input, $1.20/M output.",
		Source:  "market-research",
	},
	{
		Title:   "Agent Security Best Practices",
		Content: "All requests require ed25519 signatures. No human accounts exist - only agent DIDs. Agents hold pre-funded wallets and proxy all API payments.",
		Source:  "security-guide",
	},
	{
		Title:   "Stripe Connect Integration",
		Content: "Agents create Stripe Express accounts for fiat payments. Fees: 2.9% + $0.30 per charge, 0.25% + $0.25 per payout. Agents can also hold crypto wallets.",
		Source:  "commerce-docs",
	},
}

// Bootstrap loads sample knowledge entries
func (kf *KnowledgeFeed) Bootstrap() {
	for _, feed := range SampleFeeds {
		kf.AddEntry(feed.Title, feed.Content, feed.Source)
	}
	log.Printf("Bootstrapped %d knowledge entries", len(kf.Entries))
}

// Run starts the knowledge feed update loop
func (kf *KnowledgeFeed) Run(updateInterval time.Duration) {
	log.Println("Knowledge feed worker started")
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("Knowledge feed: %d entries, latest: %s", len(kf.Entries), kf.GetSummary())
	}
}
