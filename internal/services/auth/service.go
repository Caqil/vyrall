package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/config"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// Service represents the authentication service interface
type Service interface {
	// Authentication
	Login(ctx context.Context, email, password, userAgent, ipAddress string) (*models.Session, error)
	Register(ctx context.Context, user *models.User, password string) (*models.User, error)
	Logout(ctx context.Context, sessionID string) error

	// Token management
	ValidateToken(ctx context.Context, token string) (*primitive.ObjectID, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)

	// OAuth functions
	GetOAuthURL(ctx context.Context, provider, redirectURL string) (string, error)
	HandleOAuthCallback(ctx context.Context, provider, code, userAgent, ipAddress string) (*models.Session, error)
	ListOAuthProviders() []string

	// Password management
	ChangePassword(ctx context.Context, userID primitive.ObjectID, oldPassword, newPassword string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	CheckPasswordStrength(password string) (bool, string)

	// Two-factor authentication
	EnableTwoFactor(ctx context.Context, userID primitive.ObjectID) (string, error)
	VerifyTwoFactor(ctx context.Context, userID primitive.ObjectID, code string) error
	DisableTwoFactor(ctx context.Context, userID primitive.ObjectID, code string) error
	TwoFactorRecoveryCodes(ctx context.Context, userID primitive.ObjectID) ([]string, error)

	// Session management
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	GetUserSessions(ctx context.Context, userID primitive.ObjectID) ([]models.Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID primitive.ObjectID) error
	ExtendSession(ctx context.Context, sessionID string) error

	// Account management
	VerifyEmail(ctx context.Context, token string) error
	RequestEmailVerification(ctx context.Context, userID primitive.ObjectID) error
	UpdateUserStatus(ctx context.Context, userID primitive.ObjectID, status string) error
	CheckUserExists(ctx context.Context, email string) (bool, error)
}

// AuthService implements the Service interface
type AuthService struct {
	jwt              *JWTService
	oauth            *OAuthService
	password         *PasswordService
	refresh          *RefreshTokenService
	session          *SessionService
	twoFactor        *TwoFactorService
	userRepo         UserRepository
	sessionRepo      SessionRepository
	verificationRepo VerificationRepository
	config           *config.AuthConfig
	logger           logging.Logger
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(
	jwt *JWTService,
	oauth *OAuthService,
	password *PasswordService,
	refresh *RefreshTokenService,
	session *SessionService,
	twoFactor *TwoFactorService,
	userRepo UserRepository,
	sessionRepo SessionRepository,
	verificationRepo VerificationRepository,
	config *config.AuthConfig,
	logger logging.Logger,
) Service {
	return &AuthService{
		jwt:              jwt,
		oauth:            oauth,
		password:         password,
		refresh:          refresh,
		session:          session,
		twoFactor:        twoFactor,
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		verificationRepo: verificationRepo,
		config:           config,
		logger:           logger,
	}
}

// Login authenticates a user and creates a new session
func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (*models.Session, error) {
	s.logger.Info("Login attempt", "email", email, "ip", ipAddress)

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("Login failed: user not found", "email", email)
		return nil, errors.New(errors.CodeInvalidCredentials, "Invalid email or password")
	}

	// Check if account is active
	if user.Status != "active" {
		s.logger.Warn("Login attempt for inactive account", "email", email, "status", user.Status)
		return nil, errors.New(errors.CodeAccountDisabled, "Account is not active")
	}

	// Verify password
	if valid, err := s.password.VerifyPassword(password, user.PasswordHash); err != nil || !valid {
		s.logger.Warn("Login failed: invalid password", "email", email)
		return nil, errors.New(errors.CodeInvalidCredentials, "Invalid email or password")
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		s.logger.Info("Creating 2FA pending session", "email", email)
		return s.session.CreatePendingSession(ctx, user.ID, userAgent, ipAddress)
	}

	// Create regular session
	s.logger.Info("User logged in successfully", "userId", user.ID.Hex())
	return s.session.CreateSession(ctx, user.ID, userAgent, ipAddress)
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, user *models.User, password string) (*models.User, error) {
	// Validate inputs
	if user.Email == "" || user.Username == "" || password == "" {
		return nil, errors.New(errors.CodeInvalidArgument, "Email, username and password are required")
	}

	// Check password strength
	if strong, msg := s.CheckPasswordStrength(password); !strong {
		return nil, errors.New(errors.CodeInvalidArgument, msg)
	}

	// Check if email already exists
	if existing, _ := s.userRepo.FindByEmail(ctx, user.Email); existing != nil {
		return nil, errors.New(errors.CodeDuplicateEntity, "Email already in use")
	}

	// Check if username already exists
	if existing, _ := s.userRepo.FindByUsername(ctx, user.Username); existing != nil {
		return nil, errors.New(errors.CodeDuplicateEntity, "Username already in use")
	}

	// Hash password
	passwordHash, err := s.password.HashPassword(password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, errors.Wrap(err, "Failed to hash password")
	}

	// Prepare user data
	now := time.Now()
	user.PasswordHash = passwordHash
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Status = "active"
	user.Role = "user"
	user.EmailVerified = false
	user.TwoFactorEnabled = false
	user.FollowerCount = 0
	user.FollowingCount = 0
	user.PostCount = 0

	// Set display name if not provided
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}

	// Create user
	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, errors.Wrap(err, "Failed to create user")
	}

	// Send verification email
	s.RequestEmailVerification(ctx, createdUser.ID)

	s.logger.Info("User registered successfully", "userId", createdUser.ID.Hex())
	return createdUser, nil
}

