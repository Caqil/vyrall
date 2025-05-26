package auth

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthService defines the interface for authentication operations (just the relevant methods for registration)
type AuthService interface {
	Register(user *models.User, password string) (primitive.ObjectID, error)
}

// RegisterRequest represents a registration request body
type RegisterRequest struct {
	Email       string    `json:"email" binding:"required,email"`
	Password    string    `json:"password" binding:"required,min=8"`
	Username    string    `json:"username" binding:"required,min=3,max=30"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
}

// Register creates a new user account
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	authService := c.MustGet("authService").(AuthService)

	// Create user model from request
	user := &models.User{
		Email:         req.Email,
		Username:      req.Username,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		DisplayName:   req.FirstName + " " + req.LastName,
		DateOfBirth:   req.DateOfBirth,
		Gender:        req.Gender,
		Status:        "pending_verification",
		Role:          "user",
		EmailVerified: false,
		IsPrivate:     true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Settings: models.UserSettings{
			LanguagePreference: "en",
			ThemePreference:    "light",
			AutoPlayVideos:     true,
			ShowOnlineStatus:   true,
			NotificationPreferences: models.NotificationPreferences{
				Push: models.NotificationTypes{
					Likes:          true,
					Comments:       true,
					Follows:        true,
					Messages:       true,
					TaggedPosts:    true,
					Stories:        true,
					LiveStreams:    true,
					GroupActivity:  true,
					EventReminders: true,
					SystemUpdates:  true,
				},
				Email: models.NotificationTypes{
					Likes:          false,
					Comments:       false,
					Follows:        true,
					Messages:       true,
					TaggedPosts:    true,
					Stories:        false,
					LiveStreams:    false,
					GroupActivity:  true,
					EventReminders: true,
					SystemUpdates:  true,
				},
				InApp: models.NotificationTypes{
					Likes:          true,
					Comments:       true,
					Follows:        true,
					Messages:       true,
					TaggedPosts:    true,
					Stories:        true,
					LiveStreams:    true,
					GroupActivity:  true,
					EventReminders: true,
					SystemUpdates:  true,
				},
				MessagePreviews: true,
			},
			PrivacySettings: models.PrivacySettings{
				WhoCanSeeMyPosts:        "followers",
				WhoCanSendMeMessages:    "followers",
				WhoCanSeeMyFriends:      "followers",
				WhoCanTagMe:             "followers",
				WhoCanSeeMyStories:      "followers",
				HideMyOnlineStatus:      false,
				HideMyLastSeen:          false,
				HideMyProfileFromSearch: false,
			},
		},
	}

	userID, err := authService.Register(user, req.Password)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Registration failed", err)
		return
	}

	response.Success(c, http.StatusCreated, "Registration successful. Please verify your email.", gin.H{
		"user_id": userID.Hex(),
	})
}
