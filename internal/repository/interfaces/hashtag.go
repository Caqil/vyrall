package interfaces

import (
	"context"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HashtagRepository defines the interface for hashtag data access
type HashtagRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, hashtag *models.Hashtag) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Hashtag, error)
	Update(ctx context.Context, hashtag *models.Hashtag) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByName(ctx context.Context, name string) (*models.Hashtag, error)
	GetByNames(ctx context.Context, names []string) ([]*models.Hashtag, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Hashtag, error)

	// Trending hashtags
	GetTrending(ctx context.Context, limit int) ([]*models.Hashtag, error)
	GetTrendingByCategory(ctx context.Context, category string, limit int) ([]*models.Hashtag, error)
	UpdateTrendingStatus(ctx context.Context, id primitive.ObjectID, isTrending bool, rank int) error

	// Post association
	IncrementPostCount(ctx context.Context, id primitive.ObjectID) error
	DecrementPostCount(ctx context.Context, id primitive.ObjectID) error
	GetPostsByHashtag(ctx context.Context, hashtagName string, limit, offset int) ([]primitive.ObjectID, int, error)

	// User interaction
	IncrementFollowerCount(ctx context.Context, id primitive.ObjectID) error
	DecrementFollowerCount(ctx context.Context, id primitive.ObjectID) error

	// HashtagFollow operations
	CreateFollow(ctx context.Context, follow *models.HashtagFollow) (primitive.ObjectID, error)
	DeleteFollow(ctx context.Context, userID, hashtagID primitive.ObjectID) error
	GetFollowedHashtags(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Hashtag, int, error)
	IsFollowing(ctx context.Context, userID, hashtagID primitive.ObjectID) (bool, error)
	GetHashtagFollowers(ctx context.Context, hashtagID primitive.ObjectID, limit, offset int) ([]primitive.ObjectID, int, error)

	// Content moderation
	SetRestricted(ctx context.Context, id primitive.ObjectID, isRestricted bool) error
	GetRestrictedHashtags(ctx context.Context) ([]*models.Hashtag, error)

	// Analytics
	GetPopularHashtags(ctx context.Context, timeRange time.Duration, limit int) ([]*models.Hashtag, error)
	GetHashtagGrowth(ctx context.Context, id primitive.ObjectID, period string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetHashtagEngagement(ctx context.Context, id primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)

	// Search and discovery
	Search(ctx context.Context, query string, limit int) ([]*models.Hashtag, error)
	GetRelatedHashtags(ctx context.Context, hashtagName string, limit int) ([]*models.Hashtag, error)
	GetRecommendedHashtags(ctx context.Context, userID primitive.ObjectID, limit int) ([]*models.Hashtag, error)

	// Bulk operations
	BulkCreate(ctx context.Context, hashtags []*models.Hashtag) ([]primitive.ObjectID, error)
	BulkUpdate(ctx context.Context, hashtags []*models.Hashtag) (int, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Hashtag, int, error)
}
