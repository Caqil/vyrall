package interfaces

import (
	"context"

	"github.com/yourusername/social-media-platform/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentRepository defines the interface for comment data access
type CommentRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, comment *models.Comment) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Comment, error)
	Update(ctx context.Context, comment *models.Comment) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByPostID(ctx context.Context, postID primitive.ObjectID, limit, offset int) ([]*models.Comment, int, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Comment, int, error)
	GetReplies(ctx context.Context, parentID primitive.ObjectID, limit, offset int) ([]*models.Comment, int, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Comment, error)

	// Comment tree operations
	GetCommentThread(ctx context.Context, rootID primitive.ObjectID, limit, offset int) ([]*models.Comment, int, error)
	GetTopLevelComments(ctx context.Context, postID primitive.ObjectID, limit, offset int) ([]*models.Comment, int, error)

	// Counters
	IncrementLikeCount(ctx context.Context, id primitive.ObjectID, amount int) error
	IncrementReplyCount(ctx context.Context, id primitive.ObjectID, amount int) error
	UpdateReactionCounts(ctx context.Context, id primitive.ObjectID, reactionCounts map[string]int) error

	// Comment management
	PinComment(ctx context.Context, id primitive.ObjectID, isPinned bool) error
	HideComment(ctx context.Context, id primitive.ObjectID, isHidden bool) error

	// Content moderation
	GetReportedComments(ctx context.Context, status string, limit, offset int) ([]*models.Comment, int, error)
	FlagAsInappropriate(ctx context.Context, id primitive.ObjectID, reason string) error

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Comment, int, error)

	// Analytics
	GetMostLikedComments(ctx context.Context, postID primitive.ObjectID, limit int) ([]*models.Comment, error)
	GetUserCommentActivity(ctx context.Context, userID primitive.ObjectID, startDate, endDate string) (int, error)
}
