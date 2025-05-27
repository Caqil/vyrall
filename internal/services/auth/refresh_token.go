package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/Caqil/vyrall/internal/pkg/config"
	"github.com/Caqil/vyrall/internal/pkg/errors"
)

// RefreshTokenService handles refresh token operations
type RefreshTokenService struct {
	config *config.JWTConfig
}

// NewRefreshTokenService creates a new refresh token service
func NewRefreshTokenService(config *config.JWTConfig) *RefreshTokenService {
	return &RefreshTokenService{
		config: config,
	}
}

// GenerateRefreshToken generates a new refresh token
func (s *RefreshTokenService) GenerateRefreshToken() (string, error) {
	// Generate random bytes
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", errors.Wrap(err, "Failed to generate random bytes")
	}

	// Encode to base64
	token := base64.URLEncoding.EncodeToString(b)

	return token, nil
}

// ValidateRefreshToken validates a refresh token
func (s *RefreshTokenService) ValidateRefreshToken(token string) (bool, error) {
	// Refresh tokens are validated against the database in the auth service
	// This method is more of a placeholder for any additional validation logic

	// Check if token is empty
	if token == "" {
		return false, errors.New(errors.CodeInvalidToken, "Refresh token is empty")
	}

	// Decode base64 to check format
	_, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false, errors.New(errors.CodeInvalidToken, "Invalid refresh token format")
	}

	return true, nil
}

// GetExpirationTime returns the expiration time for refresh tokens
func (s *RefreshTokenService) GetExpirationTime() time.Time {
	return time.Now().Add(time.Duration(s.config.RefreshExpirationDays) * 24 * time.Hour)
}
