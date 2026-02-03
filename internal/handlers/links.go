package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"anyx/internal/config"
	"anyx/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LinkHandler отвечает за работу со ссылками.
type LinkHandler struct {
	db  *gorm.DB
	cfg config.Config
}

// NewLinkHandler создает LinkHandler.
func NewLinkHandler(db *gorm.DB, cfg config.Config) *LinkHandler {
	return &LinkHandler{db: db, cfg: cfg}
}

// List возвращает список ссылок пользователя.
func (h *LinkHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	var links []models.InviteLink
	if err := h.db.Where("user_id = ?", userID).Order("created_at desc").Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить ссылки"})
		return
	}

	response := make([]gin.H, 0, len(links))
	for _, link := range links {
		response = append(response, gin.H{
			"id":         link.ID,
			"code":       link.Code,
			"target_url": link.TargetURL,
			"created_at": link.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// Create создает инвайт-ссылку.
func (h *LinkHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		TargetURL string `json:"target_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}
	if strings.TrimSpace(req.TargetURL) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ссылка обязательна"})
		return
	}

	code, err := generateCode(6)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать ссылку"})
		return
	}

	link := models.InviteLink{UserID: userID, Code: code, TargetURL: req.TargetURL}
	if err := h.db.Create(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить ссылку"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": link.Code, "id": link.ID})
}

// Stats возвращает статистику кликов.
func (h *LinkHandler) Stats(c *gin.Context) {
	userID := c.GetUint("user_id")
	var link models.InviteLink
	if err := h.db.Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ссылка не найдена"})
		return
	}

	var clicks []models.Click
	if err := h.db.Where("link_id = ?", link.ID).Order("clicked_at desc").Find(&clicks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить статистику"})
		return
	}

	entries := make([]gin.H, 0, len(clicks))
	for _, click := range clicks {
		entries = append(entries, gin.H{"clicked_at": click.ClickedAt})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   link.Code,
		"target": link.TargetURL,
		"total":  len(clicks),
		"clicks": entries,
	})
}

// Redirect фиксирует клик и редиректит.
func (h *LinkHandler) Redirect(c *gin.Context) {
	var link models.InviteLink
	if err := h.db.Where("code = ?", c.Param("code")).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ссылка не найдена"})
		return
	}

	click := models.Click{LinkID: link.ID, ClickedAt: time.Now()}
	_ = h.db.Create(&click).Error

	c.Redirect(http.StatusFound, link.TargetURL)
}

func generateCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
