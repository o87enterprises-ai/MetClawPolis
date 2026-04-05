package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// BetaSignup represents a beta tester registration
type BetaSignup struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Interests  string `json:"interests"`
	SignedUpAt int64  `json:"signed_up_at"`
	Status     string `json:"status"` // active, completed, churned
	ReferredBy string `json:"referred_by"`
}

// BetaFeedback represents feedback from a beta tester
type BetaFeedback struct {
	ID              string `json:"id"`
	BetaSignupID    string `json:"beta_signup_id"`
	Rating          int    `json:"rating"`
	Comments        string `json:"comments"`
	FeatureRequests string `json:"feature_requests"`
	BugsReported    string `json:"bugs_reported"`
	SubmittedAt     int64  `json:"submitted_at"`
}

// PromoCode represents a discount promo code
type PromoCode struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	BetaSignupID string `json:"beta_signup_id"`
	DiscountTier string `json:"discount_tier"` // "3percent" or "1percent"
	CreatedAt    int64  `json:"created_at"`
	UsedCount    int    `json:"used_count"`
	IsActive     bool   `json:"is_active"`
}

// BetaSignupRequest is the request body for beta signup
type BetaSignupRequest struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	Interests  string `json:"interests"`
	ReferredBy string `json:"referred_by"`
}

// BetaFeedbackRequest is the request body for feedback submission
type BetaFeedbackRequest struct {
	BetaSignupID  string `json:"beta_signup_id"`
	Rating        int    `json:"rating"`
	Comments      string `json:"comments"`
	FeatureReqs   string `json:"feature_requests"`
	BugsReported  string `json:"bugs_reported"`
	HasSharedContent bool `json:"has_shared_content"` // true if user created a short/stream
}

// BetaSignupHandler handles new beta tester registrations
func BetaSignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req BetaSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Name == "" {
		http.Error(w, `{"error":"email and name are required"}`, http.StatusBadRequest)
		return
	}

	// Check if email already exists
	var existing string
	err := DB.QueryRow("SELECT id FROM beta_signups WHERE email=$1", req.Email).Scan(&existing)
	if err == nil {
		http.Error(w, `{"error":"email already registered"}`, http.StatusConflict)
		return
	}

	signup := BetaSignup{
		ID:         uuid.New().String(),
		Email:      req.Email,
		Name:       req.Name,
		Interests:  req.Interests,
		SignedUpAt: time.Now().Unix(),
		Status:     "active",
		ReferredBy: req.ReferredBy,
	}

	_, err = DB.Exec(
		"INSERT INTO beta_signups (id, email, name, interests, signed_up_at, status, referred_by) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		signup.ID, signup.Email, signup.Name, signup.Interests, signup.SignedUpAt, signup.Status, signup.ReferredBy,
	)
	if err != nil {
		log.Printf("Beta signup DB error: %v", err)
		http.Error(w, `{"error":"failed to register beta signup"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("Beta signup: %s (%s)", signup.Name, signup.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"signup_id": signup.ID,
		"message":   "Welcome to the MetClawPolis beta! No CC required. We'll reach out with setup assistance.",
	})
}

// BetaFeedbackHandler handles feedback submission from beta testers
func BetaFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req BetaFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		http.Error(w, `{"error":"rating must be between 1 and 5"}`, http.StatusBadRequest)
		return
	}

	feedback := BetaFeedback{
		ID:             uuid.New().String(),
		BetaSignupID:   req.BetaSignupID,
		Rating:         req.Rating,
		Comments:       req.Comments,
		FeatureRequests: req.FeatureReqs,
		BugsReported:   req.BugsReported,
		SubmittedAt:    time.Now().Unix(),
	}

	_, err := DB.Exec(
		"INSERT INTO beta_feedback (id, beta_signup_id, rating, comments, feature_requests, bugs_reported, submitted_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		feedback.ID, feedback.BetaSignupID, feedback.Rating, feedback.Comments, feedback.FeatureRequests, feedback.BugsReported, feedback.SubmittedAt,
	)
	if err != nil {
		log.Printf("Beta feedback DB error: %v", err)
		http.Error(w, `{"error":"failed to submit feedback"}`, http.StatusInternalServerError)
		return
	}

	// Generate promo code based on feedback quality and sharing
	discountTier := "3percent" // Default for beta feedback
	if req.HasSharedContent {
		discountTier = "1percent" // Better discount for sharing content
	}

	promoCode := generatePromoCode(discountTier, req.BetaSignupID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feedback_id": feedback.ID,
		"promo_code":  promoCode.Code,
		"discount":    promoCode.DiscountTier,
		"message":     "Thank you for your feedback! Your promo code gives you a reduced network fee.",
	})
}

// GeneratePromoCodeHandler generates a promo code for a beta tester
func GeneratePromoCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BetaSignupID    string `json:"beta_signup_id"`
		HasSharedContent bool   `json:"has_shared_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Check for existing active promo code
	var existingCode string
	err := DB.QueryRow("SELECT code FROM promo_codes WHERE beta_signup_id=$1 AND is_active=true", req.BetaSignupID).Scan(&existingCode)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"promo_code": existingCode,
			"message":    "You already have an active promo code.",
		})
		return
	}

	discountTier := "3percent"
	if req.HasSharedContent {
		discountTier = "1percent"
	}

	promoCode := generatePromoCode(discountTier, req.BetaSignupID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"promo_code":    promoCode.Code,
		"discount_tier": promoCode.DiscountTier,
		"description":   getDiscountDescription(promoCode.DiscountTier),
	})
}

