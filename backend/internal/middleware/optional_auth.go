package middleware

import (
	"english-learning/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// OptionalAuth middleware для опциональной проверки JWT токена
// Если токен есть и валиден - устанавливает user_id и user_role в контекст
// Если токена нет или он невалиден - просто продолжает выполнение
func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Нет токена - продолжаем без авторизации
			c.Next()
			return
		}

		// Формат: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Неверный формат - продолжаем без авторизации
			c.Next()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateToken(token, jwtSecret)
		if err != nil {
			// Токен невалиден - продолжаем без авторизации
			c.Next()
			return
		}

		// Токен валиден - сохраняем данные пользователя в контекст
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}
