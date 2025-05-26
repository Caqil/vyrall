package live

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

// CommentService defines the interface for live stream comment operations
type CommentService interface {
	AddComment(ctx context.Context, comment *models.LiveStreamComment) (primitive.ObjectID, error)
	GetComments(ctx context.Context, streamID primitive.ObjectID, limit, offset int) ([]*models.LiveStreamComment, int, error)
	DeleteComment(ctx context.Context, commentID, userID primitive.ObjectID) error
	PinComment(ctx context.Context, commentID, streamID, userID primitive.ObjectID, isPinned bool) error
	GetPinnedComments(ctx context.Context, streamID primitive.ObjectID) ([]*models.LiveStreamComment, error)
	IsCommentOwner(ctx context.Context, commentID, userID primitive.ObjectID) (bool, error)
}

// AddComment handles adding a comment to a live stream
func AddComment(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse request body
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate comment content
	if len(req.Content) < 1 || len(req.Content) > 500 {
		response.Error(c, http.StatusBadRequest, "Comment must be between 1 and 500 characters", nil)
		return
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists and is active
	stream, err := liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	if stream.Status != "active" {
		response.Error(c, http.StatusBadRequest, "Cannot comment on an inactive stream", nil)
		return
	}

	// Check if the user is banned from the stream
	isBanned, err := liveStreamService.IsUserBanned(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is banned", err)
		return
	}

	if isBanned {
		response.Error(c, http.StatusForbidden, "You are banned from commenting on this stream", nil)
		return
	}

	// Get the comment service
	commentService := c.MustGet("commentService").(CommentService)

	// Create new comment
	comment := &models.LiveStreamComment{
		StreamID:  streamID,
		UserID:    userID.(primitive.ObjectID),
		Content:   req.Content,
		CreatedAt: time.Now(),
		IsHidden:  false,
		IsPinned:  false,
	}

	// Add the comment
	commentID, err := commentService.AddComment(c.Request.Context(), comment)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to add comment", err)
		return
	}

	// Get the user service to include user details
	userService := c.MustGet("userService").(UserService)
	user, err := userService.GetUserByID(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		// If user details retrieval fails, still return the comment
		c.Error(err)
		response.Success(c, http.StatusCreated, "Comment added successfully", gin.H{
			"comment_id": commentID.Hex(),
			"stream_id":  streamID.Hex(),
			"content":    req.Content,
			"created_at": comment.CreatedAt,
		})
		return
	}

	// Return the comment with user details
	response.Success(c, http.StatusCreated, "Comment added successfully", gin.H{
		"comment_id":    commentID.Hex(),
		"stream_id":     streamID.Hex(),
		"content":       req.Content,
		"created_at":    comment.CreatedAt,
		"user_id":       user.ID.Hex(),
		"username":      user.Username,
		"display_name":  user.DisplayName,
		"profile_image": user.ProfileImage,
	})
}

// GetComments returns comments for a live stream
func GetComments(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	} else if limit > 100 {
		limit = 100 // Cap at 100 for performance
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists
	_, err = liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Get the comment service
	commentService := c.MustGet("commentService").(CommentService)

	// Get the comments
	comments, total, err := commentService.GetComments(c.Request.Context(), streamID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve comments", err)
		return
	}

	// If there are no comments, return an empty array
	if len(comments) == 0 {
		response.SuccessWithPagination(c, http.StatusOK, "No comments found", []interface{}{}, limit, offset, total)
		return
	}

	// Get the user service to include user details
	userService := c.MustGet("userService").(UserService)

	// Collect all user IDs
	userIDs := make([]primitive.ObjectID, 0, len(comments))
	for _, comment := range comments {
		userIDs = append(userIDs, comment.UserID)
	}

	// Get user details
	users, err := userService.GetUsersByIDs(c.Request.Context(), userIDs)
	if err != nil {
		// If user details retrieval fails, still return the comments
		c.Error(err)
		response.SuccessWithPagination(c, http.StatusOK, "Comments retrieved successfully", comments, limit, offset, total)
		return
	}

	// Create a map of user ID to user for quick lookup
	userMap := make(map[string]*models.User)
	for _, user := range users {
		userMap[user.ID.Hex()] = user
	}

	// Combine comments with user details
	commentResponses := make([]map[string]interface{}, 0, len(comments))
	for _, comment := range comments {
		user, exists := userMap[comment.UserID.Hex()]

		commentResponse := map[string]interface{}{
			"id":         comment.ID.Hex(),
			"stream_id":  comment.StreamID.Hex(),
			"user_id":    comment.UserID.Hex(),
			"content":    comment.Content,
			"created_at": comment.CreatedAt,
			"is_pinned":  comment.IsPinned,
		}

		if exists {
			commentResponse["username"] = user.Username
			commentResponse["display_name"] = user.DisplayName
			commentResponse["profile_image"] = user.ProfileImage
		}

		commentResponses = append(commentResponses, commentResponse)
	}

	response.SuccessWithPagination(c, http.StatusOK, "Comments retrieved successfully", commentResponses, limit, offset, total)
}

