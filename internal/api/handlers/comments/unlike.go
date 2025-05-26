package comments

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnlikeCommentService defines the interface for comment unliking operations
type UnlikeCommentService interface {
	GetCommentByID(commentID primitive.ObjectID) (*models.Comment, error)
	GetLikeByUserAndContent(userID, contentID primitive.ObjectID, contentType string) (*models.Like, error)
	DeleteLike(likeID primitive.ObjectID) error
	DecrementCommentLikeCount(commentID primitive.ObjectID) error
	UpdateReactionCounts(commentID primitive.ObjectID, reactionType string, increment bool) error
}

// UnlikeComment handles removing a like from a comment
func UnlikeComment(c *gin.Context) {
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

	unlikeService := c.MustGet("unlikeCommentService").(UnlikeCommentService)

	// Check if comment exists
	_, err = unlikeService.GetCommentByID(commentID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Comment not found", err)
		return
	}

	// Check if user has liked this comment
	like, err := unlikeService.GetLikeByUserAndContent(userID.(primitive.ObjectID), commentID, "comment")
	if err != nil || like == nil {
		response.Error(c, http.StatusBadRequest, "You have not liked this comment", nil)
		return
	}

	// Store the reaction type before deleting the like
	reactionType := like.ReactionType

	// Delete the like
	if err := unlikeService.DeleteLike(like.ID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unlike comment", err)
		return
	}

	// Decrement comment like count
	if err := unlikeService.DecrementCommentLikeCount(commentID); err != nil {
		// Log the error but don't fail the request
		// TODO: Implement proper logging
		// logger.Error("Failed to decrement comment like count", err)
	}

	// Update reaction counts in the comment
	if err := unlikeService.UpdateReactionCounts(commentID, reactionType, false); err != nil {
		// Log the error but don't fail the request
		// TODO: Implement proper logging
		// logger.Error("Failed to update comment reaction counts", err)
	}

	response.Success(c, http.StatusOK, "Comment unliked successfully", nil)
}
