package hashtags

import (
	"net/http"
	"strconv"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostService defines the interface for retrieving posts by hashtag
type PostService interface {
	GetPostsByHashtag(ctx context.Context, hashtagName string, filter map[string]interface{}, sort map[string]int, limit, offset int) ([]*models.Post, int, error)
}

// GetHashtagPosts retrieves posts that contain a specific hashtag
func GetHashtagPosts(c *gin.Context) {
	// Get hashtag name from URL parameter
	hashtagName := c.Param("hashtag")

	// Remove # prefix if present
	if len(hashtagName) > 0 && hashtagName[0] == '#' {
		hashtagName = hashtagName[1:]
	}

	if hashtagName == "" {
		response.Error(c, http.StatusBadRequest, "Invalid hashtag name", nil)
		return
	}

	// Get query parameters
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	// Parse pagination params
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get the hashtag service to check if it exists and is not restricted
	hashtagService := c.MustGet("hashtagService").(HashtagService)

	// Check if hashtag exists and isn't restricted
	hashtag, err := hashtagService.GetHashtagByName(c.Request.Context(), hashtagName)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Hashtag not found", err)
		return
	}

	// Check if hashtag is restricted and user is not admin
	if hashtag.IsRestricted {
		userID, exists := c.Get("userID")
		if !exists {
			response.Error(c, http.StatusForbidden, "This hashtag is restricted", nil)
			return
		}

		isAdmin, err := hashtagService.IsAdminUser(c.Request.Context(), userID.(primitive.ObjectID))
		if err != nil || !isAdmin {
			response.Error(c, http.StatusForbidden, "This hashtag is restricted", nil)
			return
		}
	}

	// Get the post service
	postService := c.MustGet("postService").(PostService)

	// Build filter
	filter := map[string]interface{}{
		"is_hidden":  false,
		"deleted_at": nil,
		"privacy":    "public", // Only show public posts for hashtag searches
	}

	// Build sort
	sort := map[string]int{}
	if sortOrder == "desc" {
		sort[sortBy] = -1
	} else {
		sort[sortBy] = 1
	}

	// Get posts with the hashtag
	posts, total, err := postService.GetPostsByHashtag(c.Request.Context(), hashtagName, filter, sort, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve hashtag posts", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Hashtag posts retrieved successfully", posts, limit, offset, total)
}
