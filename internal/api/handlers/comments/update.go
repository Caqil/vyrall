package comments

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateCommentService defines the interface for comment updating operations
type UpdateCommentService interface {
	GetCommentByID(commentID primitive.ObjectID) (*models.Comment, error)
	UpdateComment(comment *models.Comment) error
	NotifyMentionedUsers(comment *models.Comment) error
	AddToEditHistory(commentID primitive.ObjectID, editRecord models.EditRecord) error
}

// UpdateCommentRequest represents a request to update a comment
type UpdateCommentRequest struct {
	Content        string   `json:"content" binding:"required"`
	MentionedUsers []string `json:"mentioned_users,omitempty"`
	MediaFiles     []string `json:"media_files,omitempty"`
	EditReason     string   `json:"edit_reason,omitempty"`
}

// UpdateComment handles updating an existing comment
func UpdateComment(c *gin.Context) {
	commentID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	var req UpdateCommentRequest
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

	updateService := c.MustGet("updateCommentService").(UpdateCommentService)

	// Get the comment
	comment, err := updateService.GetCommentByID(commentID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Comment not found", err)
		return
	}

	// Check if the user is authorized to update the comment
	if comment.UserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to update this comment", nil)
		return
	}

	// Check if the comment is deleted
	if comment.DeletedAt != nil {
		response.Error(c, http.StatusBadRequest, "Cannot update a deleted comment", nil)
		return
	}

	// Check if the comment is hidden
	if comment.IsHidden {
		response.Error(c, http.StatusBadRequest, "Cannot update a hidden comment", nil)
		return
	}

	// Process mentioned users
	var mentionedUsers []primitive.ObjectID
	for _, userIDStr := range req.MentionedUsers {
		mentionedUserID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			continue // Skip invalid IDs
		}
		mentionedUsers = append(mentionedUsers, mentionedUserID)
	}

	// Process media files
	var mediaFiles []models.Media
	for _, mediaIDStr := range req.MediaFiles {
		mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
		if err != nil {
			continue // Skip invalid IDs
		}
		mediaFiles = append(mediaFiles, models.Media{ID: mediaID})
	}

	// Create edit record
	editRecord := models.EditRecord{
		Content:  comment.Content, // Store the previous content
		EditedAt: time.Now(),
		EditorID: userID.(primitive.ObjectID),
		Reason:   req.EditReason,
	}

	// Add to edit history
	if err := updateService.AddToEditHistory(commentID, editRecord); err != nil {
		// Log the error but don't fail the request
		// TODO: Implement proper logging
		// logger.Error("Failed to add to edit history", err)
	}

	// Update comment fields
	comment.Content = req.Content
	comment.MentionedUsers = mentionedUsers
	comment.MediaFiles = mediaFiles
	comment.IsEdited = true
	comment.UpdatedAt = time.Now()

	// Update the comment
	if err := updateService.UpdateComment(comment); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update comment", err)
		return
	}

	// Notify newly mentioned users
	if len(mentionedUsers) > 0 {
		go updateService.NotifyMentionedUsers(comment)
	}

	response.Success(c, http.StatusOK, "Comment updated successfully", comment)
}
