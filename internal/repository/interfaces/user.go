package interfaces

import (
	"context"

	"vyrall/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *models.User) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Specialized queries
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.User, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.User, int, error)
	GetSuggestions(ctx context.Context, userID primitive.ObjectID, limit int) ([]*models.User, error)
	GetPopular(ctx context.Context, limit int) ([]*models.User, error)

	// User status and verification
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error
	UpdateVerificationStatus(ctx context.Context, id primitive.ObjectID, isVerified bool) error

	// User settings
	UpdateSettings(ctx context.Context, id primitive.ObjectID, settings models.UserSettings) error
	UpdateNotificationPreferences(ctx context.Context, id primitive.ObjectID, prefs models.NotificationPreferences) error
	UpdatePrivacySettings(ctx context.Context, id primitive.ObjectID, settings models.PrivacySettings) error

	// User profile
	UpdateProfile(ctx context.Context, id primitive.ObjectID, profileData map[string]interface{}) error
	IncrementFollowerCount(ctx context.Context, id primitive.ObjectID, amount int) error
	IncrementFollowingCount(ctx context.Context, id primitive.ObjectID, amount int) error
	IncrementPostCount(ctx context.Context, id primitive.ObjectID, amount int) error

	// Authentication related
	ChangePassword(ctx context.Context, id primitive.ObjectID, hashedPassword string) error
	UpdateLastActive(ctx context.Context, id primitive.ObjectID) error
	SetEmailVerified(ctx context.Context, id primitive.ObjectID, verified bool) error
	SetTwoFactorEnabled(ctx context.Context, id primitive.ObjectID, enabled bool) error

	// Pagination and filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.User, int, error)
}
