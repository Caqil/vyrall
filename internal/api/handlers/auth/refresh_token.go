package auth

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// AuthService defines the interface for authentication operations (just the relevant methods for token refresh)
type AuthService interface {
	RefreshToken(refreshToken string) (*models.Session, error)
}

// RefreshToken generates a new access token using a refresh token
func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	session, err := authService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	response.Success(c, http.StatusOK, "Token refreshed successfully", gin.H{
		"token":         session.Token,
		"refresh_token": session.RefreshToken,
		"expires_at":    session.ExpiresAt,
	})
}
