package admin

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserService interface for user management operations
type UserService interface {
	GetUsers(filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.User, int, error)
	GetUserByID(id primitive.ObjectID) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id primitive.ObjectID) error
	SuspendUser(id primitive.ObjectID, reason string, duration time.Duration) error
	UnsuspendUser(id primitive.ObjectID) error
	GetUserActivity(id primitive.ObjectID, limit, offset int) ([]map[string]interface{}, int, error)
	GetUserContent(id primitive.ObjectID, contentType string, limit, offset int) ([]map[string]interface{}, int, error)
	VerifyUser(id primitive.ObjectID) error
	UnverifyUser(id primitive.ObjectID) error
	GetUserVerificationRequests(status string, limit, offset int) ([]*models.VerificationRequest, int, error)
	ProcessVerificationRequest(id primitive.ObjectID, status, reason string) error
	GetUserRoles() ([]map[string]interface{}, error)
	AssignUserRole(userID primitive.ObjectID, role string) error
}

// ListUsers returns a list of users with filtering and pagination
func ListUsers(c *gin.Context) {
	// Get query parameters
	status := c.DefaultQuery("status", "all")
	role := c.DefaultQuery("role", "all")
	verified := c.DefaultQuery("verified", "all")
	query := c.DefaultQuery("query", "")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	limit, offset := getPaginationParams(c)

	userService := c.MustGet("userService").(UserService)

	// Build filter
	filter := make(map[string]interface{})
	if status != "all" {
		filter["status"] = status
	}
	if role != "all" {
		filter["role"] = role
	}
	if verified != "all" {
		isVerified := verified == "true"
		filter["is_verified"] = isVerified
	}
	if query != "" {
		// Search by username, email, or name
		filter["$or"] = []map[string]interface{}{
			{"username": map[string]interface{}{"$regex": query, "$options": "i"}},
			{"email": map[string]interface{}{"$regex": query, "$options": "i"}},
			{"display_name": map[string]interface{}{"$regex": query, "$options": "i"}},
		}
	}

	// Build sort
	sort := make(map[string]int)
	if sortOrder == "desc" {
		sort[sortBy] = -1
	} else {
		sort[sortBy] = 1
	}

	users, total, err := userService.GetUsers(filter, sort, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve users", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Users retrieved successfully", users, limit, offset, total)
}

// GetUserDetail returns details about a specific user
func GetUserDetail(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	userService := c.MustGet("userService").(UserService)

	user, err := userService.GetUserByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user details", err)
		return
	}

	response.Success(c, http.StatusOK, "User details retrieved successfully", user)
}

// UpdateUserProfile updates a user's profile
func UpdateUserProfile(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	userService := c.MustGet("userService").(UserService)

	// Get the existing user
	user, err := userService.GetUserByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user", err)
		return
	}

	// Bind the request body to the user
	if err := c.ShouldBindJSON(user); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Ensure ID doesn't change
	user.ID = id

	// Update the user
	if err := userService.UpdateUser(user); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update user", err)
		return
	}

	response.Success(c, http.StatusOK, "User updated successfully", user)
}

// DeleteUser deletes a user account
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	userService := c.MustGet("userService").(UserService)

	if err := userService.DeleteUser(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	response.Success(c, http.StatusOK, "User deleted successfully", nil)
}

// SuspendUser suspends a user account
func SuspendUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var req struct {
		Reason   string `json:"reason" binding:"required"`
		Duration string `json:"duration" binding:"required"` // e.g., "24h", "7d", "30d"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userService := c.MustGet("userService").(UserService)

	// Parse duration
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid duration format", err)
		return
	}

	if err := userService.SuspendUser(id, req.Reason, duration); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to suspend user", err)
		return
	}

	response.Success(c, http.StatusOK, "User suspended successfully", nil)
}

// UnsuspendUser removes a suspension from a user account
func UnsuspendUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	userService := c.MustGet("userService").(UserService)

	if err := userService.UnsuspendUser(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unsuspend user", err)
		return
	}

	response.Success(c, http.StatusOK, "User unsuspended successfully", nil)
}

// GetUserActivity returns a user's activity history
func GetUserActivity(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	limit, offset := getPaginationParams(c)
	userService := c.MustGet("userService").(UserService)

	activity, total, err := userService.GetUserActivity(id, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user activity", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "User activity retrieved successfully", activity, limit, offset, total)
}

// GetUserContent returns content created by a user
func GetUserContent(c *gin.Context) {
	idStr := c.Param("id")
	contentType := c.DefaultQuery("type", "post")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	limit, offset := getPaginationParams(c)
	userService := c.MustGet("userService").(UserService)

	// Validate content type
	validTypes := map[string]bool{
		"post":        true,
		"comment":     true,
		"story":       true,
		"message":     true,
		"event":       true,
		"group":       true,
		"live_stream": true,
	}

	if !validTypes[contentType] {
		response.Error(c, http.StatusBadRequest, "Invalid content type", nil)
		return
	}

	content, total, err := userService.GetUserContent(id, contentType, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user content", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "User content retrieved successfully", content, limit, offset, total)
}

// VerifyUser marks a user account as verified (blue check)
func VerifyUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	userService := c.MustGet("userService").(UserService)

	if err := userService.VerifyUser(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify user", err)
		return
	}

	response.Success(c, http.StatusOK, "User verified successfully", nil)
}

// UnverifyUser removes verified status from a user account
func UnverifyUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	userService := c.MustGet("userService").(UserService)

	if err := userService.UnverifyUser(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unverify user", err)
		return
	}

	response.Success(c, http.StatusOK, "User unverified successfully", nil)
}

// GetVerificationRequests returns pending verification requests
func GetVerificationRequests(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	limit, offset := getPaginationParams(c)

	userService := c.MustGet("userService").(UserService)

	// Validate status
	validStatuses := map[string]bool{
		"pending":  true,
		"approved": true,
		"rejected": true,
		"all":      true,
	}

	if !validStatuses[status] {
		response.Error(c, http.StatusBadRequest, "Invalid status", nil)
		return
	}

	requests, total, err := userService.GetUserVerificationRequests(status, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve verification requests", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Verification requests retrieved successfully", requests, limit, offset, total)
}

// ProcessVerificationRequest approves or rejects a verification request
func ProcessVerificationRequest(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request ID", err)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userService := c.MustGet("userService").(UserService)

	// Validate status
	if req.Status != "approved" && req.Status != "rejected" {
		response.Error(c, http.StatusBadRequest, "Invalid status. Must be 'approved' or 'rejected'", nil)
		return
	}

	if err := userService.ProcessVerificationRequest(id, req.Status, req.Reason); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to process verification request", err)
		return
	}

	response.Success(c, http.StatusOK, "Verification request processed successfully", nil)
}

// GetUserRoles returns available user roles
func GetUserRoles(c *gin.Context) {
	userService := c.MustGet("userService").(UserService)

	roles, err := userService.GetUserRoles()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user roles", err)
		return
	}

	response.Success(c, http.StatusOK, "User roles retrieved successfully", roles)
}

// AssignUserRole assigns a role to a user
func AssignUserRole(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userService := c.MustGet("userService").(UserService)

	if err := userService.AssignUserRole(id, req.Role); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to assign role to user", err)
		return
	}

	response.Success(c, http.StatusOK, "Role assigned successfully", nil)
}
