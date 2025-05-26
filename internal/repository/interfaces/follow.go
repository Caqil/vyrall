package interfaces

import (
	"context"

	"vyrall/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FollowRepository defines the interface for follow relationship data access
type FollowRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, follow *models.Follow) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Follow, error)
	Update(ctx context.Context, follow *models.Follow) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByFollowerAndFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (*models.Follow, error)
	GetFollowers(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Follow, int, error)
	GetFollowing(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Follow, int, error)

	// Status management
	UpdateStatus(ctx context.Context, followerID, followingID primitive.ObjectID, status string) error
	GetPendingFollowRequests(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Follow, int, error)

	// Specialized queries
	GetMutualFollows(ctx context.Context, user1ID, user2ID primitive.ObjectID, limit, offset int) ([]*models.Follow, int, error)
	GetFollowerIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error)
	GetFollowingIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error)

	// Follow suggestions
	GetFollowSuggestions(ctx context.Context, userID primitive.ObjectID, limit int) ([]primitive.ObjectID, error)

	// Notification settings
	UpdateNotifyPosts(ctx context.Context, followerID, followingID primitive.ObjectID, notify bool) error

	// Bulk operations
	BulkCreate(ctx context.Context, follows []*models.Follow) ([]primitive.ObjectID, error)
	BulkDelete(ctx context.Context, ids []primitive.ObjectID) (int, error)

	// Count operations
	CountFollowers(ctx context.Context, userID primitive.ObjectID) (int, error)
	CountFollowing(ctx context.Context, userID primitive.ObjectID) (int, error)

	// Check operations
	IsFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (bool, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Follow, int, error)
}
