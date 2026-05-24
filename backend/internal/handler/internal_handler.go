package handler

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// InternalBillingHandler handles internal billing operations from gpt2api.
// Shared secret auth (X-Internal-Secret) + user_id in request body.
type InternalBillingHandler struct {
	secret    string
	userRepo  service.UserRepository
	entClient *ent.Client
	mu        sync.Mutex
	frozen    map[string]*frozenBalance
}

type frozenBalance struct {
	UserID uint64
	Amount float64
	Model  string
}

// NewInternalBillingHandler creates the handler.
func NewInternalBillingHandler(userRepo service.UserRepository, entClient *ent.Client) *InternalBillingHandler {
	secret := os.Getenv("INTERNAL_SECRET")
	if secret == "" {
		secret = "sub2api-internal-secret-change-me"
	}
	return &InternalBillingHandler{
		secret:    secret,
		userRepo:  userRepo,
		entClient: entClient,
		frozen:    make(map[string]*frozenBalance),
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

// PreDeduct checks balance and freezes the amount.
func (h *InternalBillingHandler) PreDeduct(c *gin.Context) {
	if !h.checkSecret(c) {
		c.JSON(http.StatusForbidden, preDeductResp{Error: "unauthorized"})
		return
	}

	var req preDeductReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 || req.UserID == 0 {
		c.JSON(http.StatusBadRequest, preDeductResp{Error: "invalid request"})
		return
	}

	ctx := context.Background()
	user, err := h.entClient.User.Get(ctx, int64(req.UserID))
	if err != nil {
		c.JSON(http.StatusNotFound, preDeductResp{Error: "user not found"})
		return
	}
	if user.Balance < req.Amount {
		c.JSON(http.StatusPaymentRequired, preDeductResp{Error: "insufficient balance"})
		return
	}

	frozenID := req.TaskID
	if frozenID == "" {
		frozenID = "frozen_" + time.Now().Format("20060102150405")
	}

	h.mu.Lock()
	h.frozen[frozenID] = &frozenBalance{UserID: req.UserID, Amount: req.Amount, Model: req.Model}
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

// Settle actually deducts the frozen amount from the user's balance.
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

	// Actually deduct from user balance
	if err := h.userRepo.DeductBalance(context.Background(), int64(frozen.UserID), frozen.Amount); err != nil {
		slog.Error("internal.billing.settle.deduct_failed", "user_id", frozen.UserID, "amount", frozen.Amount, "error", err)
		c.JSON(http.StatusInternalServerError, settleResp{Error: "deduct failed"})
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

// Refund releases a frozen balance without deduction.
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
	_, ok := h.frozen[req.FrozenID]
	if ok {
		delete(h.frozen, req.FrozenID)
	}
	h.mu.Unlock()

	if !ok {
		c.JSON(http.StatusOK, refundResp{Success: true})
		return
	}

	slog.Info("internal.billing.refund", "frozen_id", req.FrozenID)

	c.JSON(http.StatusOK, refundResp{Success: true})
}
