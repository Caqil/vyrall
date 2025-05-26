package auth

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthService defines the interface for authentication operations (just the relevant methods for logout)
type AuthService interface {
	Logout(token string) error
	TerminateSession(sessionID primitive.ObjectID) error
	TerminateAllSessions(userID primitive.ObjectID) error
	GetActiveSessions(userID primitive.ObjectID) ([]models.Session, error)
}

// Logout terminates a user session
func Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		response.Error(c, http.StatusBadRequest, "No authorization token provided", nil)
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	authService := c.MustGet("authService").(AuthService)

	if err := authService.Logout(token); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to logout", err)
		return
	}

	response.Success(c, http.StatusOK, "Logout successful", nil)
}

// GetActiveSessions returns all active sessions for a user
func GetActiveSessions(c *gin.Context) {
	// Get user ID from the authenticated context
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	sessions, err := authService.GetActiveSessions(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve active sessions", err)
		return
	}

	response.Success(c, http.StatusOK, "Active sessions retrieved successfully", sessions)
}

// TerminateSession ends a specific session
func TerminateSession(c *gin.Context) {
	sessionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid session ID", err)
		return
	}

	// Get user ID from the authenticated context
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	// Get session info to verify ownership
	sessions, err := authService.GetActiveSessions(userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify session ownership", err)
		return
	}

	// Verify session belongs to user
	var sessionBelongsToUser bool
	for _, session := range sessions {
		if session.ID == sessionID {
			sessionBelongsToUser = true
			break
		}
	}

	if !sessionBelongsToUser {
		response.Error(c, http.StatusForbidden, "Session does not belong to authenticated user", nil)
		return
	}

	if err := authService.TerminateSession(sessionID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to terminate session", err)
		return
	}

	response.Success(c, http.StatusOK, "Session terminated successfully", nil)
}

// TerminateAllSessions ends all active sessions for a user except the current one
func TerminateAllSessions(c *gin.Context) {
	// Get user ID from the authenticated context
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	if err := authService.TerminateAllSessions(userID.(primitive.ObjectID)); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to terminate all sessions", err)
		return
	}

	response.Success(c, http.StatusOK, "All sessions terminated successfully", nil)
}
