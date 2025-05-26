package interfaces

import (
	"context"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationRepository defines the interface for notification data access
type NotificationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, notification *models.Notification) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Notification, error)
	Update(ctx context.Context, notification *models.Notification) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Notification, int, error)
	GetUnreadByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Notification, int, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Notification, error)

	// Status management
	MarkAsRead(ctx context.Context, id primitive.ObjectID) error
	MarkAllAsRead(ctx context.Context, userID primitive.ObjectID) (int, error)
	MarkAsSent(ctx context.Context, id primitive.ObjectID) error
	UpdateDeliveryStatus(ctx context.Context, id primitive.ObjectID, channel string, status string) error

	// Grouping operations
	GetGroupedNotifications(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Notification, int, error)
	GetByGroupKey(ctx context.Context, userID primitive.ObjectID, groupKey string) ([]*models.Notification, error)

	// Time-based operations
	GetNotificationsByTimeRange(ctx context.Context, userID primitive.ObjectID, startTime, endTime time.Time, limit, offset int) ([]*models.Notification, int, error)
	DeleteExpiredNotifications(ctx context.Context) (int, error)

	// Counts
	GetUnreadCount(ctx context.Context, userID primitive.ObjectID) (int, error)

	// Notification management
	HideNotification(ctx context.Context, id primitive.ObjectID) error
	UpdateExpiryTime(ctx context.Context, id primitive.ObjectID, expiresAt time.Time) error

	// User preferences
	GetUserNotificationPreferences(ctx context.Context, userID primitive.ObjectID) (map[string]models.NotificationPreferences, error)

	// Batch operations
	CreateBatch(ctx context.Context, notifications []*models.Notification) ([]primitive.ObjectID, error)
	MarkMultipleAsRead(ctx context.Context, ids []primitive.ObjectID) (int, error)
	DeleteMultiple(ctx context.Context, ids []primitive.ObjectID) (int, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Notification, int, error)
}
