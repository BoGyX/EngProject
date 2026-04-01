package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetGreeting возвращает простое приветствие
// @Summary Получить приветствие
// @Description Простой endpoint который возвращает приветствие
// @Tags greeting
// @Produce json
// @Success 200 {object} map[string]string
// @Router /greeting [get]
func GetGreeting(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Привет!",
		"status":  "ok",
	})
}
