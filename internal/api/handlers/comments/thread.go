package comments

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ThreadService defines the interface for comment thread operations
type ThreadService interface {
	GetCommentByID(commentID primitive.ObjectID) (*models.Comment, error)
	GetCommentThread(rootID primitive.ObjectID, limit, offset int) ([]*models.Comment, int, error)
	GetThreadParents(commentID primitive.ObjectID) ([]*models.Comment, error)
	EnrichCommentsWithUserData(comments []*models.Comment) ([]*models.Comment, error)
	CheckUserLikedComments(userID primitive.ObjectID, commentIDs []primitive.ObjectID) (map[primitive.ObjectID]bool, error)
}

// GetCommentThread handles fetching a thread of comments
func GetCommentThread(c *gin.Context) {
	commentID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	// Parse query parameters
	limit, offset := getPaginationParams(c)
	includeParents := c.DefaultQuery("include_parents", "true") == "true"

	threadService := c.MustGet("threadService").(ThreadService)

	// Check if comment exists
	comment, err := threadService.GetCommentByID(commentID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Comment not found", err)
		return
	}

	// Prepare thread data structure
	threadData := make(map[string]interface{})

	// Get parents if requested and if this is a nested comment
	if includeParents && comment.ParentID != nil {
		// Get all parents in the thread hierarchy
		parents, err := threadService.GetThreadParents(commentID)
		if err != nil {
			// Log the error but continue
			// TODO: Implement proper logging
			// logger.Error("Failed to retrieve thread parents", err)
		} else {
			// Enrich parents with user data
			enrichedParents, err := threadService.EnrichCommentsWithUserData(parents)
			if err != nil {
				// Log the error but use original parents
				// TODO: Implement proper logging
				// logger.Error("Failed to enrich parents with user data", err)
				enrichedParents = parents
			}

			// Add parent comments to thread data
			threadData["parents"] = enrichedParents
		}
	}

	// Get the thread (replies to this comment and potentially nested replies)
	thread, total, err := threadService.GetCommentThread(commentID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve comment thread", err)
		return
	}

	// Get user ID from authenticated user if available
	userID, exists := c.Get("userID")
	if exists && len(thread) > 0 {
		// Check which comments the user has liked
		var threadIDs []primitive.ObjectID
		for _, comment := range thread {
			threadIDs = append(threadIDs, comment.ID)
		}

		// Add the root comment ID if we're including it
		threadIDs = append(threadIDs, commentID)

		likedComments, err := threadService.CheckUserLikedComments(userID.(primitive.ObjectID), threadIDs)
		if err == nil {
			// Add liked status to thread comments
			for i, comment := range thread {
				thread[i].UserLiked = likedComments[comment.ID]
			}

			// Add liked status to root comment
			comment.UserLiked = likedComments[commentID]
		}
	}

	// Enrich thread with user data
	enrichedThread, err := threadService.EnrichCommentsWithUserData(thread)
	if err != nil {
		// Log the error but proceed with original thread
		// TODO: Implement proper logging
		// logger.Error("Failed to enrich thread with user data", err)
		enrichedThread = thread
	}

	// Add the root comment with user data
	enrichedRoot, err := threadService.EnrichCommentsWithUserData([]*models.Comment{comment})
	if err != nil || len(enrichedRoot) == 0 {
		// Log the error but use original comment
		// TODO: Implement proper logging
		// logger.Error("Failed to enrich root comment with user data", err)
		threadData["root_comment"] = comment
	} else {
		threadData["root_comment"] = enrichedRoot[0]
	}

	// Add the thread to the response
	threadData["thread"] = enrichedThread
	threadData["total"] = total
	threadData["limit"] = limit
	threadData["offset"] = offset

	response.Success(c, http.StatusOK, "Comment thread retrieved successfully", threadData)
}
