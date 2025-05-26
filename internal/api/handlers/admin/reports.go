package admin

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportService interface for report operations
type ReportService interface {
	GetReports(status, contentType string, limit, offset int) ([]map[string]interface{}, int, error)
	GetReportByID(id primitive.ObjectID) (map[string]interface{}, error)
	UpdateReportStatus(id primitive.ObjectID, status, notes string, moderatorID primitive.ObjectID) error
	BulkUpdateReportStatus(ids []primitive.ObjectID, status string, moderatorID primitive.ObjectID) (int, error)
	GetReportStats(startDate, endDate time.Time) (map[string]interface{}, error)
	GetReportsByUser(userID primitive.ObjectID, limit, offset int) ([]map[string]interface{}, int, error)
	GetReportsByContentID(contentType string, contentID primitive.ObjectID, limit, offset int) ([]map[string]interface{}, int, error)
	GetReportsByReasonCode(reasonCode string, limit, offset int) ([]map[string]interface{}, int, error)
}

// GetReports returns a list of user reports
func GetReports(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	contentType := c.DefaultQuery("content_type", "all")
	limit, offset := getPaginationParams(c)

	reportService := c.MustGet("reportService").(ReportService)

	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"reviewed":  true,
		"actioned":  true,
		"dismissed": true,
		"all":       true,
	}

	if !validStatuses[status] {
		response.Error(c, http.StatusBadRequest, "Invalid status", nil)
		return
	}

	reports, total, err := reportService.GetReports(status, contentType, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve reports", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Reports retrieved successfully", reports, limit, offset, total)
}

// GetReportDetail returns details about a specific report
func GetReportDetail(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid report ID", err)
		return
	}

	reportService := c.MustGet("reportService").(ReportService)

	report, err := reportService.GetReportByID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve report details", err)
		return
	}

	response.Success(c, http.StatusOK, "Report details retrieved successfully", report)
}

// UpdateReportStatus updates the status of a report
func UpdateReportStatus(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid report ID", err)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get moderator ID from authenticated user
	moderatorID := c.MustGet("userID").(primitive.ObjectID)
	reportService := c.MustGet("reportService").(ReportService)

	// Validate status
	validStatuses := map[string]bool{
		"reviewed":  true,
		"actioned":  true,
		"dismissed": true,
	}

	if !validStatuses[req.Status] {
		response.Error(c, http.StatusBadRequest, "Invalid status", nil)
		return
	}

	if err := reportService.UpdateReportStatus(id, req.Status, req.Notes, moderatorID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update report status", err)
		return
	}

	response.Success(c, http.StatusOK, "Report status updated successfully", nil)
}

// BulkUpdateReportStatus updates the status of multiple reports
func BulkUpdateReportStatus(c *gin.Context) {
	var req struct {
		ReportIDs []string `json:"report_ids" binding:"required"`
		Status    string   `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"reviewed":  true,
		"actioned":  true,
		"dismissed": true,
	}

	if !validStatuses[req.Status] {
		response.Error(c, http.StatusBadRequest, "Invalid status", nil)
		return
	}

	// Get moderator ID from authenticated user
	moderatorID := c.MustGet("userID").(primitive.ObjectID)
	reportService := c.MustGet("reportService").(ReportService)

	// Convert string IDs to ObjectIDs
	var reportIDs []primitive.ObjectID
	for _, idStr := range req.ReportIDs {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid report ID: "+idStr, err)
			return
		}
		reportIDs = append(reportIDs, id)
	}

	updatedCount, err := reportService.BulkUpdateReportStatus(reportIDs, req.Status, moderatorID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update report statuses", err)
		return
	}

	response.Success(c, http.StatusOK, "Reports updated successfully", gin.H{
		"updated_count": updatedCount,
	})
}

// GetReportStats returns statistics about user reports
func GetReportStats(c *gin.Context) {
	// Parse date parameters with defaults (last 30 days)
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid end date format", err)
		return
	}

	reportService := c.MustGet("reportService").(ReportService)

	stats, err := reportService.GetReportStats(startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve report statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Report statistics retrieved successfully", stats)
}

// GetReportsByUser returns reports created by or about a specific user
func GetReportsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	limit, offset := getPaginationParams(c)
	reportService := c.MustGet("reportService").(ReportService)

	reports, total, err := reportService.GetReportsByUser(userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user reports", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "User reports retrieved successfully", reports, limit, offset, total)
}

// GetReportsByContent returns reports about a specific content item
func GetReportsByContent(c *gin.Context) {
	contentType := c.Param("type")
	contentIDStr := c.Param("content_id")

	contentID, err := primitive.ObjectIDFromHex(contentIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid content ID", err)
		return
	}

	limit, offset := getPaginationParams(c)
	reportService := c.MustGet("reportService").(ReportService)

	reports, total, err := reportService.GetReportsByContentID(contentType, contentID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve content reports", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Content reports retrieved successfully", reports, limit, offset, total)
}

// GetReportsByReason returns reports with a specific reason code
func GetReportsByReason(c *gin.Context) {
	reasonCode := c.Param("reason_code")
	limit, offset := getPaginationParams(c)

	reportService := c.MustGet("reportService").(ReportService)

	reports, total, err := reportService.GetReportsByReasonCode(reasonCode, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve reports by reason", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Reports retrieved successfully", reports, limit, offset, total)
}
