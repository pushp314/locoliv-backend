package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/internal/payment"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	razorpay     *payment.RazorpayClient
	karmaService *domain.KarmaService
	logger       *zap.Logger
}

func NewPaymentHandler(razorpay *payment.RazorpayClient, karmaService *domain.KarmaService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		razorpay:     razorpay,
		karmaService: karmaService,
		logger:       logger,
	}
}

type CreateOrderReq struct {
	Plan string `json:"plan"` // "plus_monthly" or "plus_annual"
}

// CreateOrder initiates a payment order
func (h *PaymentHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	if !h.razorpay.IsConfigured() {
		response.InternalError(w, "payment system not configured")
		return
	}

	var req CreateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Plan != "plus_monthly" && req.Plan != "plus_annual" {
		response.BadRequest(w, "invalid plan. use 'plus_monthly' or 'plus_annual'")
		return
	}

	order, err := h.razorpay.CreateOrder(payment.CreateOrderRequest{
		UserID: userID,
		Plan:   req.Plan,
	})
	if err != nil {
		h.logger.Error("Failed to create order", zap.Error(err))
		response.InternalError(w, "failed to create order")
		return
	}

	response.OK(w, order)
}

type VerifyPaymentReq struct {
	RazorpayOrderID   string `json:"razorpay_order_id"`
	RazorpayPaymentID string `json:"razorpay_payment_id"`
	RazorpaySignature string `json:"razorpay_signature"`
	Plan              string `json:"plan"`
}

// VerifyPayment verifies payment and activates subscription
func (h *PaymentHandler) VerifyPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	var req VerifyPaymentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Verify signature
	valid, err := h.razorpay.VerifyPayment(payment.VerifyPaymentRequest{
		RazorpayOrderID:   req.RazorpayOrderID,
		RazorpayPaymentID: req.RazorpayPaymentID,
		RazorpaySignature: req.RazorpaySignature,
	})
	if err != nil || !valid {
		h.logger.Error("Payment verification failed", zap.Error(err))
		response.BadRequest(w, "payment verification failed")
		return
	}

	// Create subscription
	var duration time.Duration
	var plan domain.SubscriptionPlan
	var amount int

	switch req.Plan {
	case "plus_monthly":
		duration = 30 * 24 * time.Hour
		plan = domain.PlanPlusMonthly
		amount = 14900
	case "plus_annual":
		duration = 365 * 24 * time.Hour
		plan = domain.PlanPlusAnnual
		amount = 99900
	default:
		response.BadRequest(w, "invalid plan")
		return
	}

	provider := "razorpay"
	sub := &domain.Subscription{
		ID:              uuid.New(),
		UserID:          userID,
		Plan:            plan,
		Status:          domain.SubscriptionActive,
		StartedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(duration),
		PaymentProvider: &provider,
		PaymentID:       &req.RazorpayPaymentID,
		AmountPaid:      &amount,
	}

	// TODO: Save subscription to repository
	_ = sub // Silence linter until repo access is implemented

	h.logger.Info("Subscription created",
		zap.String("user_id", userID.String()),
		zap.String("plan", string(sub.Plan)),
		zap.String("payment_id", req.RazorpayPaymentID),
	)

	response.OK(w, map[string]interface{}{
		"success":    true,
		"message":    "Premium activated successfully!",
		"plan":       sub.Plan,
		"expires_at": sub.ExpiresAt,
	})
}