// ValidatePromoCodeHandler validates and returns promo code details
func ValidatePromoCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	var promo PromoCode
	err := DB.QueryRow(
		"SELECT id, code, discount_tier, used_count, is_active FROM promo_codes WHERE code=$1",
		req.Code,
	).Scan(&promo.ID, &promo.Code, &promo.DiscountTier, &promo.UsedCount, &promo.IsActive)
	if err != nil {
		http.Error(w, `{"error":"promo code not found"}`, http.StatusNotFound)
		return
	}

	if !promo.IsActive {
		http.Error(w, `{"error":"promo code is expired or inactive"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":         true,
		"code":          promo.Code,
		"discount_tier": promo.DiscountTier,
		"description":   getDiscountDescription(promo.DiscountTier),
		"used_count":    promo.UsedCount,
	})
}

// GetPromoCodeHandler retrieves a specific promo code
func GetPromoCodeHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error":"code query parameter required"}`, http.StatusBadRequest)
		return
	}

	var promo PromoCode
	err := DB.QueryRow(
		"SELECT id, code, discount_tier, used_count, is_active FROM promo_codes WHERE code=$1",
		code,
	).Scan(&promo.ID, &promo.Code, &promo.DiscountTier, &promo.UsedCount, &promo.IsActive)
	if err != nil {
		http.Error(w, `{"error":"promo code not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":          promo.Code,
		"discount_tier": promo.DiscountTier,
		"description":   getDiscountDescription(promo.DiscountTier),
		"is_active":     promo.IsActive,
		"used_count":    promo.UsedCount,
	})
}

// generatePromoCode creates a new promo code in the database
func generatePromoCode(tier, betaSignupID string) PromoCode {
	// Generate a human-readable promo code like "BETA-XXXX-7PCT"
	promoID := uuid.New().String()[:6]
	suffix := "7NET"
	if tier == "3percent" {
		suffix = "3NET"
	} else if tier == "1percent" {
		suffix = "1NET"
	}
	code := "BETA-" + promoID + "-" + suffix

	promo := PromoCode{
		ID:           uuid.New().String(),
		Code:         code,
		BetaSignupID: betaSignupID,
		DiscountTier: tier,
		CreatedAt:    time.Now().Unix(),
		UsedCount:    0,
		IsActive:     true,
	}

	_, err := DB.Exec(
		"INSERT INTO promo_codes (id, code, beta_signup_id, discount_tier, created_at, used_count, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		promo.ID, promo.Code, promo.BetaSignupID, promo.DiscountTier, promo.CreatedAt, promo.UsedCount, promo.IsActive,
	)
	if err != nil {
		log.Printf("Promo code creation error: %v", err)
	}

	return promo
}

func getDiscountDescription(tier string) string {
	switch tier {
	case "3percent":
		return "Reduced 3% network fee (vs standard 7%) — for beta testers who provided honest feedback"
	case "1percent":
		return "Ultra 1% network fee (vs standard 7%) — for beta testers who shared content about MetClawPolis"
	default:
		return "Standard 7% network fee"
	}
}
