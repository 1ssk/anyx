package handlers

import (
	"net/http"

	"anyx/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminHandler управляет административными запросами.
type AdminHandler struct {
	db *gorm.DB
}

// NewAdminHandler создает AdminHandler.
func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// Users возвращает список пользователей.
func (h *AdminHandler) Users(c *gin.Context) {
	var users []models.User
	if err := h.db.Order("created_at desc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить пользователей"})
		return
	}

	response := make([]gin.H, 0, len(users))
	for _, user := range users {
		response = append(response, gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// UserDetail возвращает пользователя, подписку и ссылки.
func (h *AdminHandler) UserDetail(c *gin.Context) {
	var user models.User
	if err := h.db.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	var subscription models.Subscription
	_ = h.db.Where("user_id = ?", user.ID).First(&subscription).Error

	var links []models.InviteLink
	_ = h.db.Where("user_id = ?", user.ID).Order("created_at desc").Find(&links).Error

	linksResponse := make([]gin.H, 0, len(links))
	for _, link := range links {
		linksResponse = append(linksResponse, gin.H{
			"id":         link.ID,
			"code":       link.Code,
			"target_url": link.TargetURL,
			"created_at": link.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		},
		"subscription": gin.H{
			"status":             subscription.Status,
			"trial_ends_at":      subscription.TrialEndsAt,
			"current_period_end": subscription.CurrentPeriodEnd,
		},
		"links": linksResponse,
	})
}

// DeleteUser удаляет пользователя.
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить пользователя"})
		return
	}
	_ = h.db.Where("user_id = ?", id).Delete(&models.InviteLink{}).Error
	_ = h.db.Where("user_id = ?", id).Delete(&models.Subscription{}).Error

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