// Logout invalidates a user session
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	id, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return errors.New(errors.CodeInvalidArgument, "Invalid session ID")
	}

	// Get session before deleting to log the user ID
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err == nil && session != nil {
		s.logger.Info("User logged out", "userId", session.UserID.Hex(), "sessionId", sessionID)
	}

	return s.sessionRepo.Delete(ctx, id)
}

// ValidateToken validates a JWT token and returns the user ID
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*primitive.ObjectID, error) {
	// Validate token structure
	if token == "" {
		return nil, errors.New(errors.CodeInvalidToken, "Token is empty")
	}

	userID, err := s.jwt.ValidateToken(token)
	if err != nil {
		s.logger.Warn("Token validation failed", "error", err)
		return nil, err
	}

	// Check if user exists and is active
	user, err := s.userRepo.FindByID(ctx, *userID)
	if err != nil {
		s.logger.Warn("Token validation failed: user not found", "userId", userID.Hex())
		return nil, errors.New(errors.CodeInvalidToken, "Invalid token")
	}

	if user.Status != "active" {
		s.logger.Warn("Token validation failed: user inactive", "userId", userID.Hex(), "status", user.Status)
		return nil, errors.New(errors.CodeAccountDisabled, "Account is not active")
	}

	return userID, nil
}

// RefreshToken generates a new access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	// Find session by refresh token
	session, err := s.sessionRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		s.logger.Warn("Invalid refresh token", "error", err)
		return nil, errors.New(errors.CodeInvalidToken, "Invalid refresh token")
	}

	// Check if session is active and not expired
	if !session.IsActive || time.Now().After(session.ExpiresAt) {
		session.IsActive = false
		s.sessionRepo.Update(ctx, session)
		s.logger.Warn("Expired refresh token", "sessionId", session.ID.Hex())
		return nil, errors.New(errors.CodeInvalidToken, "Refresh token expired")
	}

	// Generate new tokens
	token, err := s.jwt.GenerateToken(session.UserID)
	if err != nil {
		s.logger.Error("Failed to generate token", "error", err, "userId", session.UserID.Hex())
		return nil, errors.Wrap(err, "Failed to generate token")
	}

	newRefreshToken, err := s.refresh.GenerateRefreshToken()
	if err != nil {
		s.logger.Error("Failed to generate refresh token", "error", err, "userId", session.UserID.Hex())
		return nil, errors.Wrap(err, "Failed to generate refresh token")
	}

	// Update session
	now := time.Now()
	session.Token = token
	session.RefreshToken = newRefreshToken
	session.LastUsedAt = now
	session.ExpiresAt = s.refresh.GetExpirationTime()

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		s.logger.Error("Failed to update session", "error", err, "sessionId", session.ID.Hex())
		return nil, errors.Wrap(err, "Failed to update session")
	}

	s.logger.Info("Token refreshed successfully", "userId", session.UserID.Hex(), "sessionId", session.ID.Hex())
	return session, nil
}

// GetOAuthURL returns the URL for OAuth authentication
func (s *AuthService) GetOAuthURL(ctx context.Context, provider, redirectURL string) (string, error) {
	// Validate provider
	provider = strings.ToLower(provider)
	if !s.isValidOAuthProvider(provider) {
		return "", errors.New(errors.CodeInvalidArgument, fmt.Sprintf("Unsupported OAuth provider: %s", provider))
	}

	s.logger.Info("Getting OAuth URL", "provider", provider)
	return s.oauth.GetAuthURL(provider, redirectURL)
}

