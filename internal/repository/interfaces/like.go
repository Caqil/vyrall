package interfaces

import (
	"context"

	"vyrall/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LikeRepository defines the interface for like/reaction data access
type LikeRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, like *models.Like) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Like, error)
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByUserAndContent(ctx context.Context, userID, contentID primitive.ObjectID, contentType string) (*models.Like, error)
	GetByContentID(ctx context.Context, contentID primitive.ObjectID, contentType string, limit, offset int) ([]*models.Like, int, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Like, int, error)

	// Reaction management
	UpdateReactionType(ctx context.Context, id primitive.ObjectID, reactionType string) error
	GetContentLikeCount(ctx context.Context, contentID primitive.ObjectID, contentType string) (int, error)
	GetContentReactionCounts(ctx context.Context, contentID primitive.ObjectID, contentType string) (map[string]int, error)

	// User interactions
	GetUserLikedContent(ctx context.Context, userID primitive.ObjectID, contentType string, limit, offset int) ([]primitive.ObjectID, int, error)
	CheckIfUserLiked(ctx context.Context, userID, contentID primitive.ObjectID, contentType string) (bool, string, error)

	// Reaction analytics
	GetTopReactedContent(ctx context.Context, contentType string, limit int) ([]primitive.ObjectID, error)
	GetUserMostUsedReactions(ctx context.Context, userID primitive.ObjectID) (map[string]int, error)

	// Bulk operations
	BulkCreate(ctx context.Context, likes []*models.Like) ([]primitive.ObjectID, error)
	BulkDelete(ctx context.Context, ids []primitive.ObjectID) (int, error)
	DeleteByUserAndContent(ctx context.Context, userID, contentID primitive.ObjectID, contentType string) error

	// Advanced filtering and pagination
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Like, int, error)
}
