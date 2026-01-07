package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// PlanPricing in paise (INR)
var PlanPricing = map[string]int{
	"plus_monthly": 14900, // ₹149
	"plus_annual":  99900, // ₹999
}

// RazorpayConfig holds Razorpay credentials
type RazorpayConfig struct {
	KeyID     string
	KeySecret string
}

// RazorpayClient handles Razorpay operations
type RazorpayClient struct {
	config RazorpayConfig
}

// NewRazorpayClient creates a new Razorpay client
func NewRazorpayClient(keyID, keySecret string) *RazorpayClient {
	return &RazorpayClient{
		config: RazorpayConfig{
			KeyID:     keyID,
			KeySecret: keySecret,
		},
	}
}

// IsConfigured checks if Razorpay is properly configured
func (c *RazorpayClient) IsConfigured() bool {
	return c.config.KeyID != "" && c.config.KeySecret != ""
}

// CreateOrderRequest for creating a Razorpay order
type CreateOrderRequest struct {
	UserID uuid.UUID
	Plan   string
}

// OrderResponse returned to mobile app
type OrderResponse struct {
	OrderID       string `json:"order_id"`
	Amount        int    `json:"amount"`
	Currency      string `json:"currency"`
	RazorpayKeyID string `json:"razorpay_key_id"`
	PlanName      string `json:"plan_name"`
}

// CreateOrder generates order details for Razorpay checkout
// Note: In production, you'd call Razorpay's API to create an actual order
// For MVP, we generate a reference ID and verify payment on callback
func (c *RazorpayClient) CreateOrder(req CreateOrderRequest) (*OrderResponse, error) {
	amount, ok := PlanPricing[req.Plan]
	if !ok {
		return nil, errors.New("invalid plan")
	}

	// Generate order reference (in production, call Razorpay API)
	orderID := fmt.Sprintf("order_%s", uuid.New().String()[:12])

	return &OrderResponse{
		OrderID:       orderID,
		Amount:        amount,
		Currency:      "INR",
		RazorpayKeyID: c.config.KeyID,
		PlanName:      req.Plan,
	}, nil
}

// VerifyPaymentRequest for verifying Razorpay payment
type VerifyPaymentRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id"`
	RazorpayPaymentID string `json:"razorpay_payment_id"`
	RazorpaySignature string `json:"razorpay_signature"`
}

// VerifyPayment validates the Razorpay signature
func (c *RazorpayClient) VerifyPayment(req VerifyPaymentRequest) (bool, error) {
	// Generate expected signature
	message := req.RazorpayOrderID + "|" + req.RazorpayPaymentID
	h := hmac.New(sha256.New, []byte(c.config.KeySecret))
	h.Write([]byte(message))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(expectedSignature), []byte(req.RazorpaySignature)), nil
}
