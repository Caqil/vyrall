package messages

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BackupHandler handles message backup operations
type BackupHandler struct {
	messageService *message.Service
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(messageService *message.Service) *BackupHandler {
	return &BackupHandler{
		messageService: messageService,
	}
}

// CreateBackup handles the request to create a backup of conversations
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		ConversationIDs []string `json:"conversation_ids,omitempty"` // If empty, backup all conversations
		IncludeMedia    bool     `json:"include_media,omitempty"`
		Format          string   `json:"format,omitempty"` // json, csv, html
		DateRange       struct {
			StartDate string `json:"start_date,omitempty"`
			EndDate   string `json:"end_date,omitempty"`
		} `json:"date_range,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Set default format if not provided
	if req.Format == "" {
		req.Format = "json"
	} else if req.Format != "json" && req.Format != "csv" && req.Format != "html" {
		response.ValidationError(c, "Invalid format. Supported formats: json, csv, html", nil)
		return
	}

	// Convert conversation IDs to ObjectIDs if provided
	var conversationIDs []primitive.ObjectID
	if len(req.ConversationIDs) > 0 {
		conversationIDs = make([]primitive.ObjectID, 0, len(req.ConversationIDs))
		for _, idStr := range req.ConversationIDs {
			if !validation.IsValidObjectID(idStr) {
				continue // Skip invalid IDs
			}
			convID, _ := primitive.ObjectIDFromHex(idStr)
			conversationIDs = append(conversationIDs, convID)
		}
	}

	// Parse date range if provided
	var startDate, endDate time.Time
	if req.DateRange.StartDate != "" {
		if parsed, err := time.Parse(time.RFC3339, req.DateRange.StartDate); err == nil {
			startDate = parsed
		}
	}
	if req.DateRange.EndDate != "" {
		if parsed, err := time.Parse(time.RFC3339, req.DateRange.EndDate); err == nil {
			endDate = parsed
		} else {
			endDate = time.Now()
		}
	}

	// Create backup
	backupResult, err := h.messageService.CreateBackup(
		c.Request.Context(),
		userID.(primitive.ObjectID),
		conversationIDs,
		req.IncludeMedia,
		req.Format,
		startDate,
		endDate,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create backup", err)
		return
	}

	// Return success response
	response.OK(c, "Backup created successfully", gin.H{
		"backup_id":    backupResult.BackupID,
		"status":       backupResult.Status,
		"created_at":   backupResult.CreatedAt,
		"file_size":    backupResult.FileSize,
		"file_format":  backupResult.Format,
		"expires_at":   backupResult.ExpiresAt,
		"download_url": backupResult.DownloadURL,
	})
}

// GetBackupStatus handles the request to get the status of a backup
func (h *BackupHandler) GetBackupStatus(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get backup ID from URL parameter
	backupID := c.Param("id")
	if backupID == "" {
		response.ValidationError(c, "Backup ID is required", nil)
		return
	}

	// Get backup status
	backupStatus, err := h.messageService.GetBackupStatus(c.Request.Context(), backupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get backup status", err)
		return
	}

	// Return success response
	response.OK(c, "Backup status retrieved successfully", backupStatus)
}

// DownloadBackup handles the request to download a backup
func (h *BackupHandler) DownloadBackup(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get backup ID from URL parameter
	backupID := c.Param("id")
	if backupID == "" {
		response.ValidationError(c, "Backup ID is required", nil)
		return
	}

	// Get backup file
	backupFile, fileName, contentType, err := h.messageService.DownloadBackup(c.Request.Context(), backupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to download backup", err)
		return
	}

	// Set response headers
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", contentType)

	// Stream the file
	c.Data(http.StatusOK, contentType, backupFile)
}

// ListBackups handles the request to list all backups
func (h *BackupHandler) ListBackups(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// List backups
	backups, total, err := h.messageService.ListBackups(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list backups", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Backups retrieved successfully", backups, limit, offset, total)
}

// DeleteBackup handles the request to delete a backup
func (h *BackupHandler) DeleteBackup(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get backup ID from URL parameter
	backupID := c.Param("id")
	if backupID == "" {
		response.ValidationError(c, "Backup ID is required", nil)
		return
	}

	// Delete backup
	err := h.messageService.DeleteBackup(c.Request.Context(), backupID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete backup", err)
		return
	}

	// Return success response
	response.OK(c, "Backup deleted successfully", nil)
}
