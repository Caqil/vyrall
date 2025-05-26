package interfaces

import (
	"context"

	"github.com/Caqil/vyrall/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ConversationRepository defines the interface for conversation data access
type ConversationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, conversation *models.Conversation) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Conversation, error)
	Update(ctx context.Context, conversation *models.Conversation) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SoftDelete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error

	// Query operations
	GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Conversation, int, error)
	GetByParticipantIDs(ctx context.Context, participantIDs []primitive.ObjectID) (*models.Conversation, error)
	GetDirectConversation(ctx context.Context, user1ID, user2ID primitive.ObjectID) (*models.Conversation, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Conversation, error)

	// Participant management
	AddParticipant(ctx context.Context, conversationID, userID primitive.ObjectID, role string) error
	RemoveParticipant(ctx context.Context, conversationID, userID primitive.ObjectID) error
	UpdateParticipantRole(ctx context.Context, conversationID, userID primitive.ObjectID, role string) error
	GetParticipants(ctx context.Context, conversationID primitive.ObjectID) ([]models.Participant, error)

	// Group chat operations
	UpdateGroupInfo(ctx context.Context, conversationID primitive.ObjectID, info models.GroupChatInfo) error
	AddJoinRequest(ctx context.Context, conversationID primitive.ObjectID, request models.JoinRequest) error
	UpdateJoinRequest(ctx context.Context, conversationID primitive.ObjectID, userID primitive.ObjectID, status string, reviewerID primitive.ObjectID) error
	GetPendingJoinRequests(ctx context.Context, conversationID primitive.ObjectID) ([]models.JoinRequest, error)

	// Status management
	UpdateLastMessage(ctx context.Context, conversationID, messageID, senderID primitive.ObjectID, preview string) error
	IncrementMessageCount(ctx context.Context, conversationID primitive.ObjectID) error
	UpdateParticipantLastRead(ctx context.Context, conversationID, userID, messageID primitive.ObjectID) error
	UpdateUnreadCount(ctx context.Context, conversationID, userID primitive.ObjectID, count int) error
	MarkTyping(ctx context.Context, conversationID, userID primitive.ObjectID) error

	// User preferences
	MuteConversation(ctx context.Context, conversationID, userID primitive.ObjectID, until primitive.DateTime) error
	UnmuteConversation(ctx context.Context, conversationID, userID primitive.ObjectID) error
	ArchiveConversation(ctx context.Context, conversationID, userID primitive.ObjectID) error
	UnarchiveConversation(ctx context.Context, conversationID, userID primitive.ObjectID) error

	// Encryption
	EnableEncryption(ctx context.Context, conversationID primitive.ObjectID) error
	DisableEncryption(ctx context.Context, conversationID primitive.ObjectID) error

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Conversation, int, error)

	// Search
	SearchConversations(ctx context.Context, userID primitive.ObjectID, query string, limit, offset int) ([]*models.Conversation, int, error)
}
