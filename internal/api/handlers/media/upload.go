package media

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/services/media"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UploadHandler handles media file uploads
type UploadHandler struct {
	mediaService *media.Service
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(mediaService *media.Service) *UploadHandler {
	return &UploadHandler{
		mediaService: mediaService,
	}
}

// Upload handles the file upload request
func (h *UploadHandler) Upload(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse form with file
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		response.ValidationError(c, "Failed to parse form", err.Error())
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.ValidationError(c, "No file uploaded", err.Error())
		return
	}
	defer file.Close()

	// Validate file type
	fileType := validation.GetFileType(header.Filename)
	if fileType == "unknown" {
		response.ValidationError(c, "Unsupported file type", nil)
		return
	}

	// Get additional metadata
	caption := c.PostForm("caption")
	altText := c.PostForm("altText")

	// Create media object
	media := &models.Media{
		UserID:           userID.(primitive.ObjectID),
		Type:             fileType,
		FileName:         header.Filename,
		FileSize:         header.Size,
		MimeType:         header.Header.Get("Content-Type"),
		Caption:          caption,
		AltText:          altText,
		ProcessingStatus: "pending",
		IsProcessed:      false,
	}

	// Upload to media service
	uploadedMedia, err := h.mediaService.Upload(c.Request.Context(), media, file)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to upload media", err)
		return
	}

	// Return success response
	response.Created(c, "Media uploaded successfully", uploadedMedia)
}

// BulkUpload handles multiple file uploads
func (h *UploadHandler) BulkUpload(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse form with files
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil { // 100MB max memory
		response.ValidationError(c, "Failed to parse form", err.Error())
		return
	}

	// Get files from form
	form, err := c.MultipartForm()
	if err != nil {
		response.ValidationError(c, "Failed to get multipart form", err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.ValidationError(c, "No files uploaded", nil)
		return
	}

	// Check if number of files exceeds limit
	if len(files) > 10 {
		response.ValidationError(c, "Maximum 10 files can be uploaded at once", nil)
		return
	}

	// Process each file
	uploadedMedias := make([]*models.Media, 0, len(files))
	for _, fileHeader := range files {
		// Open file
		file, err := fileHeader.Open()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to open file", err)
			return
		}

		// Validate file type
		fileType := validation.GetFileType(fileHeader.Filename)
		if fileType == "unknown" {
			file.Close() // Close file before skipping
			continue     // Skip unsupported files
		}

		// Create media object
		media := &models.Media{
			UserID:           userID.(primitive.ObjectID),
			Type:             fileType,
			FileName:         fileHeader.Filename,
			FileSize:         fileHeader.Size,
			MimeType:         fileHeader.Header.Get("Content-Type"),
			ProcessingStatus: "pending",
			IsProcessed:      false,
		}

		// Upload to media service
		uploadedMedia, err := h.mediaService.Upload(c.Request.Context(), media, file)
		if err != nil {
			file.Close() // Close file before continuing
			continue     // Skip failed uploads
		}

		file.Close() // Close file after processing
		uploadedMedias = append(uploadedMedias, uploadedMedia)
	}

	if len(uploadedMedias) == 0 {
		response.ValidationError(c, "No files were successfully uploaded", nil)
		return
	}

	// Return success response
	response.Created(c, "Media files uploaded successfully", uploadedMedias)
}
