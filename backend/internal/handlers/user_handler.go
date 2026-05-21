package handlers

import (
	"english-learning/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService         *services.UserService
	courseAccessService *services.CourseAccessService
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type CreateManagedUserRequest struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	MoodleLogin string `json:"moodle_login" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
}

type UpdateUserCourseAccessRequest struct {
	CourseIDs []int64 `json:"course_ids"`
}

type UserCourseAccessResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	CourseIDs []int64   `json:"course_ids"`
}

func NewUserHandler(userService *services.UserService, courseAccessService *services.CourseAccessService) *UserHandler {
	return &UserHandler{
		userService:         userService,
		courseAccessService: courseAccessService,
	}
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) CreateManagedUser(c *gin.Context) {
	var req CreateManagedUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fullName := strings.TrimSpace(req.FirstName + " " + req.LastName)
	user, err := h.userService.CreateManagedUser(fullName, req.MoodleLogin, req.Password)
	if err != nil {
		switch err.Error() {
		case "user with this email already exists", "user with this Moodle login already exists":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	targetUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	currentUserID, currentUserErr := getRequestUserID(c)
	if currentUserErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid current user ID format"})
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if currentUserID != nil && *currentUserID == targetUserID && req.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot remove your own admin role"})
		return
	}

	if err := h.userService.UpdateUserRole(targetUserID, req.Role); err != nil {
		if err.Error() == "invalid role" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.GetUserByID(targetUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetAllCourseAccesses(c *gin.Context) {
	accesses, err := h.courseAccessService.GetAllCourseAccesses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accesses)
}

func (h *UserHandler) GetUserCourseAccess(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	courseIDs, err := h.courseAccessService.GetAssignedCourseIDsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, UserCourseAccessResponse{
		UserID:    userID,
		CourseIDs: courseIDs,
	})
}

func (h *UserHandler) UpdateUserCourseAccess(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	grantedBy, grantedByErr := getRequestUserID(c)
	if grantedByErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid current user ID format"})
		return
	}

	var req UpdateUserCourseAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	courseIDs, err := h.courseAccessService.ReplaceUserCourseAccess(userID, req.CourseIDs, grantedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, UserCourseAccessResponse{
		UserID:    userID,
		CourseIDs: courseIDs,
	})
}

func (h *UserHandler) GrantCourseAccess(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	courseID, err := strconv.ParseInt(c.Param("course_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID format"})
		return
	}

	grantedBy, grantedByErr := getRequestUserID(c)
	if grantedByErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid current user ID format"})
		return
	}

	access, err := h.courseAccessService.GrantCourseAccess(userID, courseID, grantedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, access)
}

func (h *UserHandler) RevokeCourseAccess(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	courseID, err := strconv.ParseInt(c.Param("course_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID format"})
		return
	}

	if err := h.courseAccessService.RevokeCourseAccess(userID, courseID); err != nil {
		if err.Error() == "course access not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
