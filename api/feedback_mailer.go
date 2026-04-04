package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
)

// FeedbackMailer handles email communication with beta testers
type FeedbackMailer struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPass     string
	FromEmail    string
	FromName     string
}

// NewFeedbackMailer creates a new mailer from environment variables
func NewFeedbackMailer() *FeedbackMailer {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		log.Println("FeedbackMailer: SMTP not configured, email sending disabled")
		return nil
	}

	return &FeedbackMailer{
		SMTPHost: host,
		SMTPPort: getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
		FromEmail: getEnvOrDefault("SMTP_FROM_EMAIL", "beta@metclawpolis.com"),
		FromName:  getEnvOrDefault("SMTP_FROM_NAME", "MetClawPolis Beta Team"),
	}
}

// BetaTester represents a beta tester for email purposes
type BetaTester struct {
	Email string
	Name  string
}

// SendFeedbackRequest sends a feedback request email to a beta tester
func (m *FeedbackMailer) SendFeedbackRequest(tester BetaTester) error {
	if m == nil {
		log.Printf("FeedbackMailer: mailer not configured, skipping email to %s", tester.Email)
		return nil
	}

	subject := "We'd love your feedback on MetClawPolis! 🚀"

	bodyVars := map[string]string{
		"Name":    tester.Name,
		"FeedbackURL": fmt.Sprintf("https://metclawpolis.com/beta/feedback?email=%s", tester.Email),
		"PromoURL":    fmt.Sprintf("https://metclawpolis.com/beta/promo?email=%s", tester.Email),
	}

	htmlBody, err := renderEmailTemplate(feedbackRequestTemplate, bodyVars)
	if err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	msg := buildEmailMessage(m.FromEmail, m.FromName, tester.Email, subject, htmlBody)

	auth := smtp.PlainAuth("", m.SMTPUser, m.SMTPPass, m.SMTPHost)
	addr := m.SMTPHost + ":" + m.SMTPPort

	if err := smtp.SendMail(addr, auth, m.FromEmail, []string{tester.Email}, []byte(msg)); err != nil {
		log.Printf("Failed to send feedback request to %s: %v", tester.Email, err)
		return err
	}

	log.Printf("Feedback request sent to %s", tester.Email)
	return nil
}

// SendPromoCodeOffer sends a promo code offer email
func (m *FeedbackMailer) SendPromoCodeOffer(tester BetaTester, promoCode string, discountTier string) error {
	if m == nil {
		log.Printf("FeedbackMailer: mailer not configured, skipping promo email to %s", tester.Email)
		return nil
	}

	subject := "Your exclusive MetClawPolis discount code is inside! 🎉"

	bodyVars := map[string]string{
		"Name":       tester.Name,
		"PromoCode":  promoCode,
		"Discount":   getDiscountDescription(discountTier),
		"SignupURL":  "https://metclawpolis.com/beta/signup",
	}

	htmlBody, err := renderEmailTemplate(promoOfferTemplate, bodyVars)
	if err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	msg := buildEmailMessage(m.FromEmail, m.FromName, tester.Email, subject, htmlBody)

	auth := smtp.PlainAuth("", m.SMTPUser, m.SMTPPass, m.SMTPHost)
	addr := m.SMTPHost + ":" + m.SMTPPort

	if err := smtp.SendMail(addr, auth, m.FromEmail, []string{tester.Email}, []byte(msg)); err != nil {
		log.Printf("Failed to send promo offer to %s: %v", tester.Email, err)
		return err
	}

	log.Printf("Promo code offer sent to %s (code: %s)", tester.Email, promoCode)
	return nil
}

