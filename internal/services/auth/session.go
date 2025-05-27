package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mssola/user_agent"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/config"
	"github.com/Caqil/vyrall/internal/pkg/errors"
)

// SessionService handles session management
type SessionService struct {
	sessionRepo    SessionRepository
	jwtService     JWTService
	refreshService RefreshTokenService
	geoIPService   GeoIPService
	config         *config.SessionConfig
}

// NewSessionService creates a new session service
func NewSessionService(
	sessionRepo SessionRepository,
	jwtService JWTService,
	refreshService RefreshTokenService,
	geoIPService GeoIPService,
	config *config.SessionConfig,
) *SessionService {
	return &SessionService{
		sessionRepo:    sessionRepo,
		jwtService:     jwtService,
		refreshService: refreshService,
		geoIPService:   geoIPService,
		config:         config,
	}
}

// CreateSession creates a new session
func (s *SessionService) CreateSession(ctx context.Context, userID primitive.ObjectID, userAgent, ipAddress string) (*models.Session, error) {
	// Generate JWT token
	token, err := s.jwtService.GenerateToken(userID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate token")
	}

	// Generate refresh token
	refreshToken, err := s.refreshService.GenerateRefreshToken()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate refresh token")
	}

	// Create session
	now := time.Now()
	session := &models.Session{
		UserID:       userID,
		Token:        token,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		Device:       s.DetectDevice(userAgent),
		Location:     s.GetLocationFromIP(ipAddress),
		CreatedAt:    now,
		LastUsedAt:   now,
		ExpiresAt:    s.refreshService.GetExpirationTime(),
		IsActive:     true,
	}

	// Save session
	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create session")
	}

	return createdSession, nil
}

// CreatePendingSession creates a session that requires 2FA verification
func (s *SessionService) CreatePendingSession(ctx context.Context, userID primitive.ObjectID, userAgent, ipAddress string) (*models.Session, error) {
	// Create a temporary session without a full JWT token
	now := time.Now()

	// Generate a temporary token (not a full JWT)
	tempToken, err := generateSecureToken(32)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate temporary token")
	}

	session := &models.Session{
		UserID:       userID,
		Token:        "pending_" + tempToken, // Mark as pending
		RefreshToken: "",                     // No refresh token for pending sessions
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		Device:       s.DetectDevice(userAgent),
		Location:     s.GetLocationFromIP(ipAddress),
		CreatedAt:    now,
		LastUsedAt:   now,
		ExpiresAt:    now.Add(15 * time.Minute), // Short expiry for pending sessions
		IsActive:     true,
	}

	// Save session
	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create pending session")
	}

	return createdSession, nil
}

// CompleteTwoFactorAuth completes a pending session after 2FA verification
func (s *SessionService) CompleteTwoFactorAuth(ctx context.Context, sessionID primitive.ObjectID) (*models.Session, error) {
	// Get the pending session
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find session")
	}

	// Check if session is pending
	if !strings.HasPrefix(session.Token, "pending_") {
		return nil, errors.New(errors.CodeInvalidOperation, "Session is not pending 2FA verification")
	}

	// Generate full JWT token
	token, err := s.jwtService.GenerateToken(session.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate token")
	}

	// Generate refresh token
	refreshToken, err := s.refreshService.GenerateRefreshToken()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate refresh token")
	}

	// Update session
	now := time.Now()
	session.Token = token
	session.RefreshToken = refreshToken
	session.LastUsedAt = now
	session.ExpiresAt = s.refreshService.GetExpirationTime()

	// Save updated session
	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update session")
	}

	return session, nil
}

// GetActiveSessionCount returns the number of active sessions for a user
func (s *SessionService) GetActiveSessionCount(ctx context.Context, userID primitive.ObjectID) (int, error) {
	sessions, err := s.sessionRepo.FindByUserID(ctx, userID)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to find sessions")
	}

	// Count active and non-expired sessions
	count := 0
	now := time.Now()
	for _, session := range sessions {
		if session.IsActive && now.Before(session.ExpiresAt) {
			count++
		}
	}

	return count, nil
}

// CleanupExpiredSessions removes expired sessions
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.DeleteExpired(ctx)
}

// DetectDevice extracts device information from user agent
func (s *SessionService) DetectDevice(userAgentString string) string {
	if userAgentString == "" {
		return "unknown"
	}

	ua := user_agent.New(userAgentString)

	// Get browser
	browserName, browserVersion := ua.Browser()

	// Get OS
	os := ua.OS()

	// Get device type
	deviceType := "desktop"
	if ua.Mobile() {
		deviceType = "mobile"
	} else if strings.Contains(strings.ToLower(userAgentString), "tablet") {
		deviceType = "tablet"
	}

	return fmt.Sprintf("%s / %s %s / %s", deviceType, browserName, browserVersion, os)
}

// GetLocationFromIP gets location information from IP address
func (s *SessionService) GetLocationFromIP(ipAddress string) string {
	if ipAddress == "" || ipAddress == "127.0.0.1" || ipAddress == "::1" {
		return "local"
	}

	// Use GeoIP service to get location
	location, err := s.geoIPService.GetLocation(ipAddress)
	if err != nil {
		return "unknown"
	}

	if location.City != "" && location.Country != "" {
		return fmt.Sprintf("%s, %s", location.City, location.Country)
	} else if location.Country != "" {
		return location.Country
	}

	return "unknown"
}

// GeoIPService provides location information from IP addresses
type GeoIPService interface {
	GetLocation(ipAddress string) (*GeoLocation, error)
}

// GeoLocation contains geographical information about an IP address
type GeoLocation struct {
	Country  string
	City     string
	Region   string
	Timezone string
}
