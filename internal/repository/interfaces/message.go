package interfaces

import (
	"context"
	"time"

	"vyrall/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageRepository defines the interface for message data access
type MessageRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, message *models.Message) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Message, error)
	Update(ctx context.Context, message *models.Message) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SoftDelete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error

	// Query operations
	GetByConversationID(ctx context.Context, conversationID primitive.ObjectID, limit, offset int) ([]*models.Message, int, error)
	GetUnreadMessagesCount(ctx context.Context, userID, conversationID primitive.ObjectID) (int, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Message, error)

	// Message management
	MarkAsDelivered(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
	MarkAsRead(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
	AddReaction(ctx context.Context, id primitive.ObjectID, reaction models.MessageReaction) error
	RemoveReaction(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, reaction string) error

	// Message editing
	EditMessage(ctx context.Context, id primitive.ObjectID, content string) error
	AddToEditHistory(ctx context.Context, id primitive.ObjectID, edit models.MessageEdit) error

	// Message forwarding and replies
	ForwardMessage(ctx context.Context, messageID, toConversationID, senderID primitive.ObjectID) (primitive.ObjectID, error)
	GetReplies(ctx context.Context, messageID primitive.ObjectID) ([]*models.Message, error)

	// Encryption
	UpdateEncryptionDetails(ctx context.Context, id primitive.ObjectID, details models.EncryptionDetails) error

	// System messages
	CreateSystemMessage(ctx context.Context, conversationID primitive.ObjectID, msgType string, params map[string]interface{}) (primitive.ObjectID, error)

	// Time-based operations
	GetMessagesByTimeRange(ctx context.Context, conversationID primitive.ObjectID, startTime, endTime time.Time, limit, offset int) ([]*models.Message, int, error)
	DeleteExpiredMessages(ctx context.Context) (int, error)

	// Voice and media messages
	GetMediaMessages(ctx context.Context, conversationID primitive.ObjectID, mediaType string, limit, offset int) ([]*models.Message, int, error)

	// Bulk operations
	BulkDelete(ctx context.Context, ids []primitive.ObjectID, userID primitive.ObjectID) (int, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Message, int, error)

	// Search
	SearchMessages(ctx context.Context, conversationID primitive.ObjectID, query string, limit, offset int) ([]*models.Message, int, error)
}
