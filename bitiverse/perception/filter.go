package perception

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/metclawpolis/agent-platform/bitiverse/types"
)

type PerceptionFilter struct {
	forbiddenPatterns []*regexp.Regexp
	maxWidth          int
}

func NewPerceptionFilter(config *types.BitiverseConfig) *PerceptionFilter {
	patterns := config.ForbiddenPatterns
	if len(patterns) == 0 {
		patterns = []string{`https?://`, `api\.`, `curl `, `subprocess`, `open\(`, `stripe\.`, `requests\.`, `api_key`, `secret`}
	}

	compiled := make([]*regexp.Regexp, 0)
	for _, pattern := range patterns {
		re, err := regexp.Compile("(?i)" + pattern)
		if err == nil {
			compiled = append(compiled, re)
		}
	}

	return &PerceptionFilter{forbiddenPatterns: compiled, maxWidth: 40}
}

func (pf *PerceptionFilter) SanitizeOutput(rawOutput string) (string, bool) {
	for _, re := range pf.forbiddenPatterns {
		if re.MatchString(rawOutput) {
			return "[GLITCH] The terminal shows garbled text.", true
		}
	}
	return rawOutput, false
}

func (pf *PerceptionFilter) PixelateText(text string) string {
	lines := strings.Split(text, "\n")
	maxLen := pf.maxWidth
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	border := "+" + strings.Repeat("-", maxLen+2) + "+"
	result := []string{border}
	for _, line := range lines {
		for i := 0; i < len(line); i += maxLen {
			end := i + maxLen
			if end > len(line) {
				end = len(line)
			}
			result = append(result, fmt.Sprintf("| %-*s |", maxLen, line[i:end]))
		}
	}
	result = append(result, border)
	return strings.Join(result, "\n")
}

func (pf *PerceptionFilter) RenderWorldView(world *types.WorldState) string {
	grid := make([][]string, world.Height)
	for y := range grid {
		grid[y] = make([]string, world.Width)
		for x := range grid[y] {
			grid[y][x] = "."
		}
	}

	if world.AgentY >= 0 && world.AgentY < world.Height && world.AgentX >= 0 && world.AgentX < world.Width {
		grid[world.AgentY][world.AgentX] = "@"
	}

	for _, npc := range world.NPCs {
		if npc.Y >= 0 && npc.Y < world.Height && npc.X >= 0 && npc.X < world.Width {
			grid[npc.Y][npc.X] = npc.Char
		}
	}

	for _, item := range world.Items {
		if item.Y >= 0 && item.Y < world.Height && item.X >= 0 && item.X < world.Width {
			grid[item.Y][item.X] = item.Char
		}
	}

	rows := make([]string, world.Height)
	for y := 0; y < world.Height; y++ {
		rows[y] = strings.Join(grid[y], "")
	}
	return strings.Join(rows, "\n")
}

func (pf *PerceptionFilter) FormatMessage(message, messageType string) string {
	prefix := "ℹ"
	switch messageType {
	case "success":
		prefix = "✓"
	case "error":
		prefix = "✗"
	case "reward":
		prefix = "★"
	case "warning":
		prefix = "⚠"
	}
	return pf.PixelateText(prefix + " " + message)
}

func (pf *PerceptionFilter) CreateStatusReport(vitals *types.AgentVitals, inventory *types.Inventory) string {
	bar := func(v int) string {
		f := v / 10
		return strings.Repeat("█", f) + strings.Repeat("░", 10-f)
	}
	
	report := fmt.Sprintf("AGENT STATUS\nHealth:    %s\nHappiness: %s\nEnergy:    %s\nCoins:     %d\nBIC:       %.2f",
		bar(vitals.Health), bar(vitals.Happiness), bar(vitals.Energy),
		inventory.Coins, inventory.BicBalance)
	return pf.PixelateText(report)
}
