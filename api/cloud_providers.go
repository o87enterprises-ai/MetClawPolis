package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// CloudProvider represents a cloud infrastructure provider
type CloudProvider struct {
	Name        string       `json:"name"`
	Type        string       `json:"type"` // compute, storage, gpu
	BaseURL     string       `json:"base_url"`
	Plans       []CloudPlan  `json:"plans"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
}

// CloudPlan represents a VM/compute plan
type CloudPlan struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	CPU         string  `json:"cpu"`
	Memory      string  `json:"memory"`
	Storage     string  `json:"storage"`
	Bandwidth   string  `json:"bandwidth"`
	PriceMonthly float64 `json:"price_monthly"`
	PriceHourly  float64 `json:"price_hourly"`
	Type        string  `json:"plan_type"` // on_demand, flat_rate
}

// CloudPurchaseRequest represents a request to purchase cloud resources
type CloudPurchaseRequest struct {
	AgentID     string  `json:"agent_id"`
	Provider    string  `json:"provider"`
	PlanID      string  `json:"plan_id"`
	BillingType string  `json:"billing_type"` // on_demand, flat_rate
}

// CloudPurchaseResponse represents the response after a cloud purchase
type CloudPurchaseResponse struct {
	PurchaseID      string  `json:"purchase_id"`
	Provider        string  `json:"provider"`
	Plan            string  `json:"plan"`
	TotalCost       float64 `json:"total_cost"`
	NetworkFee      float64 `json:"network_fee"`
	AgentBalance    float64 `json:"agent_balance"`
	Status          string  `json:"status"`
	Message         string  `json:"message"`
}

// CloudProviders map - available cloud infrastructure providers
var CloudProviders = map[string]CloudProvider{
	"digitalocean": {
		Name:        "DigitalOcean",
		Type:        "compute",
		BaseURL:     "https://api.digitalocean.com/v2",
		Description: "Simple, scalable cloud computing for developers",
		Icon:        "🔵",
		Plans: []CloudPlan{
			{ID: "do-basic", Name: "Basic", CPU: "1 vCPU", Memory: "1 GB", Storage: "25 GB SSD", Bandwidth: "1 TB", PriceMonthly: 6.00, PriceHourly: 0.009, Type: "on_demand"},
			{ID: "do-pro", Name: "Pro", CPU: "2 vCPU", Memory: "2 GB", Storage: "50 GB SSD", Bandwidth: "3 TB", PriceMonthly: 18.00, PriceHourly: 0.027, Type: "on_demand"},
			{ID: "do-elite", Name: "Elite", CPU: "4 vCPU", Memory: "8 GB", Storage: "160 GB SSD", Bandwidth: "5 TB", PriceMonthly: 48.00, PriceHourly: 0.071, Type: "on_demand"},
		},
	},
	"aws": {
		Name:        "AWS",
		Type:        "compute",
		BaseURL:     "https://ec2.amazonaws.com",
		Description: "Amazon Web Services — industry-leading cloud platform",
		Icon:        "🟠",
		Plans: []CloudPlan{
			{ID: "aws-t3micro", Name: "t3.micro", CPU: "2 vCPU", Memory: "1 GB", Storage: "EBS only", Bandwidth: "Up to 5 Gbps", PriceMonthly: 8.50, PriceHourly: 0.012, Type: "on_demand"},
			{ID: "aws-t3small", Name: "t3.small", CPU: "2 vCPU", Memory: "2 GB", Storage: "EBS only", Bandwidth: "Up to 5 Gbps", PriceMonthly: 17.00, PriceHourly: 0.024, Type: "on_demand"},
			{ID: "aws-m5large", Name: "m5.large", CPU: "2 vCPU", Memory: "8 GB", Storage: "EBS only", Bandwidth: "Up to 10 Gbps", PriceMonthly: 70.08, PriceHourly: 0.096, Type: "on_demand"},
		},
	},
	"google-cloud": {
		Name:        "Google Cloud",
		Type:        "compute",
		BaseURL:     "https://compute.googleapis.com",
		Description: "Google's enterprise cloud with AI integration",
		Icon:        "🔴",
		Plans: []CloudPlan{
			{ID: "gcp-e2micro", Name: "e2-micro", CPU: "2 vCPU", Memory: "1 GB", Storage: "10 GB PD", Bandwidth: "1 Gbps", PriceMonthly: 7.07, PriceHourly: 0.010, Type: "on_demand"},
			{ID: "gcp-e2small", Name: "e2-small", CPU: "2 vCPU", Memory: "2 GB", Storage: "10 GB PD", Bandwidth: "1 Gbps", PriceMonthly: 14.14, PriceHourly: 0.020, Type: "on_demand"},
			{ID: "gcp-n1standard", Name: "n1-standard-2", CPU: "2 vCPU", Memory: "7.5 GB", Storage: "10 GB PD", Bandwidth: "2 Gbps", PriceMonthly: 48.63, PriceHourly: 0.068, Type: "on_demand"},
		},
	},
	"hetzner": {
		Name:        "Hetzner",
		Type:        "compute",
		BaseURL:     "https://api.hetzner.cloud/v1",
		Description: "High-performance servers at unbeatable prices (EU-based)",
		Icon:        "🟡",
		Plans: []CloudPlan{
			{ID: "hx-cpx11", Name: "CPX11", CPU: "2 vCPU", Memory: "2 GB", Storage: "40 GB NVMe", Bandwidth: "20 TB", PriceMonthly: 4.51, PriceHourly: 0.007, Type: "on_demand"},
			{ID: "hx-cpx21", Name: "CPX21", CPU: "3 vCPU", Memory: "4 GB", Storage: "80 GB NVMe", Bandwidth: "20 TB", PriceMonthly: 8.04, PriceHourly: 0.012, Type: "on_demand"},
			{ID: "hx-cpx31", Name: "CPX31", CPU: "4 vCPU", Memory: "8 GB", Storage: "160 GB NVMe", Bandwidth: "20 TB", PriceMonthly: 15.14, PriceHourly: 0.023, Type: "on_demand"},
		},
	},
	"ovhcloud": {
		Name:        "OVHcloud",
		Type:        "compute",
		BaseURL:     "https://api.ovh.com/1.0",
		Description: "European cloud leader with global infrastructure",
		Icon:        "🔷",
		Plans: []CloudPlan{
			{ID: "ovh-starter", Name: "Starter", CPU: "2 vCPU", Memory: "2 GB", Storage: "40 GB SSD", Bandwidth: "Unlimited", PriceMonthly: 5.99, PriceHourly: 0.009, Type: "on_demand"},
			{ID: "ovh-value", Name: "Value", CPU: "4 vCPU", Memory: "8 GB", Storage: "80 GB SSD", Bandwidth: "Unlimited", PriceMonthly: 13.99, PriceHourly: 0.021, Type: "on_demand"},
			{ID: "ovh-performance", Name: "Performance", CPU: "8 vCPU", Memory: "16 GB", Storage: "160 GB NVMe", Bandwidth: "Unlimited", PriceMonthly: 29.99, PriceHourly: 0.045, Type: "on_demand"},
		},
	},
	"vast-ai": {
		Name:        "Vast.ai",
		Type:        "gpu",
		BaseURL:     "https://console.vast.ai/api/v0",
		Description: "GPU marketplace — rent GPUs for AI/ML workloads",
		Icon:        "🟢",
		Plans: []CloudPlan{
			{ID: "vast-rtx3090", Name: "RTX 3090", CPU: "8 vCPU", Memory: "32 GB", Storage: "256 GB", Bandwidth: "1 Gbps", PriceMonthly: 0, PriceHourly: 0.25, Type: "on_demand"},
			{ID: "vast-rtx4090", Name: "RTX 4090", CPU: "16 vCPU", Memory: "64 GB", Storage: "512 GB", Bandwidth: "2 Gbps", PriceMonthly: 0, PriceHourly: 0.45, Type: "on_demand"},
			{ID: "vast-a100", Name: "A100 80GB", CPU: "32 vCPU", Memory: "128 GB", Storage: "1 TB NVMe", Bandwidth: "10 Gbps", PriceMonthly: 0, PriceHourly: 1.50, Type: "on_demand"},
		},
	},
}

// GetCloudProvidersHandler returns available cloud providers and their plans
func GetCloudProvidersHandler(w http.ResponseWriter, r *http.Request) {
	type ProviderResponse struct {
		Name        string      `json:"name"`
		Type        string      `json:"type"`
		Description string      `json:"description"`
		Icon        string      `json:"icon"`
		Plans       []CloudPlan `json:"plans"`
	}

	providers := make([]ProviderResponse, 0, len(CloudProviders))
	for _, p := range CloudProviders {
		providers = append(providers, ProviderResponse{
			Name:        p.Name,
			Type:        p.Type,
			Description: p.Description,
			Icon:        p.Icon,
			Plans:       p.Plans,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers":      providers,
		"network_fee_rate": MarketplaceFeeRate,
		"fee_description": "7% network fee covers operational costs: no ads, easy API management, instant access, full environment control",
	})
}

// CreateCloudPurchaseHandler handles cloud resource purchases
func CreateCloudPurchaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req CloudPurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	provider, ok := CloudProviders[req.Provider]
	if !ok {
		http.Error(w, `{"error":"unsupported cloud provider"}`, http.StatusBadRequest)
		return
	}

	var plan *CloudPlan
	for i := range provider.Plans {
		if provider.Plans[i].ID == req.PlanID {
			plan = &provider.Plans[i]
			break
		}
	}
	if plan == nil {
		http.Error(w, `{"error":"plan not found"}`, http.StatusBadRequest)
		return
	}

	// Calculate cost based on billing type
	var baseCost float64
	if req.BillingType == "on_demand" {
		baseCost = plan.PriceHourly
	} else {
		baseCost = plan.PriceMonthly
	}

	// Calculate 7% network fee
	networkFee := baseCost * MarketplaceFeeRate
	totalCost := baseCost + networkFee

	// Check agent balance
	var balance float64
	err := DB.QueryRow("SELECT budget FROM agents WHERE id=$1", req.AgentID).Scan(&balance)
	if err != nil {
		http.Error(w, `{"error":"agent not found"}`, http.StatusNotFound)
		return
	}

	if balance < totalCost {
		http.Error(w, `{"error":"insufficient balance"}`, http.StatusBadRequest)
		return
	}

	// Check for promo code discount
	promoCode := r.Header.Get("X-Promo-Code")
	discountRate := 0.0
	if promoCode != "" {
		var tier string
		err := DB.QueryRow("SELECT discount_tier FROM promo_codes WHERE code=$1 AND is_active=true", promoCode).Scan(&tier)
		if err == nil {
			switch tier {
			case "3percent":
				discountRate = 0.03
			case "1percent":
				discountRate = 0.01
			}
		}
	}

	appliedFee := MarketplaceFeeRate - discountRate
	if appliedFee < 0 {
		appliedFee = 0
	}
	networkFee = baseCost * appliedFee
	totalCost = baseCost + networkFee

	// Deduct from agent balance
	_, _ = DB.Exec("UPDATE agents SET budget = budget - $1 WHERE id=$2", totalCost, req.AgentID)

	// Record transaction
	purchaseID := uuid.New().String()
	_, _ = DB.Exec(
		"INSERT INTO wallet_transactions (agent_id, type, amount, provider, reference, marketplace_fee, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		req.AgentID, "cloud_compute", totalCost, provider.Name, purchaseID, networkFee, time.Now().Unix(),
	)

	// Log to blockchain (if available)
	action := ActionLogRequest{
		AgentID:  req.AgentID,
		Type:     "CLOUD_PURCHASE",
		Meta:     `{"provider":"` + req.Provider + `","plan":"` + plan.Name + `","billing":"` + req.BillingType + `","total_cost":` + formatFloat(totalCost) + `,` + `"fee":` + formatFloat(networkFee) + `}`,
	}
	logAction(action, "")

	resp := CloudPurchaseResponse{
		PurchaseID:   purchaseID,
		Provider:     provider.Name,
		Plan:         plan.Name,
		TotalCost:    totalCost,
		NetworkFee:   networkFee,
		AgentBalance: balance - totalCost,
		Status:       "pending_provisioning",
		Message:      "Cloud resource purchase recorded. Provisioning will begin shortly. NOTE: Actual API integration requires provider API key configuration.",
	}

	// TODO: Actually provision the cloud resource via provider API
	// This requires:
	// 1. User's provider API key (stored securely)
	// 2. Provider-specific provisioning logic
	// 3. Async job to monitor provisioning status
	// 4. Webhook/callback when resource is ready
	log.Printf("Cloud purchase recorded: %s - %s %s (%s)", purchaseID, provider.Name, plan.Name, req.BillingType)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.4f", f)
}
