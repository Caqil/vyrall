package comments

import (
	"fmt"
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ListCommentsService defines the interface for listing comments
type ListCommentsService interface {
	GetPostByID(postID primitive.ObjectID) (*models.Post, error)
	GetCommentsByPostID(postID primitive.ObjectID, sortBy string, limit, offset int) ([]*models.Comment, int, error)
	CheckUserLikedComments(userID primitive.ObjectID, commentIDs []primitive.ObjectID) (map[primitive.ObjectID]bool, error)
	EnrichCommentsWithUserData(comments []*models.Comment) ([]*models.Comment, error)
}

// ListComments handles listing comments for a post
func ListComments(c *gin.Context) {
	postID, err := primitive.ObjectIDFromHex(c.Param("post_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err)
		return
	}

	// Parse query parameters
	sortBy := c.DefaultQuery("sort_by", "recent") // Options: recent, popular
	limit, offset := getPaginationParams(c)

	commentService := c.MustGet("listCommentsService").(ListCommentsService)

	// Check if post exists
	_, err = commentService.GetPostByID(postID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Post not found", err)
		return
	}

	// Get comments for the post
	comments, total, err := commentService.GetCommentsByPostID(postID, sortBy, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve comments", err)
		return
	}

	// Get user ID from authenticated user if available
	userID, exists := c.Get("userID")
	if exists && len(comments) > 0 {
		// Check which comments the user has liked
		var commentIDs []primitive.ObjectID
		for _, comment := range comments {
			commentIDs = append(commentIDs, comment.ID)
		}

		likedComments, err := commentService.CheckUserLikedComments(userID.(primitive.ObjectID), commentIDs)
		if err == nil {
			// Add liked status to comments
			for i, comment := range comments {
				// This assumes we have a way to add the liked status to the comment
				// For example, if the Comment struct has a UserLiked field:
				// comments[i].UserLiked = likedComments[comment.ID]

				// As an alternative, we could create a wrapper DTO that includes this info
				comments[i].UserLiked = likedComments[comment.ID]
			}
		}
	}

	// Enrich comments with user data
	enrichedComments, err := commentService.EnrichCommentsWithUserData(comments)
	if err != nil {
		// Log the error but proceed with the original comments
		// TODO: Implement proper logging
		// logger.Error("Failed to enrich comments with user data", err)
		enrichedComments = comments
	}

	response.SuccessWithPagination(c, http.StatusOK, "Comments retrieved successfully", enrichedComments, limit, offset, total)
}

// Helper function to get pagination parameters
func getPaginationParams(c *gin.Context) (int, int) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit := 10
	if parsedLimit, err := parseInt(limitStr); err == nil && parsedLimit > 0 {
		if parsedLimit > 50 {
			limit = 50 // Cap at 50
		} else {
			limit = parsedLimit
		}
	}

	offset := 0
	if parsedOffset, err := parseInt(offsetStr); err == nil && parsedOffset >= 0 {
		offset = parsedOffset
	}

	return limit, offset
}

// Helper function to parse int from string
func parseInt(str string) (int, error) {
	var value int
	_, err := fmt.Sscanf(str, "%d", &value)
	return value, err
}
