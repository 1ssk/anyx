package handlers

import (
	"net/http"

	"anyx/internal/config"

	"github.com/gin-gonic/gin"
)

// ConfigHandler возвращает публичные параметры фронта.
func ConfigHandler(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app_domain":    cfg.AppDomain,
			"monthly_price": cfg.MonthlyPrice,
			"trial_days":    cfg.TrialDays,
		})
	}
}
