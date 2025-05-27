package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExportDataHandler handles user data export operations
type ExportDataHandler struct {
	userService *user.Service
}

// NewExportDataHandler creates a new export data handler
func NewExportDataHandler(userService *user.Service) *ExportDataHandler {
	return &ExportDataHandler{
		userService: userService,
	}
}

// RequestDataExport handles the request to export user data
func (h *ExportDataHandler) RequestDataExport(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		DataTypes []string `json:"data_types,omitempty"` // profile, posts, media, etc.
		Format    string   `json:"format,omitempty"`     // json, csv, html
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Set default format if not provided
	if req.Format == "" {
		req.Format = "json"
	} else if req.Format != "json" && req.Format != "csv" && req.Format != "html" {
		response.ValidationError(c, "Invalid format. Must be 'json', 'csv', or 'html'", nil)
		return
	}

	// Request data export
	exportID, err := h.userService.RequestDataExport(c.Request.Context(), userID.(primitive.ObjectID), req.DataTypes, req.Format)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to request data export", err)
		return
	}

	// Return success response
	response.OK(c, "Data export requested successfully", gin.H{
		"export_id": exportID.Hex(),
		"status":    "pending",
		"message":   "You will be notified when your data export is ready for download",
	})
}

// GetDataExportStatus handles the request to get the status of a data export
func (h *ExportDataHandler) GetDataExportStatus(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get export ID from URL parameter
	exportIDStr := c.Param("id")
	if exportIDStr == "" {
		// If no specific export ID, get all exports
		exports, err := h.userService.GetUserDataExports(c.Request.Context(), userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to retrieve data exports", err)
			return
		}

		// Return success response
		response.OK(c, "Data exports retrieved successfully", exports)
		return
	}

	// If specific export ID provided, validate and get status
	if !validation.IsValidObjectID(exportIDStr) {
		response.ValidationError(c, "Invalid export ID", nil)
		return
	}
	exportID, _ := primitive.ObjectIDFromHex(exportIDStr)

	// Get export status
	status, err := h.userService.GetDataExportStatus(c.Request.Context(), userID.(primitive.ObjectID), exportID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get export status", err)
		return
	}

	// Return success response
	response.OK(c, "Export status retrieved successfully", status)
}

// DownloadDataExport handles the request to download an exported data file
func (h *ExportDataHandler) DownloadDataExport(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get export ID from URL parameter
	exportIDStr := c.Param("id")
	if !validation.IsValidObjectID(exportIDStr) {
		response.ValidationError(c, "Invalid export ID", nil)
		return
	}
	exportID, _ := primitive.ObjectIDFromHex(exportIDStr)

	// Get export file
	fileData, fileName, contentType, err := h.userService.DownloadDataExport(c.Request.Context(), userID.(primitive.ObjectID), exportID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to download data export", err)
		return
	}

	// Set headers for file download
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, fileData)
}

// DeleteDataExport handles the request to delete a data export
func (h *ExportDataHandler) DeleteDataExport(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get export ID from URL parameter
	exportIDStr := c.Param("id")
	if !validation.IsValidObjectID(exportIDStr) {
		response.ValidationError(c, "Invalid export ID", nil)
		return
	}
	exportID, _ := primitive.ObjectIDFromHex(exportIDStr)

	// Delete the export
	err := h.userService.DeleteDataExport(c.Request.Context(), userID.(primitive.ObjectID), exportID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete data export", err)
		return
	}

	// Return success response
	response.OK(c, "Data export deleted successfully", nil)
}