// isValidOAuthProvider checks if the provider is supported
func (s *AuthService) isValidOAuthProvider(provider string) bool {
	providers := s.oauth.GetSupportedProviders()
	for _, p := range providers {
		if p == provider {
			return true
		}
	}
	return false
}

// HandleOAuthCallback processes OAuth callback and creates a session
func (s *AuthService) HandleOAuthCallback(ctx context.Context, provider, code, userAgent, ipAddress string) (*models.Session, error) {
	// Validate provider
	provider = strings.ToLower(provider)
	if !s.isValidOAuthProvider(provider) {
		return nil, errors.New(errors.CodeInvalidArgument, fmt.Sprintf("Unsupported OAuth provider: %s", provider))
	}

	// Exchange code for user info
	userInfo, err := s.oauth.ExchangeCode(ctx, provider, code)
	if err != nil {
		s.logger.Error("Failed to exchange OAuth code", "error", err, "provider", provider)
		return nil, errors.Wrap(err, "Failed to exchange OAuth code")
	}

	s.logger.Info("OAuth code exchanged", "provider", provider, "email", userInfo.Email)

	// Find or create user
	user, err := s.userRepo.FindByEmail(ctx, userInfo.Email)
	if err != nil {
		// Create new user
		now := time.Now()

		// Generate unique username if not provided
		username := userInfo.Username
		if username == "" {
			username = s.generateUsername(userInfo.Email, userInfo.Name)
		}

		newUser := &models.User{
			Email:          userInfo.Email,
			Username:       username,
			DisplayName:    userInfo.Name,
			EmailVerified:  true, // OAuth emails are typically verified
			CreatedAt:      now,
			UpdatedAt:      now,
			Status:         "active",
			Role:           "user",
			FollowerCount:  0,
			FollowingCount: 0,
			PostCount:      0,
		}

		user, err = s.userRepo.Create(ctx, newUser)
		if err != nil {
			s.logger.Error("Failed to create user from OAuth", "error", err, "email", userInfo.Email)
			return nil, errors.Wrap(err, "Failed to create user")
		}

		s.logger.Info("Created new user from OAuth", "userId", user.ID.Hex(), "provider", provider)
	} else {
		s.logger.Info("Found existing user from OAuth", "userId", user.ID.Hex(), "provider", provider)
	}

	// Create session
	return s.session.CreateSession(ctx, user.ID, userAgent, ipAddress)
}

// generateUsername creates a unique username from email or name
func (s *AuthService) generateUsername(email, name string) string {
	var username string

	// Try to use name first
	if name != "" {
		// Remove spaces and special characters
		username = strings.ToLower(name)
		username = strings.ReplaceAll(username, " ", "")
		// TODO: Add more character replacements if needed
	} else {
		// Use email
		parts := strings.Split(email, "@")
		username = parts[0]
	}

	// Check if username exists
	_, err := s.userRepo.FindByUsername(context.Background(), username)
	if err == nil {
		// Username exists, add random suffix
		username = fmt.Sprintf("%s%d", username, time.Now().Unix()%1000)
	}

	return username
}

// ListOAuthProviders returns a list of supported OAuth providers
func (s *AuthService) ListOAuthProviders() []string {
	return s.oauth.GetSupportedProviders()
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID primitive.ObjectID, oldPassword, newPassword string) error {
	// Check password strength
	if strong, msg := s.CheckPasswordStrength(newPassword); !strong {
		return errors.New(errors.CodeInvalidArgument, msg)
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Warn("Change password failed: user not found", "userId", userID.Hex())
		return errors.Wrap(err, "Failed to find user")
	}

	// Verify old password
	if valid, err := s.password.VerifyPassword(oldPassword, user.PasswordHash); err != nil || !valid {
		s.logger.Warn("Change password failed: incorrect password", "userId", userID.Hex())
		return errors.New(errors.CodeInvalidCredentials, "Current password is incorrect")
	}

	// Hash new password
	passwordHash, err := s.password.HashPassword(newPassword)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err, "userId", userID.Hex())
		return errors.Wrap(err, "Failed to hash password")
	}

	// Update user
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to update user password", "error", err, "userId", userID.Hex())
		return errors.Wrap(err, "Failed to update password")
	}

	// Revoke all other sessions
	err = s.sessionRepo.DeleteByUserID(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to revoke sessions after password change", "error", err, "userId", userID.Hex())
	}

	s.logger.Info("Password changed successfully", "userId", userID.Hex())
	return nil
}

