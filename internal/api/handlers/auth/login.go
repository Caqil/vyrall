package auth

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// AuthService defines the interface for authentication operations (just the relevant methods for login)
type AuthService interface {
	Login(email, password string) (*models.Session, error)
	ValidateToken(token string) (*models.User, error)
	VerifyTwoFactor(userID primitive.ObjectID, code string) error
}

// LoginRequest represents a login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login authenticates a user and returns a session token
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	session, err := authService.Login(req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Authentication failed", err)
		return
	}

	// Check if two-factor authentication is required
	if session.UserID != primitive.NilObjectID && session.Token == "" {
		response.Success(c, http.StatusOK, "Two-factor authentication required", gin.H{
			"two_factor_required": true,
			"user_id":             session.UserID.Hex(),
		})
		return
	}

	response.Success(c, http.StatusOK, "Login successful", gin.H{
		"token":         session.Token,
		"refresh_token": session.RefreshToken,
		"expires_at":    session.ExpiresAt,
		"user_id":       session.UserID.Hex(),
	})
}

// VerifyTwoFactorLogin verifies the 2FA code during login
func VerifyTwoFactorLogin(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Code   string `json:"code" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	// Verify the 2FA code
	if err := authService.VerifyTwoFactor(userID, req.Code); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid verification code", err)
		return
	}

	// Generate a session now that 2FA is verified
	session, err := authService.Login("", "") // Using empty credentials here as 2FA is already verified
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create session after 2FA verification", err)
		return
	}

	response.Success(c, http.StatusOK, "Two-factor authentication verified", gin.H{
		"token":         session.Token,
		"refresh_token": session.RefreshToken,
		"expires_at":    session.ExpiresAt,
		"user_id":       session.UserID.Hex(),
	})
}
