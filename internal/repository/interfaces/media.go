package interfaces

import (
	"context"

	"vyrall/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaRepository defines the interface for media data access
type MediaRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, media *models.Media) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Media, error)
	Update(ctx context.Context, media *models.Media) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetByUserID(ctx context.Context, userID primitive.ObjectID, mediaType string, limit, offset int) ([]*models.Media, int, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Media, error)

	// Media processing
	UpdateProcessingStatus(ctx context.Context, id primitive.ObjectID, status string, isProcessed bool) error
	MarkAsProcessed(ctx context.Context, id primitive.ObjectID) error
	GetUnprocessedMedia(ctx context.Context, limit int) ([]*models.Media, error)

	// Media metadata
	UpdateMetadata(ctx context.Context, id primitive.ObjectID, metadata map[string]interface{}) error
	UpdateDimensions(ctx context.Context, id primitive.ObjectID, width, height int) error
	UpdateDuration(ctx context.Context, id primitive.ObjectID, duration float64) error

	// Media accessibility
	UpdateAltText(ctx context.Context, id primitive.ObjectID, altText string) error
	UpdateCaption(ctx context.Context, id primitive.ObjectID, caption string) error

	// Media URLs
	UpdateURL(ctx context.Context, id primitive.ObjectID, url string) error
	UpdateThumbnailURL(ctx context.Context, id primitive.ObjectID, thumbnailURL string) error

	// Media usage
	GetMediaUsage(ctx context.Context, userID primitive.ObjectID) (int64, error)
	GetMediaByType(ctx context.Context, mediaType string, limit, offset int) ([]*models.Media, int, error)

	// Bulk operations
	BulkCreate(ctx context.Context, medias []*models.Media) ([]primitive.ObjectID, error)
	BulkDelete(ctx context.Context, ids []primitive.ObjectID) (int, error)

	// Media gallery
	GetUserGallery(ctx context.Context, userID primitive.ObjectID, mediaTypes []string, limit, offset int) ([]*models.Media, int, error)

	// Content association
	GetMediaByContentID(ctx context.Context, contentType string, contentID primitive.ObjectID) ([]*models.Media, error)
	AssociateWithContent(ctx context.Context, mediaID primitive.ObjectID, contentType string, contentID primitive.ObjectID) error
	DisassociateFromContent(ctx context.Context, mediaID primitive.ObjectID, contentType string, contentID primitive.ObjectID) error

	// Media search
	Search(ctx context.Context, query string, mediaTypes []string, limit, offset int) ([]*models.Media, int, error)

	// Advanced filtering
	List(ctx context.Context, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Media, int, error)
}