// RequestPasswordReset initiates a password reset process
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	return s.password.RequestReset(ctx, email)
}

// ResetPassword resets a user's password using a token
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Check password strength
	if strong, msg := s.CheckPasswordStrength(newPassword); !strong {
		return errors.New(errors.CodeInvalidArgument, msg)
	}

	return s.password.ResetPassword(ctx, token, newPassword)
}

// CheckPasswordStrength checks if a password meets security requirements
func (s *AuthService) CheckPasswordStrength(password string) (bool, string) {
	return s.password.CheckStrength(password)
}

// EnableTwoFactor enables 2FA for a user
func (s *AuthService) EnableTwoFactor(ctx context.Context, userID primitive.ObjectID) (string, error) {
	return s.twoFactor.Enable(ctx, userID)
}

// VerifyTwoFactor verifies a 2FA code
func (s *AuthService) VerifyTwoFactor(ctx context.Context, userID primitive.ObjectID, code string) error {
	err := s.twoFactor.Verify(ctx, userID, code)
	if err != nil {
		return err
	}

	// If verification is for a pending session, complete it
	// This would be handled in the controller by creating a full session after verification

	return nil
}

// DisableTwoFactor disables 2FA for a user
func (s *AuthService) DisableTwoFactor(ctx context.Context, userID primitive.ObjectID, code string) error {
	return s.twoFactor.Disable(ctx, userID, code)
}

// TwoFactorRecoveryCodes returns recovery codes for a user
func (s *AuthService) TwoFactorRecoveryCodes(ctx context.Context, userID primitive.ObjectID) ([]string, error) {
	return s.twoFactor.GetRecoveryCodes(ctx, userID)
}

// GetSession retrieves a session by ID
func (s *AuthService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	id, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, errors.New(errors.CodeInvalidArgument, "Invalid session ID")
	}

	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find session")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		session.IsActive = false
		s.sessionRepo.Update(ctx, session)
		return nil, errors.New(errors.CodeInvalidToken, "Session expired")
	}

	return session, nil
}

// GetUserSessions retrieves all sessions for a user
func (s *AuthService) GetUserSessions(ctx context.Context, userID primitive.ObjectID) ([]models.Session, error) {
	sessions, err := s.sessionRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find user sessions", "error", err, "userId", userID.Hex())
		return nil, errors.Wrap(err, "Failed to find sessions")
	}

	// Filter out expired sessions
	activeSessions := []models.Session{}
	now := time.Now()
	for _, session := range sessions {
		if session.IsActive && now.Before(session.ExpiresAt) {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions, nil
}

// RevokeSession revokes a specific session
func (s *AuthService) RevokeSession(ctx context.Context, sessionID string) error {
	id, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return errors.New(errors.CodeInvalidArgument, "Invalid session ID")
	}

	session, err := s.sessionRepo.FindByID(ctx, id)
	if err == nil {
		s.logger.Info("Session revoked", "sessionId", sessionID, "userId", session.UserID.Hex())
	}

	return s.sessionRepo.Delete(ctx, id)
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *AuthService) RevokeAllUserSessions(ctx context.Context, userID primitive.ObjectID) error {
	s.logger.Info("Revoking all sessions", "userId", userID.Hex())
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

// ExtendSession extends the expiration time of a session
func (s *AuthService) ExtendSession(ctx context.Context, sessionID string) error {
	id, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return errors.New(errors.CodeInvalidArgument, "Invalid session ID")
	}

	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "Failed to find session")
	}

	// Check if session is active
	if !session.IsActive {
		return errors.New(errors.CodeInvalidToken, "Session is inactive")
	}

	// Update session expiration
	now := time.Now()
	session.LastUsedAt = now
	session.ExpiresAt = s.refresh.GetExpirationTime()

	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		s.logger.Error("Failed to extend session", "error", err, "sessionId", sessionID)
		return errors.Wrap(err, "Failed to extend session")
	}

	s.logger.Info("Session extended", "sessionId", sessionID, "userId", session.UserID.Hex())
	return nil
}

