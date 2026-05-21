package handler

import (
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// InternalBillingHandler handles internal billing operations from gpt2api.
type InternalBillingHandler struct {
	sharedSecret string
	mu           sync.Mutex
	frozen       map[string]*frozenBalance
}

type frozenBalance struct {
	UserID  uint64
	Amount  float64
	Model   string
	Expires time.Time
}

// NewInternalBillingHandler creates the handler.
func NewInternalBillingHandler() *InternalBillingHandler {
	secret := os.Getenv("INTERNAL_SECRET")
	if secret == "" {
		secret = "sub2api-internal-secret-change-me"
	}
	return &InternalBillingHandler{
		sharedSecret: secret,
		frozen:       make(map[string]*frozenBalance),
	}
}

// Auth validates the shared secret for internal API calls.
func (h *InternalBillingHandler) Auth(c *gin.Context) {
	secret := c.GetHeader("X-Internal-Secret")
	if secret == "" || secret != h.sharedSecret {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}
	c.Next()
}

// RegisterRoutes registers internal billing routes on the given router group.
func (h *InternalBillingHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.Use(h.Auth)
	r.POST("/pre-deduct", h.PreDeduct)
	r.POST("/settle", h.Settle)
	r.POST("/refund", h.Refund)
}

type preDeductReq struct {
	UserID uint64  `json:"user_id"`
	Amount float64 `json:"amount"` // USD
	Model  string  `json:"model"`
	TaskID string  `json:"task_id"`
}

type preDeductResp struct {
	Success  bool   `json:"success"`
	FrozenID string `json:"frozen_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

// PreDeduct freezes a user's balance before generation.
// POST /api/v1/internal/billing/pre-deduct
func (h *InternalBillingHandler) PreDeduct(c *gin.Context) {
	var req preDeductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, preDeductResp{Error: "invalid request"})
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, preDeductResp{Error: "amount must be positive"})
		return
	}

	// TODO: verify user exists and has enough balance from DB
	// For now, just freeze the amount

	frozenID := req.TaskID
	if frozenID == "" {
		frozenID = "frozen_" + time.Now().Format("20060102150405") + "_" + string(rune(req.UserID))
	}

	h.mu.Lock()
	h.frozen[frozenID] = &frozenBalance{
		UserID:  req.UserID,
		Amount:  req.Amount,
		Model:   req.Model,
		Expires: time.Now().Add(10 * time.Minute),
	}
	h.mu.Unlock()

	slog.Info("internal.billing.pre_deduct",
		"user_id", req.UserID,
		"amount", req.Amount,
		"model", req.Model,
		"frozen_id", frozenID,
	)

	c.JSON(http.StatusOK, preDeductResp{Success: true, FrozenID: frozenID})
}

type settleReq struct {
	FrozenID string `json:"frozen_id"`
}

type settleResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Settle confirms a frozen balance deduction.
// POST /api/v1/internal/billing/settle
func (h *InternalBillingHandler) Settle(c *gin.Context) {
	var req settleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, settleResp{Error: "invalid request"})
		return
	}

	h.mu.Lock()
	frozen, ok := h.frozen[req.FrozenID]
	if ok {
		delete(h.frozen, req.FrozenID)
	}
	h.mu.Unlock()

	if !ok {
		slog.Warn("internal.billing.settle.not_found", "frozen_id", req.FrozenID)
		// Idempotent: already settled is OK
		c.JSON(http.StatusOK, settleResp{Success: true})
		return
	}

	slog.Info("internal.billing.settle",
		"user_id", frozen.UserID,
		"amount", frozen.Amount,
		"model", frozen.Model,
	)

	// TODO: actually deduct from user balance in DB

	c.JSON(http.StatusOK, settleResp{Success: true})
}

type refundReq struct {
	FrozenID string `json:"frozen_id"`
}

type refundResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Refund releases a frozen balance back to the user.
// POST /api/v1/internal/billing/refund
func (h *InternalBillingHandler) Refund(c *gin.Context) {
	var req refundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, refundResp{Error: "invalid request"})
		return
	}

	h.mu.Lock()
	frozen, ok := h.frozen[req.FrozenID]
	if ok {
		delete(h.frozen, req.FrozenID)
	}
	h.mu.Unlock()

	if !ok {
		slog.Warn("internal.billing.refund.not_found", "frozen_id", req.FrozenID)
		c.JSON(http.StatusOK, refundResp{Success: true})
		return
	}

	slog.Info("internal.billing.refund",
		"user_id", frozen.UserID,
		"amount", frozen.Amount,
		"model", frozen.Model,
	)

	c.JSON(http.StatusOK, refundResp{Success: true})
}

// cleanupExpired removes frozen balances past their expiration.
func (h *InternalBillingHandler) cleanupExpired() {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	for id, frozen := range h.frozen {
		if now.After(frozen.Expires) {
			slog.Info("internal.billing.expired_refund", "frozen_id", id, "user_id", frozen.UserID, "amount", frozen.Amount)
			delete(h.frozen, id)
		}
	}
}

// StartCleanup runs periodic cleanup of expired frozen balances.
func (h *InternalBillingHandler) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			h.cleanupExpired()
		}
	}()
}
