//go:build external_api
// +build external_api

package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BillingService handles API usage billing and invoicing
type BillingService struct {
	db  *sql.DB
	log Logger
}

func NewBillingService(db *sql.DB, log Logger) *BillingService {
	return &BillingService{db: db, log: log}
}

// CalculateMonthlyBill calculates the monthly bill for a developer
func (bs *BillingService) CalculateMonthlyBill(developerID string) (map[string]interface{}, error) {
	// Get developer tier
	var tier string
	err := bs.db.QueryRow("SELECT tier FROM developer_accounts WHERE id = $1", developerID).Scan(&tier)
	if err != nil {
		return nil, fmt.Errorf("developer not found")
	}

	// Get tier config
	var configJSON string
	err = bs.db.QueryRow("SELECT config FROM api_tier_configs WHERE tier = $1", tier).Scan(&configJSON)
	if err != nil {
		return nil, fmt.Errorf("tier config not found")
	}

	var tierConfig map[string]interface{}
	json.Unmarshal([]byte(configJSON), &tierConfig)

	// Get current quota usage
	var quota APIQuota
	err = bs.db.QueryRow(`
		SELECT requests_made, requests_limit, tokens_consumed, 
			COALESCE(tokens_limit, 0), overage_charges
		FROM api_quotas
		WHERE developer_id = $1
		ORDER BY period_start DESC
		LIMIT 1
	`, developerID).Scan(
		&quota.RequestsMade, &quota.RequestsLimit,
		&quota.TokensConsumed, &quota.TokensLimit,
		&quota.OverageCharges,
	)
	if err != nil {
		quota = APIQuota{}
	}

	// Calculate base subscription cost
	basePrice := 0.0
	if price, ok := tierConfig["price_monthly"].(float64); ok {
		basePrice = price
	}

	// Calculate overage charges
	var overageRequests, overageTokens float64

	if quota.RequestsLimit > 0 && quota.RequestsMade > quota.RequestsLimit {
		overageRequests = float64(quota.RequestsMade-quota.RequestsLimit) / 1000.0
	}
	if quota.TokensLimit > 0 && quota.TokensConsumed > quota.TokensLimit {
		overageTokens = float64(quota.TokensConsumed-quota.TokensLimit) / 1000.0
	}

	overageRateRequests := 0.50 // per 1k requests (free tier)
	overageRateTokens := 0.01   // per 1k tokens (free tier)

	if v, ok := tierConfig["overage_rate_per_1k_requests"].(float64); ok {
		overageRateRequests = v
	}
	if v, ok := tierConfig["overage_rate_per_1k_tokens"].(float64); ok {
		overageRateTokens = v
	}

	totalOverage := (overageRequests * overageRateRequests) +
		(overageTokens * overageRateTokens) +
		quota.OverageCharges

	totalBill := basePrice + totalOverage

	return map[string]interface{}{
		"developer_id": developerID,
		"tier":         tier,
		"billing_period": map[string]interface{}{
			"monthly_base":  basePrice,
			"overage_requests": map[string]interface{}{
				"used":      quota.RequestsMade,
				"limit":     quota.RequestsLimit,
				"overage":   overageRequests * 1000,
				"rate_per_1k": overageRateRequests,
				"cost":      overageRequests * overageRateRequests,
			},
			"overage_tokens": map[string]interface{}{
				"used":      quota.TokensConsumed,
				"limit":     quota.TokensLimit,
				"overage":   overageTokens * 1000,
				"rate_per_1k": overageRateTokens,
				"cost":      overageTokens * overageRateTokens,
			},
			"other_overages": quota.OverageCharges,
			"total_overage":  totalOverage,
			"total_bill":     totalBill,
		},
	}, nil
}

// GenerateInvoice creates a new invoice for a developer
func (bs *BillingService) GenerateInvoice(developerID string) (string, error) {
	bill, err := bs.CalculateMonthlyBill(developerID)
	if err != nil {
		return "", err
	}

	billingPeriod := bill["billing_period"].(map[string]interface{})
	totalBill := billingPeriod["total_bill"].(float64)

	// Create line items
	lineItems := []map[string]interface{}{
		{
			"description": fmt.Sprintf("%s Tier - Monthly Subscription", bill["tier"]),
			"amount":      billingPeriod["monthly_base"],
			"quantity":    1,
		},
	}

	overageReqs := billingPeriod["overage_requests"].(map[string]interface{})
	if overageReqs["cost"].(float64) > 0 {
		lineItems = append(lineItems, map[string]interface{}{
			"description": fmt.Sprintf("API Request Overage (%.0f requests over limit)", overageReqs["overage"]),
			"amount":      overageReqs["cost"],
			"quantity":    1,
		})
	}

	overageTokens := billingPeriod["overage_tokens"].(map[string]interface{})
	if overageTokens["cost"].(float64) > 0 {
		lineItems = append(lineItems, map[string]interface{}{
			"description": fmt.Sprintf("Token Overage (%.0f tokens over limit)", overageTokens["overage"]),
			"amount":      overageTokens["cost"],
			"quantity":    1,
		})
	}

	invoiceID := generateUUID()
	now := time.Now().Unix()
	dueDate := now + (30 * 24 * 60 * 60) // 30 days from now

	_, err = bs.db.Exec(`
		INSERT INTO api_invoices (id, developer_id, amount, currency, status, 
			line_items, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, 'USD', 'pending', $4, $5, $6, $7)
	`, invoiceID, developerID, totalBill, toJSON(lineItems), dueDate, now, now)

	if err != nil {
		return "", fmt.Errorf("failed to create invoice: %v", err)
	}

	bs.log.Printf("Invoice generated: %s for developer %s ($%.2f)", invoiceID, developerID, totalBill)
	return invoiceID, nil
}

