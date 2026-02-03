package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Пользователь сервиса.
type User struct {
	ID           uint      `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time
}

// Ссылка-инвайт, созданная пользователем.
type InviteLink struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	Code      string    `gorm:"uniqueIndex"`
	TargetURL string    `gorm:"not null"`
	CreatedAt time.Time
}

// Клик по инвайт-ссылке.
type Click struct {
	ID        uint      `gorm:"primaryKey"`
	LinkID    uint      `gorm:"index"`
	ClickedAt time.Time `gorm:"index"`
}

// Сессия пользователя для простой авторизации через cookie.
type Session struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	Token     string    `gorm:"uniqueIndex"`
	CreatedAt time.Time
}

const (
	sessionCookie = "invite_session"
	sessionTTL    = 30 * 24 * time.Hour
)

func main() {
	db := setupDatabase()

	r := gin.Default()
	r.Static("/static", "web/static")

	r.GET("/", func(c *gin.Context) {
		c.File("web/index.html")
	})
	r.GET("/dashboard", func(c *gin.Context) {
		c.File("web/dashboard.html")
	})

	r.POST("/api/register", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
			return
		}
		if strings.TrimSpace(req.Email) == "" || len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email и пароль обязательны, пароль минимум 6 символов"})
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя"})
			return
		}

		user := User{Email: strings.ToLower(req.Email), PasswordHash: string(passwordHash)}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь уже существует"})
			return
		}

		setSessionCookie(c, db, user.ID)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/api/login", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
			return
		}

		var user User
		if err := db.Where("email = ?", strings.ToLower(req.Email)).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные"})
			return
		}

		setSessionCookie(c, db, user.ID)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/api", authMiddleware(db))
	auth.GET("/me", func(c *gin.Context) {
		user := c.MustGet("user").(User)
		c.JSON(http.StatusOK, gin.H{"email": user.Email})
	})

	auth.GET("/links", func(c *gin.Context) {
		user := c.MustGet("user").(User)
		var links []InviteLink
		if err := db.Where("user_id = ?", user.ID).Order("created_at desc").Find(&links).Error; err != nil {
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
	})

	auth.POST("/links", func(c *gin.Context) {
		user := c.MustGet("user").(User)
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

		link := InviteLink{
			UserID:    user.ID,
			Code:      code,
			TargetURL: req.TargetURL,
		}
		if err := db.Create(&link).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить ссылку"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": link.Code, "id": link.ID})
	})

	auth.GET("/links/:id/stats", func(c *gin.Context) {
		user := c.MustGet("user").(User)
		var link InviteLink
		if err := db.Where("id = ? AND user_id = ?", c.Param("id"), user.ID).First(&link).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ссылка не найдена"})
			return
		}

		var clicks []Click
		if err := db.Where("link_id = ?", link.ID).Order("clicked_at desc").Find(&clicks).Error; err != nil {
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
	})

	r.GET("/i/:code", func(c *gin.Context) {
		var link InviteLink
		if err := db.Where("code = ?", c.Param("code")).First(&link).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ссылка не найдена"})
			return
		}

		click := Click{LinkID: link.ID, ClickedAt: time.Now()}
		_ = db.Create(&click).Error

		c.Redirect(http.StatusFound, link.TargetURL)
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Status(http.StatusNotFound)
			return
		}
		c.File(filepath.Join("web", "index.html"))
	})

	_ = r.Run(":8080")
}

func setupDatabase() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("invite.db"), &gorm.Config{})
	if err != nil {
		panic("не удалось открыть базу данных")
	}

	if err := db.AutoMigrate(&User{}, &InviteLink{}, &Click{}, &Session{}); err != nil {
		panic("не удалось мигрировать базу данных")
	}

	return db
}

func generateCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func authMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := userFromSession(c, db)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func userFromSession(c *gin.Context, db *gorm.DB) (User, error) {
	token, err := c.Cookie(sessionCookie)
	if err != nil || token == "" {
		return User{}, errors.New("no session")
	}

	var session Session
	if err := db.Where("token = ?", token).First(&session).Error; err != nil {
		return User{}, err
	}
	if time.Since(session.CreatedAt) > sessionTTL {
		return User{}, errors.New("session expired")
	}

	var user User
	if err := db.First(&user, session.UserID).Error; err != nil {
		return User{}, err
	}

	return user, nil
}

func setSessionCookie(c *gin.Context, db *gorm.DB, userID uint) {
	token, _ := generateCode(32)
	session := Session{UserID: userID, Token: token}
	_ = db.Create(&session).Error
	c.SetCookie(sessionCookie, token, int(sessionTTL.Seconds()), "/", "", false, true)
}
