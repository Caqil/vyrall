package interfaces

import (
	"context"

	"github.com/Caqil/vyrall/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupRepository defines the interface for group data access
type GroupRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, group *models.Group) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Group, error)
	Update(ctx context.Context, group *models.Group) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SoftDelete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByName(ctx context.Context, name string) (*models.Group, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, role string, limit, offset int) ([]*models.Group, int, error)
	GetByCategoryIDs(ctx context.Context, categoryIDs []primitive.ObjectID, limit, offset int) ([]*models.Group, int, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Group, error)

	// Member management
	AddMember(ctx context.Context, groupID, userID primitive.ObjectID, role string) error
	RemoveMember(ctx context.Context, groupID, userID primitive.ObjectID) error
	UpdateMemberRole(ctx context.Context, groupID, userID primitive.ObjectID, role string) error
	GetMembers(ctx context.Context, groupID primitive.ObjectID, role string, limit, offset int) ([]*models.GroupMember, int, error)
	GetMember(ctx context.Context, groupID, userID primitive.ObjectID) (*models.GroupMember, error)

	// Join requests
	AddJoinRequest(ctx context.Context, groupID primitive.ObjectID, request models.GroupJoinRequest) error
	UpdateJoinRequest(ctx context.Context, groupID, userID primitive.ObjectID, status, reason string, reviewerID primitive.ObjectID) error
	GetPendingJoinRequests(ctx context.Context, groupID primitive.ObjectID, limit, offset int) ([]models.GroupJoinRequest, int, error)

	// Content management
	IncrementPostCount(ctx context.Context, groupID primitive.ObjectID) error
	DecrementPostCount(ctx context.Context, groupID primitive.ObjectID) error
	IncrementMemberCount(ctx context.Context, groupID primitive.ObjectID) error
	DecrementMemberCount(ctx context.Context, groupID primitive.ObjectID) error
	IncrementEventCount(ctx context.Context, groupID primitive.ObjectID) error
	DecrementEventCount(ctx context.Context, groupID primitive.ObjectID) error

	// Group settings
	UpdateRules(ctx context.Context, groupID primitive.ObjectID, rules []models.GroupRule) error
	UpdateFeatures(ctx context.Context, groupID primitive.ObjectID, features models.GroupFeatures) error
	UpdateVisibility(ctx context.Context, groupID primitive.ObjectID, isPublic, isVisible bool) error
	UpdateJoinSettings(ctx context.Context, groupID primitive.ObjectID, requireApproval bool) error

	// Admin operations
	AddAdmin(ctx context.Context, groupID, userID primitive.ObjectID) error
	RemoveAdmin(ctx context.Context, groupID, userID primitive.ObjectID) error
	AddModerator(ctx context.Context, groupID, userID primitive.ObjectID) error
	RemoveModerator(ctx context.Context, groupID, userID primitive.ObjectID) error

	// Invite link
	GenerateInviteLink(ctx context.Context, groupID primitive.ObjectID) (string, error)
	ValidateInviteLink(ctx context.Context, link string) (primitive.ObjectID, error)

	// Group discovery
	GetRecommendedGroups(ctx context.Context, userID primitive.ObjectID, limit int) ([]*models.Group, error)
	GetPopularGroups(ctx context.Context, limit int) ([]*models.Group, error)
	GetNewGroups(ctx context.Context, limit int) ([]*models.Group, error)

	// Activity tracking
	UpdateLastActivity(ctx context.Context, groupID primitive.ObjectID) error
	UpdateMemberLastActivity(ctx context.Context, groupID, userID primitive.ObjectID) error

	// Analytics
	GetGroupStats(ctx context.Context, groupID primitive.ObjectID) (models.GroupAnalytics, error)

	// Search and filtering
	Search(ctx context.Context, query string, filter map[string]interface{}, limit, offset int) ([]*models.Group, int, error)
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Group, int, error)
}
