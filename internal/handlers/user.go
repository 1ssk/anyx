package handlers

import (
	"net/http"
	"time"

	"anyx/internal/config"
	"anyx/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler отвечает за данные пользователя.
type UserHandler struct {
	db  *gorm.DB
	cfg config.Config
}

// NewUserHandler создает UserHandler.
func NewUserHandler(db *gorm.DB, cfg config.Config) *UserHandler {
	return &UserHandler{db: db, cfg: cfg}
}

// Me возвращает профиль и подписку.
func (h *UserHandler) Me(c *gin.Context) {
	userID := c.GetUint("user_id")
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	var subscription models.Subscription
	_ = h.db.Where("user_id = ?", userID).First(&subscription).Error

	status := subscription.Status
	if status == "" {
		status = "trial"
	}

	trialLeft := 0
	if !subscription.TrialEndsAt.IsZero() {
		trialLeft = int(time.Until(subscription.TrialEndsAt).Hours() / 24)
		if trialLeft < 0 {
			trialLeft = 0
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"email": user.Email,
		"subscription": gin.H{
			"status":            status,
			"trial_ends_at":     subscription.TrialEndsAt,
			"current_period_end": subscription.CurrentPeriodEnd,
			"trial_days_left":   trialLeft,
			"monthly_price":     h.cfg.MonthlyPrice,
		},
	})
}
