package interfaces

import (
	"context"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostRepository defines the interface for post data access
type PostRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, post *models.Post) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Post, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error

	// Retrieval operations
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Post, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Post, int, error)
	GetByGroupID(ctx context.Context, groupID primitive.ObjectID, limit, offset int) ([]*models.Post, int, error)
	GetByHashtag(ctx context.Context, hashtag string, limit, offset int) ([]*models.Post, int, error)
	GetByLocation(ctx context.Context, latitude, longitude float64, radiusKm float64, limit, offset int) ([]*models.Post, int, error)

	// Feed operations
	GetFeedForUser(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Post, int, error)
	GetTimelineForUser(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Post, int, error)
	GetExploreForUser(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Post, int, error)

	// Content moderation
	GetReportedPosts(ctx context.Context, status string, limit, offset int) ([]*models.Post, int, error)
	UpdatePostVisibility(ctx context.Context, id primitive.ObjectID, isHidden bool) error
	FlagAsInappropriate(ctx context.Context, id primitive.ObjectID, reason string) error

	// Scheduling
	GetScheduledPosts(ctx context.Context, userID primitive.ObjectID, startTime, endTime time.Time) ([]*models.Post, error)
	PublishScheduledPost(ctx context.Context, id primitive.ObjectID) error

	// Counters and metrics
	IncrementLikeCount(ctx context.Context, id primitive.ObjectID, amount int) error
	IncrementCommentCount(ctx context.Context, id primitive.ObjectID, amount int) error
	IncrementShareCount(ctx context.Context, id primitive.ObjectID, amount int) error
	IncrementViewCount(ctx context.Context, id primitive.ObjectID, amount int) error
	UpdateReactionCounts(ctx context.Context, id primitive.ObjectID, reactionCounts map[string]int) error

	// Post management
	PinPost(ctx context.Context, id primitive.ObjectID, isPinned bool) error
	ArchivePost(ctx context.Context, id primitive.ObjectID, isArchived bool) error
	FeaturePost(ctx context.Context, id primitive.ObjectID, isFeatured bool) error

	// Search
	Search(ctx context.Context, query string, filter map[string]interface{}, limit, offset int) ([]*models.Post, int, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Post, int, error)

	// Polls
	GetPostsWithActivePoll(ctx context.Context, limit, offset int) ([]*models.Post, int, error)
	UpdatePoll(ctx context.Context, postID primitive.ObjectID, poll *models.Poll) error
}
