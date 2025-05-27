package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/config"
	"github.com/Caqil/vyrall/internal/pkg/errors"
)

// PasswordService handles password operations
type PasswordService struct {
	userRepo         UserRepository
	verificationRepo VerificationRepository
	emailService     EmailService
	config           *config.PasswordConfig
}

// NewPasswordService creates a new password service
func NewPasswordService(
	userRepo UserRepository,
	verificationRepo VerificationRepository,
	emailService EmailService,
	config *config.PasswordConfig,
) *PasswordService {
	return &PasswordService{
		userRepo:         userRepo,
		verificationRepo: verificationRepo,
		emailService:     emailService,
		config:           config,
	}
}

// HashPassword hashes a password
func (s *PasswordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New(errors.CodeInvalidArgument, "Password cannot be empty")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "Failed to hash password")
	}

	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against a hash
func (s *PasswordService) VerifyPassword(password, hash string) (bool, error) {
	if password == "" || hash == "" {
		return false, errors.New(errors.CodeInvalidArgument, "Password or hash cannot be empty")
	}

	// Compare the password with the hash
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, errors.Wrap(err, "Failed to verify password")
	}

	return true, nil
}

// RequestReset initiates a password reset
func (s *PasswordService) RequestReset(ctx context.Context, email string) error {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not
		return nil
	}

	// Generate reset token
	token, err := generateSecureToken(32)
	if err != nil {
		return errors.Wrap(err, "Failed to generate reset token")
	}

	// Create verification record
	now := time.Now()
	verification := &models.Verification{
		UserID:           user.ID,
		Type:             "password_reset",
		VerificationCode: token,
		Status:           "pending",
		CreatedAt:        now,
		ExpiresAt:        now.Add(time.Duration(s.config.ResetTokenExpiryHours) * time.Hour),
	}

	// Save verification record
	_, err = s.verificationRepo.Create(ctx, verification)
	if err != nil {
		return errors.Wrap(err, "Failed to create verification record")
	}

	// Send reset email
	resetLink := s.config.ResetBaseURL + "?token=" + token

	emailData := map[string]interface{}{
		"Username":  user.Username,
		"ResetLink": resetLink,
		"ExpiresIn": s.config.ResetTokenExpiryHours,
	}

	err = s.emailService.SendTemplatedEmail(
		user.Email,
		"Reset Your Password",
		"password_reset",
		emailData,
	)

	if err != nil {
		return errors.Wrap(err, "Failed to send reset email")
	}

	return nil
}

// ResetPassword resets a user's password using a token
func (s *PasswordService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Find verification record
	verification, err := s.verificationRepo.FindByToken(ctx, token, "password_reset")
	if err != nil {
		return errors.New(errors.CodeInvalidToken, "Invalid reset token")
	}

	// Check if token is expired
	if time.Now().After(verification.ExpiresAt) {
		return errors.New(errors.CodeInvalidToken, "Reset token expired")
	}

	// Check if token has already been used
	if verification.Status != "pending" {
		return errors.New(errors.CodeInvalidToken, "Reset token already used")
	}

	// Check password strength
	if strong, msg := s.CheckStrength(newPassword); !strong {
		return errors.New(errors.CodeInvalidArgument, msg)
	}

	// Hash new password
	passwordHash, err := s.HashPassword(newPassword)
	if err != nil {
		return errors.Wrap(err, "Failed to hash password")
	}

	// Update user password
	user, err := s.userRepo.FindByID(ctx, verification.UserID)
	if err != nil {
		return errors.Wrap(err, "Failed to find user")
	}

	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return errors.Wrap(err, "Failed to update password")
	}

	// Mark verification record as used
	verification.Status = "used"
	verification.VerifiedAt = timePtr(time.Now())

	err = s.verificationRepo.Update(ctx, verification)
	if err != nil {
		return errors.Wrap(err, "Failed to update verification record")
	}

	// Invalidate all sessions
	// This would typically be handled by the auth service

	return nil
}

// CheckStrength checks if a password meets security requirements
func (s *PasswordService) CheckStrength(password string) (bool, string) {
	if len(password) < s.config.MinLength {
		return false, fmt.Sprintf("Password must be at least %d characters", s.config.MinLength)
	}

	// Check for uppercase
	if s.config.RequireUppercase && !containsUppercase(password) {
		return false, "Password must contain at least one uppercase letter"
	}

	// Check for lowercase
	if s.config.RequireLowercase && !containsLowercase(password) {
		return false, "Password must contain at least one lowercase letter"
	}

	// Check for numbers
	if s.config.RequireNumbers && !containsNumber(password) {
		return false, "Password must contain at least one number"
	}

	// Check for special characters
	if s.config.RequireSpecial && !containsSpecial(password) {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}

// Helper functions

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func containsUppercase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func containsLowercase(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func containsSpecial(s string) bool {
	specials := "!@#$%^&*()-_=+[]{}|;:,.<>?/~`"
	for _, r := range s {
		for _, sr := range specials {
			if r == sr {
				return true
			}
		}
	}
	return false
}

// EmailService interface for sending emails
type EmailService interface {
	SendTemplatedEmail(to, subject, template string, data map[string]interface{}) error
}