// DeleteComment handles deleting a comment from a live stream
func DeleteComment(c *gin.Context) {
	// Get comment ID from URL parameter
	commentIDStr := c.Param("comment_id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get the comment service
	commentService := c.MustGet("commentService").(CommentService)

	// Check if the user is the comment owner, stream host, or a moderator
	isOwner, err := commentService.IsCommentOwner(c.Request.Context(), commentID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check comment ownership", err)
		return
	}

	if !isOwner {
		// Get the stream ID from the comment
		comment, err := commentService.GetCommentByID(c.Request.Context(), commentID)
		if err != nil {
			response.Error(c, http.StatusNotFound, "Comment not found", err)
			return
		}

		streamID := comment.StreamID

		// Get the live stream service
		liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

		// Check if the user is the stream host or a moderator
		isHost, err := liveStreamService.IsStreamHost(c.Request.Context(), streamID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check if user is host", err)
			return
		}

		isModerator, err := liveStreamService.IsStreamModerator(c.Request.Context(), streamID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to check if user is moderator", err)
			return
		}

		if !isHost && !isModerator {
			response.Error(c, http.StatusForbidden, "You don't have permission to delete this comment", nil)
			return
		}
	}

	// Delete the comment
	err = commentService.DeleteComment(c.Request.Context(), commentID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete comment", err)
		return
	}

	response.Success(c, http.StatusOK, "Comment deleted successfully", nil)
}

// PinComment handles pinning or unpinning a comment in a live stream
func PinComment(c *gin.Context) {
	// Get stream ID and comment ID from URL parameters
	streamIDStr := c.Param("id")
	commentIDStr := c.Param("comment_id")

	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid comment ID format", err)
		return
	}

	// Get the authenticated user's ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse request body
	var req struct {
		IsPinned bool `json:"is_pinned"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the user is the stream host or a moderator
	isHost, err := liveStreamService.IsStreamHost(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is host", err)
		return
	}

	isModerator, err := liveStreamService.IsStreamModerator(c.Request.Context(), streamID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check if user is moderator", err)
		return
	}

	if !isHost && !isModerator {
		response.Error(c, http.StatusForbidden, "Only the stream host or moderators can pin comments", nil)
		return
	}

	// Get the comment service
	commentService := c.MustGet("commentService").(CommentService)

	// Pin or unpin the comment
	err = commentService.PinComment(c.Request.Context(), commentID, streamID, userID.(primitive.ObjectID), req.IsPinned)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update comment pin status", err)
		return
	}

	if req.IsPinned {
		response.Success(c, http.StatusOK, "Comment pinned successfully", nil)
	} else {
		response.Success(c, http.StatusOK, "Comment unpinned successfully", nil)
	}
}

// GetPinnedComments returns pinned comments for a live stream
func GetPinnedComments(c *gin.Context) {
	// Get stream ID from URL parameter
	streamIDStr := c.Param("id")
	streamID, err := primitive.ObjectIDFromHex(streamIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid stream ID format", err)
		return
	}

	// Get the live stream service
	liveStreamService := c.MustGet("liveStreamService").(LiveStreamService)

	// Check if the stream exists
	_, err = liveStreamService.GetStreamByID(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Stream not found", err)
		return
	}

	// Get the comment service
	commentService := c.MustGet("commentService").(CommentService)

	// Get pinned comments
	comments, err := commentService.GetPinnedComments(c.Request.Context(), streamID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve pinned comments", err)
		return
	}

	// If there are no pinned comments, return an empty array
	if len(comments) == 0 {
		response.Success(c, http.StatusOK, "No pinned comments found", []interface{}{})
		return
	}

	// Get the user service to include user details
	userService := c.MustGet("userService").(UserService)

	// Collect all user IDs
	userIDs := make([]primitive.ObjectID, 0, len(comments))
	for _, comment := range comments {
		userIDs = append(userIDs, comment.UserID)
	}

	// Get user details
	users, err := userService.GetUsersByIDs(c.Request.Context(), userIDs)
	if err != nil {
		// If user details retrieval fails, still return the comments
		c.Error(err)
		response.Success(c, http.StatusOK, "Pinned comments retrieved successfully", comments)
		return
	}

	// Create a map of user ID to user for quick lookup
	userMap := make(map[string]*models.User)
	for _, user := range users {
		userMap[user.ID.Hex()] = user
	}

	// Combine comments with user details
	commentResponses := make([]map[string]interface{}, 0, len(comments))
	for _, comment := range comments {
		user, exists := userMap[comment.UserID.Hex()]

		commentResponse := map[string]interface{}{
			"id":         comment.ID.Hex(),
			"stream_id":  comment.StreamID.Hex(),
			"user_id":    comment.UserID.Hex(),
			"content":    comment.Content,
			"created_at": comment.CreatedAt,
			"is_pinned":  comment.IsPinned,
		}

		if exists {
			commentResponse["username"] = user.Username
			commentResponse["display_name"] = user.DisplayName
			commentResponse["profile_image"] = user.ProfileImage
		}

		commentResponses = append(commentResponses, commentResponse)
	}

	response.Success(c, http.StatusOK, "Pinned comments retrieved successfully", commentResponses)
}
