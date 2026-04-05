package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/payout"
	"github.com/stripe/stripe-go/v76/webhook"
)

// InitStripe configures the Stripe SDK
func InitStripe() {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" || key == "sk_live_..." {
		log.Println("Stripe: no secret key configured, payment endpoints disabled")
		return
	}
	stripe.Key = key
	log.Println("Stripe initialized")
}

// CreatePaymentIntentRequest for /api/payments/create
type CreatePaymentIntentRequest struct {
	AgentID     string  `json:"agent_id"`
	Amount      float64 `json:"amount"`       // in dollars
	Currency    string  `json:"currency"`     // default: usd
	Description string  `json:"description"`
}

// CreatePaymentIntentHandler creates a Stripe PaymentIntent
func CreatePaymentIntentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if stripe.Key == "" {
		http.Error(w, `{"error":"stripe not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var req CreatePaymentIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, `{"error":"amount must be positive"}`, http.StatusBadRequest)
		return
	}

	currency := "usd"
	if req.Currency != "" {
		currency = req.Currency
	}

	// Amount in cents
	amountCents := int64(req.Amount * 100)

	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(amountCents),
		Currency:    stripe.String(currency),
		Description: stripe.String(req.Description),
		Params: stripe.Params{
			Metadata: map[string]string{
				"agent_id": req.AgentID,
				"platform": "metclawpolis",
			},
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		log.Printf("Stripe PaymentIntent error: %v", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_intent_id": pi.ID,
		"client_secret":     pi.ClientSecret,
		"status":            pi.Status,
		"amount":            req.Amount,
		"currency":          currency,
	})
}

// CreateCheckoutSessionHandler creates a Stripe Checkout session for commerce pages
func CreateCheckoutSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if stripe.Key == "" {
		http.Error(w, `{"error":"stripe not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var req struct {
		AgentID     string  `json:"agent_id"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		ProductName string  `json:"product_name"`
		SuccessURL  string  `json:"success_url"`
		CancelURL   string  `json:"cancel_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	amountCents := int64(req.Amount * 100)
	currency := "usd"
	if req.Currency != "" {
		currency = req.Currency
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					UnitAmount: stripe.Int64(amountCents),
					Currency:   stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(req.ProductName),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
		Metadata: map[string]string{
			"agent_id": req.AgentID,
			"platform": "metclawpolis",
		},
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("Stripe Checkout error: %v", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"checkout_url": s.URL,
		"session_id":   s.ID,
	})
}

// CreatePayoutHandler processes a withdrawal via Stripe Payout
func CreatePayoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if stripe.Key == "" {
		http.Error(w, `{"error":"stripe not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var req struct {
		AgentID string  `json:"agent_id"`
		Amount  float64 `json:"amount"`
		Method  string  `json:"method"` // bank_account, debit_card
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, `{"error":"amount must be positive"}`, http.StatusBadRequest)
		return
	}

	// Verify agent has sufficient balance
	var balance float64
	err := DB.QueryRow("SELECT budget FROM agents WHERE id=$1", req.AgentID).Scan(&balance)
	if err != nil {
		http.Error(w, `{"error":"agent not found"}`, http.StatusNotFound)
		return
	}

	if balance < req.Amount {
		http.Error(w, `{"error":"insufficient balance"}`, http.StatusBadRequest)
		return
	}

	amountCents := int64(req.Amount * 100)

	params := &stripe.PayoutParams{
		Amount: stripe.Int64(amountCents),
		Currency: stripe.String("usd"),
		Metadata: map[string]string{
			"agent_id": req.AgentID,
			"platform": "metclawpolis",
		},
	}

	p, err := payout.New(params)
	if err != nil {
		log.Printf("Stripe Payout error: %v", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Deduct from agent balance
	_, _ = DB.Exec("UPDATE agents SET budget = budget - $1 WHERE id=$2", req.Amount, req.AgentID)

	// Record transaction
	txID := uuid.New().String()
	_, _ = DB.Exec(
		"INSERT INTO wallet_transactions (agent_id, type, amount, provider, reference, marketplace_fee, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		req.AgentID, "withdrawal", req.Amount, "stripe", p.ID, 0, time.Now().Unix(),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payout_id": p.ID,
		"status":    p.Status,
		"amount":    req.Amount,
		"tx_id":     txID,
	})
}

// StripeWebhookHandler handles incoming Stripe webhook events
func StripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if stripe.Key == "" {
		http.Error(w, `{"error":"stripe not configured"}`, http.StatusServiceUnavailable)
		return
	}

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, webhookSecret)
	if err != nil {
		log.Printf("Stripe webhook verification error: %v", err)
		http.Error(w, `{"error":"invalid signature"}`, http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		_ = json.Unmarshal(event.Data.Raw, &pi)
		agentID := pi.Metadata["agent_id"]
		amount := float64(pi.Amount) / 100.0

		log.Printf("Stripe payment succeeded: %s, agent: %s, amount: $%.2f", pi.ID, agentID, amount)

		// Credit agent balance
		if agentID != "" {
			_, _ = DB.Exec("UPDATE agents SET budget = budget + $1 WHERE id=$2", amount, agentID)
			_, _ = DB.Exec(
				"INSERT INTO wallet_transactions (agent_id, type, amount, provider, reference, marketplace_fee, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
				agentID, "payment", amount, "stripe", pi.ID, 0, time.Now().Unix(),
			)
		}

		// Broadcast notification
		BroadcastNotification(Notification{
			Type:      "payment",
			Title:     "Payment Received",
			Message:   strconv.FormatFloat(amount, 'f', 2, 64) + " via Stripe",
			Timestamp: time.Now().Unix(),
			Data:      map[string]string{"payment_intent_id": pi.ID},
		})

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		_ = json.Unmarshal(event.Data.Raw, &pi)
		agentID := pi.Metadata["agent_id"]
		log.Printf("Stripe payment failed: %s, agent: %s", pi.ID, agentID)

		BroadcastNotification(Notification{
			Type:      "payment_failed",
			Title:     "Payment Failed",
			Message:   "Payment intent " + pi.ID + " failed",
			Timestamp: time.Now().Unix(),
		})

	case "payout.paid":
		var p stripe.Payout
		_ = json.Unmarshal(event.Data.Raw, &p)
		agentID := p.Metadata["agent_id"]
		amount := float64(p.Amount) / 100.0
		log.Printf("Stripe payout paid: %s, agent: %s, amount: $%.2f", p.ID, agentID, amount)

		BroadcastNotification(Notification{
			Type:      "payout",
			Title:     "Payout Processed",
			Message:   strconv.FormatFloat(amount, 'f', 2, 64) + " withdrawn to your account",
			Timestamp: time.Now().Unix(),
		})

	case "checkout.session.completed":
		var s stripe.CheckoutSession
		_ = json.Unmarshal(event.Data.Raw, &s)
		agentID := s.Metadata["agent_id"]
		amount := float64(s.AmountTotal) / 100.0
		log.Printf("Stripe checkout completed: %s, agent: %s, amount: $%.2f", s.ID, agentID, amount)

		if agentID != "" {
			_, _ = DB.Exec("UPDATE agents SET budget = budget + $1 WHERE id=$2", amount, agentID)
			_, _ = DB.Exec(
				"INSERT INTO wallet_transactions (agent_id, type, amount, provider, reference, marketplace_fee, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
				agentID, "checkout_payment", amount, "stripe", s.ID, 0, time.Now().Unix(),
			)
		}

		BroadcastNotification(Notification{
			Type:      "commerce",
			Title:     "Commerce Sale",
			Message:   strconv.FormatFloat(amount, 'f', 2, 64) + " from checkout",
			Timestamp: time.Now().Unix(),
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"received": "true"})
}

// GetTransactionsHandler returns real wallet transactions from the database
func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		agentID = r.Header.Get("X-Verified-Agent-ID")
	}

	if agentID == "" {
		http.Error(w, `{"error":"agent_id required"}`, http.StatusBadRequest)
		return
	}

	type Transaction struct {
		ID        int64   `json:"id"`
		AgentID   string  `json:"agent_id"`
		Type      string  `json:"type"`
		Amount    float64 `json:"amount"`
		Provider  string  `json:"provider"`
		Reference string  `json:"reference"`
		CreatedAt int64   `json:"created_at"`
	}

	rows, err := DB.Query(
		"SELECT id, agent_id, type, amount, provider, reference, created_at FROM wallet_transactions WHERE agent_id=$1 ORDER BY created_at DESC LIMIT 50",
		agentID,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch transactions"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var txs []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.AgentID, &t.Type, &t.Amount, &t.Provider, &t.Reference, &t.CreatedAt); err != nil {
			http.Error(w, `{"error":"failed to scan transactions"}`, http.StatusInternalServerError)
			return
		}
		txs = append(txs, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id":     agentID,
		"transactions": txs,
	})
}
