package auth

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// AuthService defines the interface for authentication operations (just the relevant methods for email verification)
type AuthService interface {
	VerifyEmail(token string) error
	ResendVerificationEmail(email string) error
}

// VerifyEmail confirms a user's email address
func VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.Error(c, http.StatusBadRequest, "No verification token provided", nil)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	if err := authService.VerifyEmail(token); err != nil {
		response.Error(c, http.StatusBadRequest, "Email verification failed", err)
		return
	}

	response.Success(c, http.StatusOK, "Email verified successfully", nil)
}

// ResendVerificationEmail sends a new verification email
func ResendVerificationEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	if err := authService.ResendVerificationEmail(req.Email); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to resend verification email", err)
		return
	}

	response.Success(c, http.StatusOK, "Verification email sent successfully", nil)
}
