package comments

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LikeCommentService defines the interface for comment liking operations
type LikeCommentService interface {
	GetCommentByID(commentID primitive.ObjectID) (*models.Comment, error)
	GetLikeByUserAndContent(userID, contentID primitive.ObjectID, contentType string) (*models.Like, error)
	CreateLike(like *models.Like) (primitive.ObjectID, error)
	IncrementCommentLikeCount(commentID primitive.ObjectID) error
	NotifyCommentLike(commentID, likerID primitive.ObjectID) error
	UpdateReactionCounts(commentID primitive.ObjectID, reactionType string, increment bool) error
}

// LikeComment handles liking a comment
func LikeComment(c *gin.Context) {
	commentID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	var req struct {
		ReactionType string `json:"reaction_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to "like" reaction if not specified
		req.ReactionType = "like"
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	likeService := c.MustGet("likeCommentService").(LikeCommentService)

	// Check if comment exists
	comment, err := likeService.GetCommentByID(commentID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Comment not found", err)
		return
	}

	// Check if the comment is hidden or deleted
	if comment.IsHidden || comment.DeletedAt != nil {
		response.Error(c, http.StatusBadRequest, "Cannot like a hidden or deleted comment", nil)
		return
	}

	// Check if user already liked this comment
	existingLike, err := likeService.GetLikeByUserAndContent(userID.(primitive.ObjectID), commentID, "comment")
	if err == nil && existingLike != nil {
		// User already liked this comment
		if existingLike.ReactionType == req.ReactionType {
			response.Error(c, http.StatusBadRequest, "You have already liked this comment with the same reaction", nil)
			return
		}

		// User is changing reaction type - we'll handle this by updating the existing like
		// This functionality would be implemented in a separate handler (UpdateLike)
		response.Error(c, http.StatusBadRequest, "You have already liked this comment. Use the update endpoint to change reaction type", nil)
		return
	}

	// Validate reaction type
	validReactions := map[string]bool{
		"like":   true,
		"love":   true,
		"haha":   true,
		"wow":    true,
		"sad":    true,
		"angry":  true,
		"thanks": true,
	}

	if !validReactions[req.ReactionType] {
		req.ReactionType = "like" // Default to like if invalid
	}

	// Create like model
	like := &models.Like{
		UserID:       userID.(primitive.ObjectID),
		ContentID:    commentID,
		ContentType:  "comment",
		ReactionType: req.ReactionType,
		CreatedAt:    time.Now(),
	}

	// Create the like
	likeID, err := likeService.CreateLike(like)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to like comment", err)
		return
	}

	// Increment comment like count
	if err := likeService.IncrementCommentLikeCount(commentID); err != nil {
		// Log the error but don't fail the request
		// TODO: Implement proper logging
		// logger.Error("Failed to increment comment like count", err)
	}

	// Update reaction counts in the comment
	if err := likeService.UpdateReactionCounts(commentID, req.ReactionType, true); err != nil {
		// Log the error but don't fail the request
		// TODO: Implement proper logging
		// logger.Error("Failed to update comment reaction counts", err)
	}

	// Notify the comment author about the like
	go likeService.NotifyCommentLike(commentID, userID.(primitive.ObjectID))

	response.Success(c, http.StatusOK, "Comment liked successfully", gin.H{
		"like_id": likeID.Hex(),
	})
}
