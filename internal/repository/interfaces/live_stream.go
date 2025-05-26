package interfaces

import (
	"context"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LiveStreamRepository defines the interface for live stream data access
type LiveStreamRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, liveStream *models.LiveStream) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.LiveStream, error)
	Update(ctx context.Context, liveStream *models.LiveStream) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByUserID(ctx context.Context, userID primitive.ObjectID, status string, limit, offset int) ([]*models.LiveStream, int, error)
	GetActive(ctx context.Context, limit, offset int) ([]*models.LiveStream, int, error)
	GetScheduled(ctx context.Context, limit, offset int) ([]*models.LiveStream, int, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.LiveStream, error)

	// Stream lifecycle management
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error
	StartStream(ctx context.Context, id primitive.ObjectID) error
	EndStream(ctx context.Context, id primitive.ObjectID) error

	// Stream metrics
	IncrementViewerCount(ctx context.Context, id primitive.ObjectID) error
	DecrementViewerCount(ctx context.Context, id primitive.ObjectID) error
	UpdatePeakViewerCount(ctx context.Context, id primitive.ObjectID) error
	IncrementTotalViews(ctx context.Context, id primitive.ObjectID) error
	IncrementLikeCount(ctx context.Context, id primitive.ObjectID) error
	IncrementCommentCount(ctx context.Context, id primitive.ObjectID) error
	IncrementShareCount(ctx context.Context, id primitive.ObjectID) error

	// Viewer management
	AddViewer(ctx context.Context, streamID, userID primitive.ObjectID, device, platform string) (primitive.ObjectID, error)
	RemoveViewer(ctx context.Context, streamID, userID primitive.ObjectID) error
	GetViewers(ctx context.Context, streamID primitive.ObjectID, limit, offset int) ([]*models.LiveStreamViewer, int, error)
	GetActiveViewers(ctx context.Context, streamID primitive.ObjectID) ([]*models.LiveStreamViewer, error)

	// Comment operations
	AddComment(ctx context.Context, comment *models.LiveStreamComment) (primitive.ObjectID, error)
	GetComments(ctx context.Context, streamID primitive.ObjectID, limit, offset int) ([]*models.LiveStreamComment, int, error)
	DeleteComment(ctx context.Context, commentID primitive.ObjectID) error
	PinComment(ctx context.Context, commentID primitive.ObjectID, isPinned bool) error
	ModerateComment(ctx context.Context, commentID primitive.ObjectID, isHidden bool, moderatorID primitive.ObjectID, reason string) error

	// Reaction operations
	AddReaction(ctx context.Context, reaction *models.LiveStreamReaction) (primitive.ObjectID, error)
	GetReactions(ctx context.Context, streamID primitive.ObjectID, limit, offset int) ([]*models.LiveStreamReaction, int, error)

	// Chat settings
	UpdateChatSettings(ctx context.Context, streamID primitive.ObjectID, settings models.LiveStreamChatSettings) error
	AddModerator(ctx context.Context, streamID, userID primitive.ObjectID) error
	RemoveModerator(ctx context.Context, streamID, userID primitive.ObjectID) error
	BanUser(ctx context.Context, streamID, userID primitive.ObjectID) error
	UnbanUser(ctx context.Context, streamID, userID primitive.ObjectID) error

	// Recording management
	EnableRecording(ctx context.Context, streamID primitive.ObjectID) error
	DisableRecording(ctx context.Context, streamID primitive.ObjectID) error
	UpdateRecordingURL(ctx context.Context, streamID primitive.ObjectID, url string) error

	// Stream discovery
	GetTopStreams(ctx context.Context, limit int) ([]*models.LiveStream, error)
	GetRecommendedStreams(ctx context.Context, userID primitive.ObjectID, limit int) ([]*models.LiveStream, error)
	GetStreamsByCategories(ctx context.Context, categories []string, limit, offset int) ([]*models.LiveStream, int, error)
	GetStreamsByTags(ctx context.Context, tags []string, limit, offset int) ([]*models.LiveStream, int, error)

	// Analytics
	GetStreamAnalytics(ctx context.Context, streamID primitive.ObjectID) (map[string]interface{}, error)

	// Time-based operations
	GetStreamsByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*models.LiveStream, int, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.LiveStream, int, error)

	// Search
	Search(ctx context.Context, query string, limit, offset int) ([]*models.LiveStream, int, error)
}