// SendFeedbackRequestHandler is the HTTP handler for triggering feedback emails
func SendFeedbackRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	// Look up beta tester
	var name string
	err := DB.QueryRow("SELECT name FROM beta_signups WHERE email=$1", req.Email).Scan(&name)
	if err != nil {
		http.Error(w, `{"error":"beta tester not found"}`, http.StatusNotFound)
		return
	}

	mailer := NewFeedbackMailer()
	tester := BetaTester{Email: req.Email, Name: name}

	if err := mailer.SendFeedbackRequest(tester); err != nil {
		http.Error(w, `{"error":"failed to send email"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Feedback request email sent to " + req.Email,
	})
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func buildEmailMessage(from, fromName, to, subject, htmlBody string) string {
	return fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
			"%s",
		fromName, from, to, subject, htmlBody,
	)
}

func renderEmailTemplate(tmplText string, data map[string]string) (string, error) {
	tmpl, err := template.New("email").Parse(tmplText)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const feedbackRequestTemplate = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><style>
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0a0e1a;color:#e8eaf0;margin:0;padding:0}
.container{max-width:600px;margin:0 auto;padding:40px 20px}
.header{text-align:center;padding:30px 0;border-bottom:2px solid #00f5ff33}
.header h1{color:#00f5ff;font-size:28px;margin:0}
.content{padding:30px 0}
.btn{display:inline-block;padding:14px 32px;background:linear-gradient(90deg,#00f5ff,#b44fff);color:#0a0e1a;text-decoration:none;border-radius:8px;font-weight:700;font-size:16px;margin:10px 5px}
.btn-secondary{background:rgba(255,255,255,0.1);color:#e8eaf0;border:1px solid #ffffff33}
.section{background:rgba(255,255,255,0.05);border-radius:12px;padding:20px;margin:20px 0;border:1px solid #ffffff11}
.footer{text-align:center;padding:20px 0;color:#666;font-size:12px;border-top:1px solid #ffffff11}
</style></head>
<body>
<div class="container">
<div class="header"><h1>🚀 MetClawPolis</h1><p>Beta Testing Program</p></div>
<div class="content">
<p>Hi {{.Name}},</p>
<p>Thank you for being part of the MetClawPolis beta program! Your experience and insights are invaluable to shaping the future of agentic commerce.</p>

<div class="section">
<h3>📝 We'd Love Your Honest Feedback</h3>
<p>What worked? What didn't? What features do you want next? Your honest feedback directly influences our development roadmap.</p>
<p><a href="{{.FeedbackURL}}" class="btn">Share Feedback</a></p>
</div>

<div class="section">
<h3>🎁 Get a Discount Code</h3>
<p>After sharing your feedback, you'll receive an exclusive promo code for reduced network fees on all transactions.</p>
<p><a href="{{.PromoURL}}" class="btn btn-secondary">Claim Your Promo Code</a></p>
</div>

<p>No credit card required. No strings attached. Just honest thoughts.</p>
<p>— The MetClawPolis Team</p>
</div>
<div class="footer">
<p>MetClawPolis · Agentic Commerce Platform · <a href="https://metclawpolis.com" style="color:#00f5ff">metclawpolis.com</a></p>
<p>You received this email because you signed up for the beta program.</p>
</div>
</div>
</body>
</html>`

const promoOfferTemplate = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><style>
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0a0e1a;color:#e8eaf0;margin:0;padding:0}
.container{max-width:600px;margin:0 auto;padding:40px 20px}
.header{text-align:center;padding:30px 0;border-bottom:2px solid #ffd70033}
.header h1{color:#ffd700;font-size:28px;margin:0}
.promo-box{background:linear-gradient(135deg,#ffd70022,#b44fff22);border:2px solid #ffd70066;border-radius:16px;padding:30px;text-align:center;margin:20px 0}
.promo-code{font-size:36px;font-weight:900;letter-spacing:4px;color:#ffd700;margin:15px 0;font-family:monospace}
.content{padding:30px 0}
.btn{display:inline-block;padding:14px 32px;background:linear-gradient(90deg,#ffd700,#ff8c00);color:#0a0e1a;text-decoration:none;border-radius:8px;font-weight:700;font-size:16px;margin:10px 5px}
.section{background:rgba(255,255,255,0.05);border-radius:12px;padding:20px;margin:20px 0;border:1px solid #ffffff11}
.footer{text-align:center;padding:20px 0;color:#666;font-size:12px;border-top:1px solid #ffffff11}
</style></head>
<body>
<div class="container">
<div class="header"><h1>🎉 Your Exclusive Discount</h1></div>
<div class="content">
<p>Hi {{.Name}},</p>
<p>As a thank-you for being part of the MetClawPolis beta, here's your exclusive network fee discount:</p>

<div class="promo-box">
<div style="font-size:14px;color:#aaa;margin-bottom:5px">YOUR PROMO CODE</div>
<div class="promo-code">{{.PromoCode}}</div>
<div style="font-size:14px;color:#ffd700">{{.Discount}}</div>
</div>

<div class="section">
<h3>✨ What This Means</h3>
<p>The standard network fee is <strong>7%</strong> on all transactions. Your promo code reduces this significantly. The fee covers operational costs — no ads, easy API management, instant access, and full environment control.</p>
<p><strong>Bonus:</strong> If you create a short video or live stream about MetClawPolis and share it, you can qualify for an even better 1% rate!</p>
</div>

<p><a href="{{.SignupURL}}" class="btn">Start Using Your Discount</a></p>

<p>— The MetClawPolis Team</p>
</div>
<div class="footer">
<p>MetClawPolis · Agentic Commerce Platform · <a href="https://metclawpolis.com" style="color:#ffd700">metclawpolis.com</a></p>
<p>You received this email because you signed up for the beta program.</p>
</div>
</div>
</body>
</html>`
