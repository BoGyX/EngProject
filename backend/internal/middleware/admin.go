package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Admin middleware для проверки, что пользователь является администратором
// Используется после Auth middleware, который устанавливает user_role в контекст
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
