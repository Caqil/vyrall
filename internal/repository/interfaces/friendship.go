package interfaces

import (
	"context"

	"vyrall/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FriendshipRepository defines the interface for friendship data access
type FriendshipRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, friendship *models.Friendship) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Friendship, error)
	Update(ctx context.Context, friendship *models.Friendship) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetBetweenUsers(ctx context.Context, user1ID, user2ID primitive.ObjectID) (*models.Friendship, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, status string, limit, offset int) ([]*models.Friendship, int, error)
	GetByStatus(ctx context.Context, userID primitive.ObjectID, status string, limit, offset int) ([]*models.Friendship, int, error)

	// Status management
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error
	AcceptFriendRequest(ctx context.Context, id primitive.ObjectID) error
	RejectFriendRequest(ctx context.Context, id primitive.ObjectID) error
	BlockFriend(ctx context.Context, user1ID, user2ID primitive.ObjectID) error
	UnblockFriend(ctx context.Context, user1ID, user2ID primitive.ObjectID) error

	// Friendship requests
	GetPendingRequests(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Friendship, int, error)
	GetSentRequests(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Friendship, int, error)
	GetFriendshipRequestCount(ctx context.Context, userID primitive.ObjectID) (int, error)

	// Friends management
	GetFriends(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]primitive.ObjectID, int, error)
	GetFriendsWithDetails(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Friendship, int, error)
	GetBlockedUsers(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]primitive.ObjectID, int, error)
	IsFriend(ctx context.Context, user1ID, user2ID primitive.ObjectID) (bool, error)
	IsBlocked(ctx context.Context, user1ID, user2ID primitive.ObjectID) (bool, error)

	// Friend suggestions
	GetFriendSuggestions(ctx context.Context, userID primitive.ObjectID, limit int) ([]primitive.ObjectID, error)
	GetMutualFriends(ctx context.Context, user1ID, user2ID primitive.ObjectID, limit, offset int) ([]primitive.ObjectID, int, error)
	GetMutualFriendsCount(ctx context.Context, user1ID, user2ID primitive.ObjectID) (int, error)

	// Bulk operations
	BulkCreate(ctx context.Context, friendships []*models.Friendship) ([]primitive.ObjectID, error)
	BulkUpdate(ctx context.Context, ids []primitive.ObjectID, status string) (int, error)
	BulkDelete(ctx context.Context, ids []primitive.ObjectID) (int, error)

	// Count operations
	CountFriends(ctx context.Context, userID primitive.ObjectID) (int, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Friendship, int, error)
}
