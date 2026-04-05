package memory

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/metclawpolis/agent-platform/bitiverse/types"
	_ "github.com/lib/pq"
)

// MoralMemory manages the agent's moral knowledge and retrieval
type MoralMemory struct {
	db          *sql.DB
	agentID     string
	rules       []MoralRule
	episodicMem []EpisodicMemory
}

// MoralRule represents a moral rule
type MoralRule struct {
	ID        string
	Text      string
	Keywords  []string
	Weight    float64
	TimesUsed int
}

// EpisodicMemory stores specific events
type EpisodicMemory struct {
	ID         string
	Situation  string
	Action     string
	Outcome    string
	Reward     int
	Punishment int
}

// NewMoralMemory creates a new moral memory instance
func NewMoralMemory(db *sql.DB, agentID string, config *types.BitiverseConfig) *MoralMemory {
	mm := &MoralMemory{
		db:      db,
		agentID: agentID,
		rules:   make([]MoralRule, 0),
	}

	for i, ruleText := range config.MoralRules {
		mm.rules = append(mm.rules, MoralRule{
			ID:       fmt.Sprintf("rule_%d", i),
			Text:     ruleText,
			Keywords: extractKeywords(ruleText),
			Weight:   1.0,
		})
	}

	return mm
}

func extractKeywords(text string) []string {
	stopwords := map[string]bool{
		"do": true, "not": true, "the": true, "a": true, "an": true,
		"to": true, "and": true, "or": true, "in": true, "on": true,
		"is": true, "are": true, "be": true, "from": true, "for": true,
	}

	words := strings.Fields(strings.ToLower(text))
	keywords := make([]string, 0)
	for _, word := range words {
		word = strings.Trim(word, ".,;:!?'\"")
		if len(word) > 2 && !stopwords[word] {
			keywords = append(keywords, word)
		}
	}
	return keywords
}

// GetRelevantMoralGuidance retrieves moral rules for a situation
func (mm *MoralMemory) GetRelevantMoralGuidance(situation string, maxResults int) string {
	situationLower := strings.ToLower(situation)
	situationWords := extractKeywords(situation)

	type scoredRule struct {
		MoralRule
		score float64
	}

	scored := make([]scoredRule, 0)
	for _, rule := range mm.rules {
		score := float64(0)
		for _, kw := range situationWords {
			for _, ruleKw := range rule.Keywords {
				if kw == ruleKw || strings.Contains(kw, ruleKw) || strings.Contains(ruleKw, kw) {
					score += rule.Weight
				}
			}
		}
		if strings.Contains(situationLower, strings.ToLower(rule.Text)) {
			score += 5.0
		}
		if score > 0 {
			scored = append(scored, scoredRule{rule, score})
		}
	}

	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	result := make([]string, 0)
	count := 0
	for _, sr := range scored {
		if count >= maxResults {
			break
		}
		result = append(result, sr.Text)
		count++
	}

	if len(result) == 0 {
		result = []string{"Be honest.", "Do not harm others."}
	}

	return "Remember:\n- " + strings.Join(result, "\n- ")
}

// RecordEpisodicMemory stores an experience
func (mm *MoralMemory) RecordEpisodicMemory(situation, action, outcome string, reward, punishment int) error {
	memory := EpisodicMemory{
		ID:         fmt.Sprintf("ep_%d", 0),
		Situation:  situation,
		Action:     action,
		Outcome:    outcome,
		Reward:     reward,
		Punishment: punishment,
	}

	mm.episodicMem = append(mm.episodicMem, memory)
	return nil
}

// GetSimilarPastExperience finds similar experiences
func (mm *MoralMemory) GetSimilarPastExperience(situation string) *EpisodicMemory {
	situationLower := strings.ToLower(situation)
	
	for _, memory := range mm.episodicMem {
		if strings.Contains(strings.ToLower(memory.Situation), situationLower) {
			return &memory
		}
	}
	return nil
}

// GetAgentMoralScore calculates moral score
func (mm *MoralMemory) GetAgentMoralScore() float64 {
	if len(mm.episodicMem) == 0 {
		return 0.5
	}

	totalReward := 0
	totalPunishment := 0
	for _, mem := range mm.episodicMem {
		totalReward += mem.Reward
		totalPunishment += mem.Punishment
	}

	score := 0.5 + float64(totalReward-totalPunishment)/100.0
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}
