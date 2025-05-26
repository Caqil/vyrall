package auth

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthService defines the interface for authentication operations (just the relevant methods for 2FA)
type AuthService interface {
	EnableTwoFactor(userID primitive.ObjectID) (string, error)
	DisableTwoFactor(userID primitive.ObjectID, code string) error
	VerifyTwoFactor(userID primitive.ObjectID, code string) error
	GenerateBackupCodes(userID primitive.ObjectID) ([]string, error)
}

// TwoFactorRequest represents a 2FA code verification request
type TwoFactorRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// EnableTwoFactor enables two-factor authentication for a user
func EnableTwoFactor(c *gin.Context) {
	// Get user ID from the authenticated context
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	secretKey, err := authService.EnableTwoFactor(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to enable two-factor authentication", err)
		return
	}

	response.Success(c, http.StatusOK, "Two-factor authentication setup initiated", gin.H{
		"secret_key": secretKey,
		"qr_code_url": "otpauth://totp/Vyrall:" + userID.(primitive.ObjectID).Hex() +
			"?secret=" + secretKey + "&issuer=Vyrall",
	})
}

// VerifyTwoFactorSetup verifies the 2FA code during setup
func VerifyTwoFactorSetup(c *gin.Context) {
	var req TwoFactorRequest
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

	if err := authService.VerifyTwoFactor(userID.(primitive.ObjectID), req.Code); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid verification code", err)
		return
	}

	// Generate backup codes
	backupCodes, err := authService.GenerateBackupCodes(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate backup codes", err)
		return
	}

	response.Success(c, http.StatusOK, "Two-factor authentication enabled successfully", gin.H{
		"backup_codes": backupCodes,
	})
}

// DisableTwoFactor disables two-factor authentication for a user
func DisableTwoFactor(c *gin.Context) {
	var req TwoFactorRequest
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

	if err := authService.DisableTwoFactor(userID.(primitive.ObjectID), req.Code); err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to disable two-factor authentication", err)
		return
	}

	response.Success(c, http.StatusOK, "Two-factor authentication disabled successfully", nil)
}
