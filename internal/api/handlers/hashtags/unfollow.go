package hashtags

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnfollowService defines the interface for unfollowing hashtags
type UnfollowService interface {
	UnfollowHashtag(ctx context.Context, userID, hashtagID primitive.ObjectID) error
	IsFollowing(ctx context.Context, userID, hashtagID primitive.ObjectID) (bool, error)
}

// UnfollowHashtag handles a user unfollowing a hashtag
func UnfollowHashtag(c *gin.Context) {
	// Get hashtag ID or name from URL parameter
	hashtagParam := c.Param("hashtag")

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the hashtag service
	hashtagService := c.MustGet("hashtagService").(HashtagService)

	// Get the unfollow service
	unfollowService := c.MustGet("unfollowService").(UnfollowService)

	// Find the hashtag by ID or name
	var hashtagID primitive.ObjectID
	var err error
	var hashtag *models.Hashtag

	// Try to parse as ObjectID first
	if primitive.IsValidObjectID(hashtagParam) {
		hashtagID, _ = primitive.ObjectIDFromHex(hashtagParam)
		hashtag, err = hashtagService.GetHashtagByID(c.Request.Context(), hashtagID)
	} else {
		// If not a valid ObjectID, treat as hashtag name
		hashtag, err = hashtagService.GetHashtagByName(c.Request.Context(), hashtagParam)
		if err == nil {
			hashtagID = hashtag.ID
		}
	}

	if err != nil {
		response.Error(c, http.StatusNotFound, "Hashtag not found", err)
		return
	}

	// Check if the user is following this hashtag
	isFollowing, err := unfollowService.IsFollowing(c.Request.Context(), userID.(primitive.ObjectID), hashtagID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check follow status", err)
		return
	}

	if !isFollowing {
		response.Error(c, http.StatusBadRequest, "You are not following this hashtag", nil)
		return
	}

	// Unfollow the hashtag
	err = unfollowService.UnfollowHashtag(c.Request.Context(), userID.(primitive.ObjectID), hashtagID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unfollow hashtag", err)
		return
	}

	response.Success(c, http.StatusOK, "Hashtag unfollowed successfully", gin.H{
		"hashtag_id":     hashtagID.Hex(),
		"hashtag_name":   hashtag.Name,
		"follower_count": hashtag.FollowerCount - 1, // Optimistically decrement
	})
}
