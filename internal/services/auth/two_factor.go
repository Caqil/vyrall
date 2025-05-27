package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"image/png"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/config"
	"github.com/Caqil/vyrall/internal/pkg/errors"
)

// TwoFactorService handles two-factor authentication
type TwoFactorService struct {
	userRepo      UserRepository
	twoFactorRepo TwoFactorRepository
	emailService  EmailService
	config        *config.TwoFactorConfig
}

// TwoFactorRepository handles 2FA data storage
type TwoFactorRepository interface {
	SaveSecret(ctx context.Context, userID primitive.ObjectID, secret string) error
	GetSecret(ctx context.Context, userID primitive.ObjectID) (string, error)
	DeleteSecret(ctx context.Context, userID primitive.ObjectID) error
	SaveRecoveryCodes(ctx context.Context, userID primitive.ObjectID, codes []string) error
	GetRecoveryCodes(ctx context.Context, userID primitive.ObjectID) ([]string, error)
	MarkRecoveryCodeUsed(ctx context.Context, userID primitive.ObjectID, code string) error
}

// NewTwoFactorService creates a new two-factor authentication service
func NewTwoFactorService(
	userRepo UserRepository,
	twoFactorRepo TwoFactorRepository,
	emailService EmailService,
	config *config.TwoFactorConfig,
) *TwoFactorService {
	return &TwoFactorService{
		userRepo:      userRepo,
		twoFactorRepo: twoFactorRepo,
		emailService:  emailService,
		config:        config,
	}
}

// Enable enables two-factor authentication for a user
func (s *TwoFactorService) Enable(ctx context.Context, userID primitive.ObjectID) (string, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", errors.Wrap(err, "Failed to find user")
	}

	// Generate secret key
	secretBytes := make([]byte, 20)
	_, err = rand.Read(secretBytes)
	if err != nil {
		return "", errors.Wrap(err, "Failed to generate secret key")
	}

	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes)

	// Save secret
	err = s.twoFactorRepo.SaveSecret(ctx, userID, secret)
	if err != nil {
		return "", errors.Wrap(err, "Failed to save secret")
	}

	// Generate recovery codes
	recoveryCodes, err := s.generateRecoveryCodes(10)
	if err != nil {
		return "", errors.Wrap(err, "Failed to generate recovery codes")
	}

	// Save recovery codes
	err = s.twoFactorRepo.SaveRecoveryCodes(ctx, userID, recoveryCodes)
	if err != nil {
		return "", errors.Wrap(err, "Failed to save recovery codes")
	}

	// Generate QR code
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.config.Issuer,
		AccountName: user.Email,
		Secret:      []byte(secret),
	})
	if err != nil {
		return "", errors.Wrap(err, "Failed to generate OTP key")
	}

	// Return the secret key for QR code generation
	return key.Secret(), nil
}

// GenerateQRCode generates a QR code for the TOTP secret
func (s *TwoFactorService) GenerateQRCode(user *models.User, secret string) ([]byte, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.config.Issuer,
		AccountName: user.Email,
		Secret:      []byte(secret),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate OTP key")
	}

	// Generate QR code
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate QR code")
	}

	err = png.Encode(&buf, img)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encode QR code")
	}

	return buf.Bytes(), nil
}

// Verify verifies a TOTP code
func (s *TwoFactorService) Verify(ctx context.Context, userID primitive.ObjectID, code string) error {
	// Check if it's a recovery code
	if len(code) == 10 && !strings.ContainsAny(code, "oO0iIlL") {
		return s.verifyRecoveryCode(ctx, userID, code)
	}

	// Get secret
	secret, err := s.twoFactorRepo.GetSecret(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to get secret")
	}

	// Verify code
	valid := totp.Validate(code, secret)
	if !valid {
		return errors.New(errors.CodeInvalidCredentials, "Invalid verification code")
	}

	return nil
}

// Disable disables two-factor authentication for a user
func (s *TwoFactorService) Disable(ctx context.Context, userID primitive.ObjectID, code string) error {
	// Verify code first
	err := s.Verify(ctx, userID, code)
	if err != nil {
		return err
	}

	// Delete secret and recovery codes
	err = s.twoFactorRepo.DeleteSecret(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to delete secret")
	}

	// Update user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to find user")
	}

	user.TwoFactorEnabled = false
	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return errors.Wrap(err, "Failed to update user")
	}

	// Notify user
	emailData := map[string]interface{}{
		"Username": user.Username,
		"Time":     time.Now().Format(time.RFC1123),
	}

	s.emailService.SendTemplatedEmail(
		user.Email,
		"Two-Factor Authentication Disabled",
		"2fa_disabled",
		emailData,
	)

	return nil
}

// GetRecoveryCodes returns recovery codes for a user
func (s *TwoFactorService) GetRecoveryCodes(ctx context.Context, userID primitive.ObjectID) ([]string, error) {
	return s.twoFactorRepo.GetRecoveryCodes(ctx, userID)
}

// RegenerateRecoveryCodes generates new recovery codes for a user
func (s *TwoFactorService) RegenerateRecoveryCodes(ctx context.Context, userID primitive.ObjectID) ([]string, error) {
	// Generate new recovery codes
	recoveryCodes, err := s.generateRecoveryCodes(10)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate recovery codes")
	}

	// Save recovery codes
	err = s.twoFactorRepo.SaveRecoveryCodes(ctx, userID, recoveryCodes)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to save recovery codes")
	}

	return recoveryCodes, nil
}

// Helper methods

func (s *TwoFactorService) verifyRecoveryCode(ctx context.Context, userID primitive.ObjectID, code string) error {
	// Get recovery codes
	codes, err := s.twoFactorRepo.GetRecoveryCodes(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to get recovery codes")
	}

	// Check if code exists
	found := false
	for _, c := range codes {
		if c == code {
			found = true
			break
		}
	}

	if !found {
		return errors.New(errors.CodeInvalidCredentials, "Invalid recovery code")
	}

	// Mark code as used
	err = s.twoFactorRepo.MarkRecoveryCodeUsed(ctx, userID, code)
	if err != nil {
		return errors.Wrap(err, "Failed to mark recovery code as used")
	}

	return nil
}

func (s *TwoFactorService) generateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)

	// Characters that are unlikely to be confused with each other
	alphabet := "abcdefghjkmnpqrstuvwxyz23456789"

	for i := 0; i < count; i++ {
		code := make([]byte, 10)
		_, err := rand.Read(code)
		if err != nil {
			return nil, err
		}

		for j := range code {
			code[j] = alphabet[int(code[j])%len(alphabet)]
		}

		codes[i] = string(code)
	}

	return codes, nil
}