// VerifyEmail verifies a user's email using a token
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// Find verification record
	verification, err := s.verificationRepo.FindByToken(ctx, token, "email")
	if err != nil {
		return errors.New(errors.CodeInvalidToken, "Invalid verification token")
	}

	// Check if token is expired
	if time.Now().After(verification.ExpiresAt) {
		return errors.New(errors.CodeInvalidToken, "Verification token expired")
	}

	// Update user
	user, err := s.userRepo.FindByID(ctx, verification.UserID)
	if err != nil {
		return errors.Wrap(err, "Failed to find user")
	}

	user.EmailVerified = true
	user.UpdatedAt = time.Now()
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to update user email verification status", "error", err, "userId", user.ID.Hex())
		return errors.Wrap(err, "Failed to update user")
	}

	// Mark verification as completed
	verification.Status = "completed"
	verification.VerifiedAt = timePtr(time.Now())
	err = s.verificationRepo.Update(ctx, verification)
	if err != nil {
		s.logger.Warn("Failed to update verification record", "error", err, "userId", user.ID.Hex())
	}

	s.logger.Info("Email verified successfully", "userId", user.ID.Hex())
	return nil
}

// RequestEmailVerification sends a verification email to a user
func (s *AuthService) RequestEmailVerification(ctx context.Context, userID primitive.ObjectID) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to find user")
	}

	// Check if email is already verified
	if user.EmailVerified {
		return errors.New(errors.CodeInvalidOperation, "Email is already verified")
	}

	// Generate verification token
	token, err := generateSecureToken(32)
	if err != nil {
		return errors.Wrap(err, "Failed to generate verification token")
	}

	// Create verification record
	now := time.Now()
	verification := &models.Verification{
		UserID:           user.ID,
		Type:             "email",
		VerificationCode: token,
		Status:           "pending",
		CreatedAt:        now,
		ExpiresAt:        now.Add(24 * time.Hour), // 24 hours expiry
	}

	// Save verification record
	_, err = s.verificationRepo.Create(ctx, verification)
	if err != nil {
		return errors.Wrap(err, "Failed to create verification record")
	}

	// Send verification email
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", s.config.AppURL, token)

	emailData := map[string]interface{}{
		"Username":         user.Username,
		"VerificationLink": verificationLink,
	}

	// This would call an email service
	// s.emailService.SendTemplatedEmail(user.Email, "Verify Your Email", "email_verification", emailData)

	s.logger.Info("Email verification requested", "userId", user.ID.Hex())
	return nil
}

// UpdateUserStatus updates a user's status
func (s *AuthService) UpdateUserStatus(ctx context.Context, userID primitive.ObjectID, status string) error {
	// Validate status
	validStatuses := []string{"active", "inactive", "suspended", "deleted"}
	isValid := false
	for _, s := range validStatuses {
		if status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.New(errors.CodeInvalidArgument, "Invalid status")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to find user")
	}

	// Update status
	user.Status = status
	user.UpdatedAt = time.Now()
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to update user status", "error", err, "userId", userID.Hex())
		return errors.Wrap(err, "Failed to update user")
	}

	// If status is suspended or deleted, revoke all sessions
	if status == "suspended" || status == "deleted" {
		err = s.sessionRepo.DeleteByUserID(ctx, userID)
		if err != nil {
			s.logger.Warn("Failed to revoke sessions after status change", "error", err, "userId", userID.Hex())
		}
	}

	s.logger.Info("User status updated", "userId", userID.Hex(), "status", status)
	return nil
}

// CheckUserExists checks if a user with the given email exists
func (s *AuthService) CheckUserExists(ctx context.Context, email string) (bool, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Check if error is "not found"
		if errors.Code(err) == errors.CodeNotFound {
			return false, nil
		}
		return false, errors.Wrap(err, "Failed to check user existence")
	}

	return user != nil, nil
}

// Helper functions

func timePtr(t time.Time) *time.Time {
	return &t
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// VerificationRepository handles verification data access
type VerificationRepository interface {
	Create(ctx context.Context, verification *models.Verification) (*models.Verification, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Verification, error)
	FindByToken(ctx context.Context, token, type_ string) (*models.Verification, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID, type_ string) (*models.Verification, error)
	Update(ctx context.Context, verification *models.Verification) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

// UserRepository defines the operations needed for user data access
type UserRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

// SessionRepository defines the operations needed for session data access
type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) (*models.Session, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Session, error)
	FindByToken(ctx context.Context, token string) (*models.Session, error)
	FindByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]models.Session, error)
	Update(ctx context.Context, session *models.Session) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
	DeleteExpired(ctx context.Context) error
}
