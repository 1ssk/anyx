package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminOnly проверяет роль администратора.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("user_role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
			c.Abort()
			return
		}
		c.Next()
	}
}
