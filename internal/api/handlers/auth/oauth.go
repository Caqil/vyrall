package auth

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// AuthService defines the interface for authentication operations (just the relevant methods for OAuth)
type AuthService interface {
	GetOAuthURL(provider string, redirectURI string) (string, error)
	HandleOAuthCallback(provider string, code string) (*models.Session, error)
}

// GetOAuthProviders returns available OAuth providers
func GetOAuthProviders(c *gin.Context) {
	providers := []string{"google", "facebook", "twitter", "apple"}

	response.Success(c, http.StatusOK, "OAuth providers retrieved successfully", gin.H{
		"providers": providers,
	})
}

// InitiateOAuth starts the OAuth flow for a provider
func InitiateOAuth(c *gin.Context) {
	provider := c.Param("provider")
	redirectURI := c.Query("redirect_uri")

	if redirectURI == "" {
		response.Error(c, http.StatusBadRequest, "No redirect URI provided", nil)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	authURL, err := authService.GetOAuthURL(provider, redirectURI)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to initiate OAuth flow", err)
		return
	}

	response.Success(c, http.StatusOK, "OAuth flow initiated", gin.H{
		"auth_url": authURL,
	})
}

// HandleOAuthCallback processes the OAuth callback
func HandleOAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")

	if code == "" {
		error := c.Query("error")
		errorDescription := c.Query("error_description")

		if error != "" {
			response.Error(c, http.StatusBadRequest, "OAuth error: "+error+". "+errorDescription, nil)
		} else {
			response.Error(c, http.StatusBadRequest, "No authorization code provided", nil)
		}
		return
	}

	authService := c.MustGet("authService").(AuthService)

	session, err := authService.HandleOAuthCallback(provider, code)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to process OAuth callback", err)
		return
	}

	response.Success(c, http.StatusOK, "OAuth authentication successful", gin.H{
		"token":         session.Token,
		"refresh_token": session.RefreshToken,
		"expires_at":    session.ExpiresAt,
		"user_id":       session.UserID.Hex(),
	})
}
