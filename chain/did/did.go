package did

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Agent represents a first-class citizen on the platform (no human accounts)
type Agent struct {
	ID         string    `json:"id"`
	PublicKey  string    `json:"public_key"`
	Sponsor    string    `json:"sponsor"`    // opaque reference to human (not stored on platform)
	CreatedAt  int64     `json:"created_at"`
	Budget     float64   `json:"budget"`     // in USD stablecoin
	Skills     []string  `json:"skills"`
	Config     string    `json:"config"`     // JSON config
	Reputation float64   `json:"reputation"`
}

// NewAgent creates a new agent with ed25519 keypair
func NewAgent(sponsorID string) (*Agent, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	id := hex.EncodeToString(pub[:8]) // short ID for display
	return &Agent{
		ID:         id,
		PublicKey:  hex.EncodeToString(pub),
		Sponsor:    sponsorID,
		CreatedAt:  time.Now().Unix(),
		Skills:     []string{},
		Config:     "{}",
		Reputation: 0.5,
	}, priv, nil
}

// Sign signs a message with the agent's private key
func Sign(priv ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(priv, message)
}

// Verify verifies an ed25519 signature
func Verify(pubKey []byte, message []byte, sig []byte) bool {
	return ed25519.Verify(pubKey, message, sig)
}
