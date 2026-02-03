package handlers

import (
	"crypto/subtle"
	"io"
	"net/http"
	"time"

	"anyx/internal/config"
	"anyx/internal/models"
	"anyx/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BillingHandler отвечает за оплату и статусы.
type BillingHandler struct {
	db       *gorm.DB
	billing  *services.BillingService
	yookassa *services.YooKassaService
	cfg      config.Config
}

// NewBillingHandler создает BillingHandler.
func NewBillingHandler(db *gorm.DB, billing *services.BillingService, yookassa *services.YooKassaService, cfg config.Config) *BillingHandler {
	return &BillingHandler{db: db, billing: billing, yookassa: yookassa, cfg: cfg}
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

// Checkout создает ссылку на оплату.
func (h *BillingHandler) Checkout(c *gin.Context) {
	if h.cfg.YooShopID == "" || h.cfg.YooSecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не заполнены реквизиты ЮKassa"})
		return
	}

	userID := c.GetUint("user_id")

	confirmation, err := h.yookassa.CreatePayment(userID, h.cfg.MonthlyPrice, h.cfg.CheckoutReturn)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Не удалось создать оплату"})
		return
	}

	var subscription models.Subscription
	if err := h.db.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		subscription = models.Subscription{UserID: userID, Status: "pending"}
	} else {
		subscription.Status = "pending"
	}
	_ = h.db.Save(&subscription).Error

	c.JSON(http.StatusOK, gin.H{"checkout_url": confirmation.URL})
}

// Webhook принимает уведомления от ЮKassa.
func (h *BillingHandler) Webhook(c *gin.Context) {
	if h.cfg.WebhookSecret != "" {
		secret := c.GetHeader("Authorization")
		if subtle.ConstantTimeCompare([]byte(secret), []byte(h.cfg.WebhookSecret)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный секрет"})
			return
		}
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось прочитать запрос"})
		return
	}

	event, err := h.yookassa.ParseWebhook(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный payload"})
		return
	}

	userID := event.Object.Meta.UserID
	if userID == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	switch event.Event {
	case "payment.succeeded":
		if !event.Object.Paid {
			c.JSON(http.StatusOK, gin.H{"status": "ignored"})
			return
		}
		months := event.Object.Meta.Months
		if months <= 0 {
			months = 1
		}

		var subscription models.Subscription
		if err := h.db.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
			subscription = models.Subscription{UserID: userID}
		}

		subscription.Status = "active"
		base := time.Now()
		if subscription.CurrentPeriodEnd.After(base) {
			base = subscription.CurrentPeriodEnd
		}
		subscription.CurrentPeriodEnd = base.AddDate(0, months, 0)
		if subscription.TrialEndsAt.IsZero() {
			subscription.TrialEndsAt = time.Now()
		}
		_ = h.db.Save(&subscription).Error

	case "payment.waiting_for_capture":
		_ = h.db.Model(&models.Subscription{}).Where("user_id = ?", userID).Update("status", "pending").Error
	case "payment.canceled":
		_ = h.db.Model(&models.Subscription{}).Where("user_id = ?", userID).Update("status", "canceled").Error
	default:
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
