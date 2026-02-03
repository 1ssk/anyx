package handlers

import (
	"net/http"
	"time"

	"anyx/internal/models"
	"anyx/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BillingHandler отвечает за оплату и статусы.
type BillingHandler struct {
	db      *gorm.DB
	billing *services.BillingService
}

// NewBillingHandler создает BillingHandler.
func NewBillingHandler(db *gorm.DB, billing *services.BillingService) *BillingHandler {
	return &BillingHandler{db: db, billing: billing}
}

// Status возвращает информацию по подписке.
func (h *BillingHandler) Status(c *gin.Context) {
	userID := c.GetUint("user_id")
	var subscription models.Subscription
	if err := h.db.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "trial"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":             subscription.Status,
		"trial_ends_at":      subscription.TrialEndsAt,
		"current_period_end": subscription.CurrentPeriodEnd,
	})
}

// Checkout создает ссылку на оплату и активирует подписку (заглушка).
func (h *BillingHandler) Checkout(c *gin.Context) {
	userID := c.GetUint("user_id")
	checkoutURL := h.billing.CheckoutURL(userID)

	// Заглушка: сразу продлеваем подписку на 30 дней.
	var subscription models.Subscription
	if err := h.db.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		subscription = models.Subscription{UserID: userID}
	}

	subscription.Status = "active"
	subscription.CurrentPeriodEnd = time.Now().AddDate(0, 1, 0)
	if subscription.TrialEndsAt.IsZero() {
		subscription.TrialEndsAt = time.Now()
	}
	_ = h.db.Save(&subscription).Error

	c.JSON(http.StatusOK, gin.H{"checkout_url": checkoutURL})
}
