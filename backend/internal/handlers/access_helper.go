package handlers

import (
	"english-learning/internal/models"
	"english-learning/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getRequestUserID(c *gin.Context) (*uuid.UUID, error) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return nil, nil
	}

	userID, err := uuid.Parse(userIDValue.(string))
	if err != nil {
		return nil, err
	}

	return &userID, nil
}

func ensureCourseAccess(c *gin.Context, courseAccessService *services.CourseAccessService, course *models.Course) bool {
	if isAdminRequest(c) {
		return true
	}

	userID, err := getRequestUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return false
	}

	if userID == nil || !course.IsPublished {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return false
	}

	hasAccess, err := courseAccessService.UserHasAccess(*userID, course.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	if !hasAccess {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return false
	}

	return true
}
