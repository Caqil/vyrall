package auth

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// AuthService defines the interface for authentication operations (just the relevant methods for password recovery)
type AuthService interface {
	ForgotPassword(email string) error
}

// ForgotPassword initiates the password reset process
func ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	if err := authService.ForgotPassword(req.Email); err != nil {
		// Always return success to prevent email enumeration attacks
		// Log the error internally but don't expose it
	}

	response.Success(c, http.StatusOK, "Password reset instructions sent to email if account exists", nil)
}
