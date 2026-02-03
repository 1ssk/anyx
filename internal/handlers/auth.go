package handlers

import (
	"net/http"
	"strings"
	"time"

	"anyx/internal/config"
	"anyx/internal/models"
	"anyx/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler отвечает за регистрацию и вход.
type AuthHandler struct {
	db         *gorm.DB
	jwtService *services.JWTService
	cfg        config.Config
}

// NewAuthHandler создает AuthHandler.
func NewAuthHandler(db *gorm.DB, jwtService *services.JWTService, cfg config.Config) *AuthHandler {
	return &AuthHandler{db: db, jwtService: jwtService, cfg: cfg}
}

// Register регистрирует пользователя и выдает токен.
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}
	if strings.TrimSpace(req.Email) == "" || len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email и пароль обязательны"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя"})
		return
	}

	user := models.User{Email: strings.ToLower(req.Email), PasswordHash: string(hash), Role: "user"}
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь уже существует"})
		return
	}

	subscription := models.Subscription{
		UserID:      user.ID,
		Status:      "trial",
		TrialEndsAt: time.Now().AddDate(0, 0, h.cfg.TrialDays),
	}
	_ = h.db.Create(&subscription).Error

	token, err := h.jwtService.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать токен"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Login выполняет вход и возвращает токен.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", strings.ToLower(req.Email)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные"})
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать токен"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
