package comments

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteCommentService defines the interface for comment deletion
type DeleteCommentService interface {
	GetCommentByID(commentID primitive.ObjectID) (*models.Comment, error)
	DeleteComment(commentID primitive.ObjectID) error
	SoftDeleteComment(commentID primitive.ObjectID) error
	DecrementPostCommentCount(postID primitive.ObjectID) error
	DecrementCommentReplyCount(commentID primitive.ObjectID) error
	GetCommentReplies(commentID primitive.ObjectID) ([]primitive.ObjectID, error)
	DeleteReplies(commentIDs []primitive.ObjectID) error
}

// DeleteComment handles the deletion of a comment
func DeleteComment(c *gin.Context) {
	commentID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	// Get user role
	userRole, _ := c.Get("userRole")
	isAdmin := userRole == "admin" || userRole == "moderator"

	commentService := c.MustGet("deleteCommentService").(DeleteCommentService)

	// Get the comment
	comment, err := commentService.GetCommentByID(commentID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Comment not found", err)
		return
	}

	// Check if the user is authorized to delete the comment
	if !isAdmin && comment.UserID != userID.(primitive.ObjectID) {
		response.Error(c, http.StatusForbidden, "You don't have permission to delete this comment", nil)
		return
	}

	// Check for permanent deletion flag
	permanentDelete := c.DefaultQuery("permanent", "false") == "true"

	// Get deletion method - hard or soft
	var deleteErr error
	if permanentDelete && isAdmin {
		// Only admins can permanently delete comments
		deleteErr = commentService.DeleteComment(commentID)
	} else {
		// Regular users or admins who didn't specify permanent deletion
		deleteErr = commentService.SoftDeleteComment(commentID)
	}

	if deleteErr != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete comment", deleteErr)
		return
	}

	// Update comment counts
	if comment.ParentID != nil {
		// This was a reply, decrement the parent comment's reply count
		if err := commentService.DecrementCommentReplyCount(*comment.ParentID); err != nil {
			// Log the error but don't fail the request
			// TODO: Implement proper logging
			// logger.Error("Failed to decrement comment reply count", err)
		}
	} else {
		// This was a top-level comment, decrement the post's comment count
		if err := commentService.DecrementPostCommentCount(comment.PostID); err != nil {
			// Log the error but don't fail the request
			// TODO: Implement proper logging
			// logger.Error("Failed to decrement post comment count", err)
		}
	}

	// If we're permanently deleting a parent comment, also delete all replies
	if permanentDelete && isAdmin && comment.ReplyCount > 0 {
		// Get all replies to this comment
		replyIDs, err := commentService.GetCommentReplies(commentID)
		if err == nil && len(replyIDs) > 0 {
			// Delete all replies
			if err := commentService.DeleteReplies(replyIDs); err != nil {
				// Log the error but don't fail the request
				// TODO: Implement proper logging
				// logger.Error("Failed to delete comment replies", err)
			}
		}
	}

	response.Success(c, http.StatusOK, "Comment deleted successfully", nil)
}
