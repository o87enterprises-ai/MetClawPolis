package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// Action represents an agent action to be logged on the PoW chain
type Action struct {
	AgentID    string `json:"agent_id"`
	Type       string `json:"type"`        // e.g., "CREATE_PAGE", "TRANSFER", "HIRE"
	InputHash  string `json:"input_hash"`
	OutputHash string `json:"output_hash"`
	Timestamp  int64  `json:"timestamp"`
	Meta       string `json:"meta"`        // additional JSON metadata
}

// Block represents a single block in the agent action chain
type Block struct {
	Index        int64  `json:"index"`
	PreviousHash string `json:"previous_hash"`
	Action       Action `json:"action"`
	Nonce        int64  `json:"nonce"`
	Hash         string `json:"hash"`
}

// ComputeHash computes SHA256 hash of the block content
func (b *Block) ComputeHash() string {
	record := strconv.FormatInt(b.Index, 10) + b.PreviousHash +
		b.Action.AgentID + b.Action.Type + b.Action.InputHash +
		b.Action.OutputHash + strconv.FormatInt(b.Action.Timestamp, 10) +
		strconv.FormatInt(b.Nonce, 10) + b.Action.Meta
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// Mine performs proof-of-work by finding a hash with the required difficulty
func (b *Block) Mine(difficulty int) {
	target := strings.Repeat("0", difficulty)
	for !strings.HasPrefix(b.Hash, target) {
		b.Nonce++
		b.Hash = b.ComputeHash()
	}
}

// Blockchain holds the chain of agent actions
type Blockchain struct {
	Chain      []Block `json:"chain"`
	Difficulty int     `json:"difficulty"`
}

// NewBlockchain creates a new blockchain with genesis block
func NewBlockchain(difficulty int) *Blockchain {
	genesis := Block{
		Index:        0,
		PreviousHash: "0",
		Hash:         "",
		Action: Action{
			AgentID:   "system",
			Type:      "GENESIS",
			Timestamp: 0,
		},
	}
	genesis.Mine(difficulty)
	return &Blockchain{
		Chain:      []Block{genesis},
		Difficulty: difficulty,
	}
}

// AddAction adds a new action to the chain with proof-of-work
func (bc *Blockchain) AddAction(action Action) Block {
	prevBlock := bc.Chain[len(bc.Chain)-1]
	newBlock := Block{
		Index:        prevBlock.Index + 1,
		PreviousHash: prevBlock.Hash,
		Action:       action,
	}
	newBlock.Mine(bc.Difficulty)
	bc.Chain = append(bc.Chain, newBlock)
	return newBlock
}

// Verify checks the integrity of the entire chain
func (bc *Blockchain) Verify() bool {
	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		previousBlock := bc.Chain[i-1]

		if currentBlock.PreviousHash != previousBlock.Hash {
			return false
		}

		target := strings.Repeat("0", bc.Difficulty)
		if !strings.HasPrefix(currentBlock.Hash, target) {
			return false
		}

		if currentBlock.ComputeHash() != currentBlock.Hash {
			return false
		}
	}
	return true
}

// GetLatestBlock returns the most recent block
func (bc *Blockchain) GetLatestBlock() Block {
	return bc.Chain[len(bc.Chain)-1]
}

// GetChainJSON returns the chain as JSON-serializable slice
func (bc *Blockchain) GetChainJSON() []Block {
	return bc.Chain
}

// CalculateInputHash creates a deterministic hash from input data
func CalculateInputHash(data ...string) string {
	combined := strings.Join(data, "|")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// FormatBalance formats a float as a balance string
func FormatBalance(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}
