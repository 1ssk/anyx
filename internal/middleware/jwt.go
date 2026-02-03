package middleware

import (
	"net/http"
	"strings"

	"anyx/internal/models"
	"anyx/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// JWTAuth проверяет JWT и сохраняет user_id в контексте.
func JWTAuth(jwtService *services.JWTService, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен"})
			c.Abort()
			return
		}

		claims, err := jwtService.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			c.Abort()
			return
		}

		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Next()
	}
}
