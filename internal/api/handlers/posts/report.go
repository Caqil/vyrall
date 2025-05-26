package posts

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/post"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportHandler handles post reporting operations
type ReportHandler struct {
	postService *post.Service
}

// NewReportHandler creates a new report handler
func NewReportHandler(postService *post.Service) *ReportHandler {
	return &ReportHandler{
		postService: postService,
	}
}

// ReportPost handles the request to report a post
func (h *ReportHandler) ReportPost(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get post ID from URL parameter
	postIDStr := c.Param("id")
	if !validation.IsValidObjectID(postIDStr) {
		response.ValidationError(c, "Invalid post ID", nil)
		return
	}
	postID, _ := primitive.ObjectIDFromHex(postIDStr)

	// Parse request body
	var req struct {
		Reason      string `json:"reason" binding:"required"`
		Description string `json:"description,omitempty"`
		Category    string `json:"category,omitempty"` // spam, inappropriate, harmful, misinformation, etc.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Report the post
	reportID, err := h.postService.ReportPost(
		c.Request.Context(),
		postID,
		userID.(primitive.ObjectID),
		req.Reason,
		req.Description,
		req.Category,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to report post", err)
		return
	}

	// Return success response
	response.Created(c, "Post reported successfully", gin.H{
		"report_id": reportID.Hex(),
	})
}

// ReportComment handles the request to report a comment
func (h *ReportHandler) ReportComment(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get comment ID from URL parameter
	commentIDStr := c.Param("id")
	if !validation.IsValidObjectID(commentIDStr) {
		response.ValidationError(c, "Invalid comment ID", nil)
		return
	}
	commentID, _ := primitive.ObjectIDFromHex(commentIDStr)

	// Parse request body
	var req struct {
		Reason      string `json:"reason" binding:"required"`
		Description string `json:"description,omitempty"`
		Category    string `json:"category,omitempty"` // spam, inappropriate, harmful, misinformation, etc.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Report the comment
	reportID, err := h.postService.ReportComment(
		c.Request.Context(),
		commentID,
		userID.(primitive.ObjectID),
		req.Reason,
		req.Description,
		req.Category,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to report comment", err)
		return
	}

	// Return success response
	response.Created(c, "Comment reported successfully", gin.H{
		"report_id": reportID.Hex(),
	})
}

// GetReportStatus handles the request to get the status of a report
func (h *ReportHandler) GetReportStatus(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get report ID from URL parameter
	reportIDStr := c.Param("id")
	if !validation.IsValidObjectID(reportIDStr) {
		response.ValidationError(c, "Invalid report ID", nil)
		return
	}
	reportID, _ := primitive.ObjectIDFromHex(reportIDStr)

	// Get report status
	report, err := h.postService.GetReportStatus(c.Request.Context(), reportID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get report status", err)
		return
	}

	// Return success response
	response.OK(c, "Report status retrieved successfully", report)
}

// ListUserReports handles the request to list reports submitted by the user
func (h *ReportHandler) ListUserReports(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get query parameters
	status := c.DefaultQuery("status", "") // pending, approved, rejected, all

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// List user reports
	reports, total, err := h.postService.ListUserReports(c.Request.Context(), userID.(primitive.ObjectID), status, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list reports", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "User reports retrieved successfully", reports, limit, offset, total)
}
