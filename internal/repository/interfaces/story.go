package interfaces

import (
	"context"
	"time"

	"github.com/yourusername/social-media-platform/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StoryRepository defines the interface for story data access
type StoryRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, story *models.Story) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Story, error)
	Update(ctx context.Context, story *models.Story) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Story, error)
	GetActiveByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Story, error)
	GetExpired(ctx context.Context, before time.Time) ([]*models.Story, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Story, error)

	// Feed operations
	GetStoriesForUser(ctx context.Context, viewerID primitive.ObjectID) (map[primitive.ObjectID][]*models.Story, error)
	GetHighlightedStoriesForUser(ctx context.Context, userID primitive.ObjectID) ([]*models.Story, error)

	// Viewer operations
	AddViewer(ctx context.Context, storyID, userID primitive.ObjectID, device string) error
	GetViewers(ctx context.Context, storyID primitive.ObjectID) ([]models.StoryViewer, error)
	IncrementViewCount(ctx context.Context, storyID primitive.ObjectID) error

	// Interaction operations
	AddReaction(ctx context.Context, storyID primitive.ObjectID, reaction models.StoryReaction) error
	RemoveReaction(ctx context.Context, storyID, userID primitive.ObjectID) error
	AddReply(ctx context.Context, storyID primitive.ObjectID, reply models.StoryReply) (primitive.ObjectID, error)
	GetReplies(ctx context.Context, storyID primitive.ObjectID) ([]models.StoryReply, error)
	IncrementReactionCount(ctx context.Context, storyID primitive.ObjectID) error
	DecrementReactionCount(ctx context.Context, storyID primitive.ObjectID) error
	IncrementReplyCount(ctx context.Context, storyID primitive.ObjectID) error

	// Story highlights
	AddToHighlight(ctx context.Context, storyID, highlightID primitive.ObjectID) error
	RemoveFromHighlight(ctx context.Context, storyID primitive.ObjectID) error
	GetStoriesByHighlightID(ctx context.Context, highlightID primitive.ObjectID) ([]*models.Story, error)

	// Interactive elements
	AddInteractiveResponse(ctx context.Context, storyID primitive.ObjectID, response models.InteractiveResponse) error
	GetInteractiveResponses(ctx context.Context, storyID primitive.ObjectID) ([]models.InteractiveResponse, error)
	UpdateInteractiveData(ctx context.Context, storyID primitive.ObjectID, data map[string]interface{}) error

	// Link tracking
	IncrementLinkClicks(ctx context.Context, storyID primitive.ObjectID) error

	// Story archiving
	ArchiveStory(ctx context.Context, storyID primitive.ObjectID) error
	GetArchivedStories(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Story, int, error)

	// Privacy management
	UpdatePrivacySettings(ctx context.Context, storyID primitive.ObjectID, privacy models.StoryPrivacy) error

	// Analytics
	GetStoryPerformance(ctx context.Context, storyID primitive.ObjectID) (map[string]interface{}, error)
	GetUserStoriesPerformance(ctx context.Context, userID primitive.ObjectID, startDate, endDate time.Time) (map[string]interface{}, error)

	// Bulk operations
	DeleteExpiredStories(ctx context.Context) (int, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Story, int, error)
}
