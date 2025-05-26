package comments

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RepliesService defines the interface for comment replies operations
type RepliesService interface {
	GetCommentByID(commentID primitive.ObjectID) (*models.Comment, error)
	GetCommentReplies(commentID primitive.ObjectID, sortBy string, limit, offset int) ([]*models.Comment, int, error)
	CheckUserLikedComments(userID primitive.ObjectID, commentIDs []primitive.ObjectID) (map[primitive.ObjectID]bool, error)
	EnrichCommentsWithUserData(comments []*models.Comment) ([]*models.Comment, error)
}

// GetReplies handles fetching replies to a comment
func GetReplies(c *gin.Context) {
	commentID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	// Parse query parameters
	sortBy := c.DefaultQuery("sort_by", "recent") // Options: recent, popular
	limit, offset := getPaginationParams(c)

	repliesService := c.MustGet("repliesService").(RepliesService)

	// Check if comment exists
	comment, err := repliesService.GetCommentByID(commentID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Comment not found", err)
		return
	}

	// Check if comment has replies
	if comment.ReplyCount == 0 {
		response.SuccessWithPagination(c, http.StatusOK, "No replies found", []interface{}{}, limit, offset, 0)
		return
	}

	// Get replies for the comment
	replies, total, err := repliesService.GetCommentReplies(commentID, sortBy, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve replies", err)
		return
	}

	// Get user ID from authenticated user if available
	userID, exists := c.Get("userID")
	if exists && len(replies) > 0 {
		// Check which replies the user has liked
		var replyIDs []primitive.ObjectID
		for _, reply := range replies {
			replyIDs = append(replyIDs, reply.ID)
		}

		likedReplies, err := repliesService.CheckUserLikedComments(userID.(primitive.ObjectID), replyIDs)
		if err == nil {
			// Add liked status to replies
			for i, reply := range replies {
				// This assumes we have a way to add the liked status to the reply
				// As in the list.go file, we assume there's a UserLiked field
				replies[i].UserLiked = likedReplies[reply.ID]
			}
		}
	}

	// Enrich replies with user data
	enrichedReplies, err := repliesService.EnrichCommentsWithUserData(replies)
	if err != nil {
		// Log the error but proceed with the original replies
		// TODO: Implement proper logging
		// logger.Error("Failed to enrich replies with user data", err)
		enrichedReplies = replies
	}

	response.SuccessWithPagination(c, http.StatusOK, "Replies retrieved successfully", enrichedReplies, limit, offset, total)
}
