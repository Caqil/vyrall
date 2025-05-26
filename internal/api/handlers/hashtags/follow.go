package hashtags

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HashtagFollowService defines the interface for hashtag follow operations
type HashtagFollowService interface {
	FollowHashtag(ctx context.Context, userID, hashtagID primitive.ObjectID) (primitive.ObjectID, error)
	IsFollowing(ctx context.Context, userID, hashtagID primitive.ObjectID) (bool, error)
	GetFollowedHashtags(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*models.Hashtag, int, error)
}

// FollowHashtag handles a user following a hashtag
func FollowHashtag(c *gin.Context) {
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

	// Get the follow service
	followService := c.MustGet("hashtagFollowService").(HashtagFollowService)

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
		// If hashtag doesn't exist, create it (optional depending on your requirements)
		if hashtagParam != "" && !primitive.IsValidObjectID(hashtagParam) {
			newHashtag := &models.Hashtag{
				Name:          hashtagParam,
				PostCount:     0,
				FollowerCount: 0,
				IsTrending:    false,
				IsRestricted:  false,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			hashtagID, err = hashtagService.CreateHashtag(c.Request.Context(), newHashtag)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to create hashtag", err)
				return
			}

			hashtag = newHashtag
			hashtag.ID = hashtagID
		} else {
			response.Error(c, http.StatusNotFound, "Hashtag not found", err)
			return
		}
	}

	// Check if the user is already following this hashtag
	isFollowing, err := followService.IsFollowing(c.Request.Context(), userID.(primitive.ObjectID), hashtagID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check follow status", err)
		return
	}

	if isFollowing {
		response.Error(c, http.StatusBadRequest, "You are already following this hashtag", nil)
		return
	}

	// Follow the hashtag
	followID, err := followService.FollowHashtag(c.Request.Context(), userID.(primitive.ObjectID), hashtagID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to follow hashtag", err)
		return
	}

	response.Success(c, http.StatusOK, "Hashtag followed successfully", gin.H{
		"follow_id":      followID.Hex(),
		"hashtag_id":     hashtagID.Hex(),
		"hashtag_name":   hashtag.Name,
		"follower_count": hashtag.FollowerCount + 1, // Optimistically increment
	})
}

// GetFollowedHashtags retrieves all hashtags followed by the user
func GetFollowedHashtags(c *gin.Context) {
	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get the follow service
	followService := c.MustGet("hashtagFollowService").(HashtagFollowService)

	// Get followed hashtags
	hashtags, total, err := followService.GetFollowedHashtags(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve followed hashtags", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Followed hashtags retrieved successfully", hashtags, limit, offset, total)
}
