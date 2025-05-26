package interfaces

import (
	"context"
	"time"

	"vyrall/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportRepository defines the interface for content reporting data access
type ReportRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, report *models.Report) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Report, error)
	Update(ctx context.Context, report *models.Report) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByReporterID(ctx context.Context, reporterID primitive.ObjectID, limit, offset int) ([]*models.Report, int, error)
	GetByContentID(ctx context.Context, contentID primitive.ObjectID, contentType string) ([]*models.Report, error)
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Report, int, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Report, error)

	// Report management
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error
	AssignModerator(ctx context.Context, id primitive.ObjectID, moderatorID primitive.ObjectID) error
	AddModeratorNotes(ctx context.Context, id primitive.ObjectID, notes string) error
	TakeAction(ctx context.Context, id primitive.ObjectID, action string, moderatorID primitive.ObjectID, notes string) error

	// Content moderation
	GetMostReportedContent(ctx context.Context, contentType string, limit int) ([]primitive.ObjectID, error)
	GetReportCountByContent(ctx context.Context, contentID primitive.ObjectID, contentType string) (int, error)
	GetReportsByReasonCode(ctx context.Context, reasonCode string, limit, offset int) ([]*models.Report, int, error)

	// Reporting statistics
	GetReportStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error)
	GetReportCountByUser(ctx context.Context, userID primitive.ObjectID) (int, error)
	GetReportedUserStats(ctx context.Context, userID primitive.ObjectID) (map[string]interface{}, error)

	// Bulk operations
	BulkUpdateStatus(ctx context.Context, ids []primitive.ObjectID, status string) (int, error)
	BulkAssignModerator(ctx context.Context, ids []primitive.ObjectID, moderatorID primitive.ObjectID) (int, error)

	// Time-based operations
	GetReportsByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*models.Report, int, error)

	// Appeal management
	SubmitAppeal(ctx context.Context, reportID primitive.ObjectID, reason string) error
	ReviewAppeal(ctx context.Context, reportID primitive.ObjectID, approved bool, reviewerID primitive.ObjectID, notes string) error
	GetAppealedReports(ctx context.Context, limit, offset int) ([]*models.Report, int, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Report, int, error)
}