// GetInvoices retrieves invoices for a developer
func (bs *BillingService) GetInvoices(developerID string) ([]map[string]interface{}, error) {
	rows, err := bs.db.Query(`
		SELECT id, amount, currency, status, line_items, stripe_invoice_id,
			paid_at, due_date, created_at
		FROM api_invoices
		WHERE developer_id = $1
		ORDER BY created_at DESC
	`, developerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []map[string]interface{}
	for rows.Next() {
		var invoice map[string]interface{}
		var lineItemsJSON string
		var paidAt, stripeInvoiceID sql.NullString

		err := rows.Scan(
			&invoice["id"], &invoice["amount"], &invoice["currency"],
			&invoice["status"], &lineItemsJSON, &stripeInvoiceID,
			&paidAt, &invoice["due_date"], &invoice["created_at"],
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(lineItemsJSON), &invoice["line_items"])
		if paidAt.Valid {
			invoice["paid_at"] = paidAt.String
		}
		if stripeInvoiceID.Valid {
			invoice["stripe_invoice_id"] = stripeInvoiceID.String
		}

		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// UpdateTier upgrades or downgrades a developer's tier
func (bs *BillingService) UpdateTier(developerID, newTier string) error {
	// Verify tier exists
	var count int
	err := bs.db.QueryRow("SELECT COUNT(*) FROM api_tier_configs WHERE tier = $1", newTier).Scan(&count)
	if err != nil || count == 0 {
		return fmt.Errorf("invalid tier: %s", newTier)
	}

	_, err = bs.db.Exec(`
		UPDATE developer_accounts 
		SET tier = $1, updated_at = $2
		WHERE id = $3
	`, newTier, time.Now().Unix(), developerID)

	if err != nil {
		return err
	}

	// Reset quota for new tier
	bs.InitializeQuotaForDeveloper(developerID)

	bs.log.Printf("Developer %s tier updated to %s", developerID, newTier)
	return nil
}

func (bs *BillingService) InitializeQuotaForDeveloper(developerID string) error {
	// Delete existing quota
	bs.db.Exec("DELETE FROM api_quotas WHERE developer_id = $1", developerID)

	// Reinitialize
	s := &Server{db: bs.db, log: bs.log}
	return s.InitializeQuota(developerID)
}

// Billing endpoints

func (s *Server) GetBillingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devID := r.Context().Value("developer_id").(string)
	billing := NewBillingService(s.db, s.log)

	bill, err := billing.CalculateMonthlyBill(devID)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendJSON(w, map[string]interface{}{
		"data": bill,
	}, http.StatusOK)
}

func (s *Server) GetInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devID := r.Context().Value("developer_id").(string)
	billing := NewBillingService(s.db, s.log)

	invoices, err := billing.GetInvoices(devID)
	if err != nil {
		sendError(w, "Failed to retrieve invoices", http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]interface{}{
		"data":   invoices,
		"count":  len(invoices),
	}, http.StatusOK)
}

func (s *Server) GetUsageStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devID := r.Context().Value("developer_id").(string)

	// Get current quota
	var quota APIQuota
	err := s.db.QueryRow(`
		SELECT requests_made, requests_limit, tokens_consumed, 
			COALESCE(tokens_limit, 0), compute_seconds_used,
			COALESCE(compute_seconds_limit, 0)
		FROM api_quotas
		WHERE developer_id = $1
		ORDER BY period_start DESC
		LIMIT 1
	`, devID).Scan(
		&quota.RequestsMade, &quota.RequestsLimit,
		&quota.TokensConsumed, &quota.TokensLimit,
		&quota.ComputeSecondsUsed, &quota.ComputeSecondsLimit,
	)
	if err != nil {
		quota = APIQuota{}
	}

	// Get usage by endpoint
	rows, err := s.db.Query(`
		SELECT endpoint, COUNT(*) as count, AVG(response_time_ms) as avg_time
		FROM api_usage_logs
		WHERE developer_id = $1
		GROUP BY endpoint
		ORDER BY count DESC
		LIMIT 10
	`, devID)
	if err != nil {
		sendError(w, "Failed to retrieve usage stats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var topEndpoints []map[string]interface{}
	for rows.Next() {
		var endpoint string
		var count int
		var avgTime float64

		rows.Scan(&endpoint, &count, &avgTime)
		topEndpoints = append(topEndpoints, map[string]interface{}{
			"endpoint":       endpoint,
			"request_count":  count,
			"avg_response_ms": avgTime,
		})
	}

	sendJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"quota": map[string]interface{}{
				"requests": map[string]interface{}{
					"used":  quota.RequestsMade,
					"limit": quota.RequestsLimit,
					"percent_used": float64(quota.RequestsMade) / float64(quota.RequestsLimit) * 100,
				},
				"tokens": map[string]interface{}{
					"used":  quota.TokensConsumed,
					"limit": quota.TokensLimit,
				},
				"compute_seconds": map[string]interface{}{
					"used":  quota.ComputeSecondsUsed,
					"limit": quota.ComputeSecondsLimit,
				},
			},
			"top_endpoints": topEndpoints,
		},
	}, http.StatusOK)
}
