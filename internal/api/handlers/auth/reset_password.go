package auth

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthService defines the interface for authentication operations (just the relevant methods for password reset)
type AuthService interface {
	ResetPassword(token, newPassword string) error
	ChangePassword(userID primitive.ObjectID, oldPassword, newPassword string) error
}

// PasswordResetRequest represents a password reset request body
type PasswordResetRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// PasswordChangeRequest represents a password change request body
type PasswordChangeRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPassword resets a user's password using a token
func ResetPassword(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	if err := authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		response.Error(c, http.StatusBadRequest, "Password reset failed", err)
		return
	}

	response.Success(c, http.StatusOK, "Password reset successful", nil)
}

// ChangePassword changes a user's password when they know their current password
func ChangePassword(c *gin.Context) {
	var req PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from the authenticated context
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	if err := authService.ChangePassword(userID.(primitive.ObjectID), req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, http.StatusBadRequest, "Password change failed", err)
		return
	}

	response.Success(c, http.StatusOK, "Password changed successfully", nil)
}
