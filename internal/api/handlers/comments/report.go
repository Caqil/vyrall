package comments

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportCommentService defines the interface for reporting comments
type ReportCommentService interface {
	GetCommentByID(commentID primitive.ObjectID) (*models.Comment, error)
	CreateReport(report *models.Report) (primitive.ObjectID, error)
	GetExistingReport(reporterID, contentID primitive.ObjectID, contentType string) (*models.Report, error)
	NotifyModerators(report *models.Report) error
}

// ReportCommentRequest represents a request to report a comment
type ReportCommentRequest struct {
	ReasonCode  string `json:"reason_code" binding:"required"`
	Description string `json:"description"`
}

// ReportComment handles reporting a comment for violation
func ReportComment(c *gin.Context) {
	commentID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	var req ReportCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	reportService := c.MustGet("reportCommentService").(ReportCommentService)

	// Check if comment exists
	comment, err := reportService.GetCommentByID(commentID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Comment not found", err)
		return
	}

	// Check if comment is already deleted or hidden
	if comment.DeletedAt != nil || comment.IsHidden {
		response.Error(c, http.StatusBadRequest, "Cannot report a deleted or hidden comment", nil)
		return
	}

	// Check if user has already reported this comment
	existingReport, err := reportService.GetExistingReport(userID.(primitive.ObjectID), commentID, "comment")
	if err == nil && existingReport != nil {
		response.Error(c, http.StatusBadRequest, "You have already reported this comment", nil)
		return
	}

	// Validate reason code
	validReasonCodes := map[string]bool{
		"spam":                  true,
		"harassment":            true,
		"hate_speech":           true,
		"inappropriate":         true,
		"misinformation":        true,
		"violence":              true,
		"illegal_content":       true,
		"intellectual_property": true,
		"other":                 true,
	}

	if !validReasonCodes[req.ReasonCode] {
		response.Error(c, http.StatusBadRequest, "Invalid reason code", nil)
		return
	}

	// Create report model
	report := &models.Report{
		ReporterID:  userID.(primitive.ObjectID),
		ContentID:   commentID,
		ContentType: "comment",
		ReasonCode:  req.ReasonCode,
		Description: req.Description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create the report
	reportID, err := reportService.CreateReport(report)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to report comment", err)
		return
	}

	// Notify moderators about the report
	go reportService.NotifyModerators(report)

	response.Success(c, http.StatusOK, "Comment reported successfully", gin.H{
		"report_id": reportID.Hex(),
	})
}
