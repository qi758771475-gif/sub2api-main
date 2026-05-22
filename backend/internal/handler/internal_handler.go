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
// Authentication is via shared secret (X-Internal-Secret header).
// user_id is passed in the request body for audit trail.
type InternalBillingHandler struct {
	secret string
	mu     sync.Mutex
	frozen map[string]*frozenBalance
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
		secret: secret,
		frozen: make(map[string]*frozenBalance),
	}
}

func (h *InternalBillingHandler) checkSecret(c *gin.Context) bool {
	return c.GetHeader("X-Internal-Secret") == h.secret
}

// RegisterRoutes registers internal billing routes.
func (h *InternalBillingHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/pre-deduct", h.PreDeduct)
	r.POST("/settle", h.Settle)
	r.POST("/refund", h.Refund)
}

type preDeductReq struct {
	UserID uint64  `json:"user_id"`
	Amount float64 `json:"amount"`
	Model  string  `json:"model"`
	TaskID string  `json:"task_id"`
}

type preDeductResp struct {
	Success  bool   `json:"success"`
	FrozenID string `json:"frozen_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (h *InternalBillingHandler) PreDeduct(c *gin.Context) {
	if !h.checkSecret(c) {
		c.JSON(http.StatusForbidden, preDeductResp{Error: "unauthorized"})
		return
	}

	var req preDeductReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, preDeductResp{Error: "invalid request"})
		return
	}

	frozenID := req.TaskID
	if frozenID == "" {
		frozenID = "frozen_" + time.Now().Format("20060102150405")
	}

	h.mu.Lock()
	h.frozen[frozenID] = &frozenBalance{
		UserID:  req.UserID,
		Amount:  req.Amount,
		Model:   req.Model,
		Expires: time.Now().Add(10 * time.Minute),
	}
	h.mu.Unlock()

	slog.Info("internal.billing.pre_deduct", "user_id", req.UserID, "amount", req.Amount, "model", req.Model, "frozen_id", frozenID)

	c.JSON(http.StatusOK, preDeductResp{Success: true, FrozenID: frozenID})
}

type settleReq struct {
	FrozenID string `json:"frozen_id"`
}

type settleResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (h *InternalBillingHandler) Settle(c *gin.Context) {
	if !h.checkSecret(c) {
		c.JSON(http.StatusForbidden, settleResp{Error: "unauthorized"})
		return
	}

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
		c.JSON(http.StatusOK, settleResp{Success: true})
		return
	}

	slog.Info("internal.billing.settle", "user_id", frozen.UserID, "amount", frozen.Amount, "model", frozen.Model)

	c.JSON(http.StatusOK, settleResp{Success: true})
}

type refundReq struct {
	FrozenID string `json:"frozen_id"`
}

type refundResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (h *InternalBillingHandler) Refund(c *gin.Context) {
	if !h.checkSecret(c) {
		c.JSON(http.StatusForbidden, refundResp{Error: "unauthorized"})
		return
	}

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
		c.JSON(http.StatusOK, refundResp{Success: true})
		return
	}

	slog.Info("internal.billing.refund", "user_id", frozen.UserID, "amount", frozen.Amount, "model", frozen.Model)

	c.JSON(http.StatusOK, refundResp{Success: true})
}

func (h *InternalBillingHandler) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			h.cleanupExpired()
		}
	}()
}

func (h *InternalBillingHandler) cleanupExpired() {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	for id, fb := range h.frozen {
		if now.After(fb.Expires) {
			slog.Info("internal.billing.expired_refund", "frozen_id", id, "user_id", fb.UserID, "amount", fb.Amount)
			delete(h.frozen, id)
		}
	}
}
