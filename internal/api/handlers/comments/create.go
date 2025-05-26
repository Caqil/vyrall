package comments

import (
	"net/http"
	"time"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentService defines the interface for comment creation
type CommentService interface {
	CreateComment(comment *models.Comment) (primitive.ObjectID, error)
	GetPostByID(postID primitive.ObjectID) (*models.Post, error)
	GetCommentByID(commentID primitive.ObjectID) (*models.Comment, error)
	IncrementPostCommentCount(postID primitive.ObjectID) error
	IncrementCommentReplyCount(commentID primitive.ObjectID) error
	NotifyMentionedUsers(comment *models.Comment) error
}

// CreateCommentRequest represents the request to create a comment
type CreateCommentRequest struct {
	PostID         string   `json:"post_id" binding:"required"`
	Content        string   `json:"content" binding:"required"`
	ParentID       string   `json:"parent_id,omitempty"`
	MentionedUsers []string `json:"mentioned_users,omitempty"`
	MediaFiles     []string `json:"media_files,omitempty"`
}

// CreateComment handles the creation of a new comment
func CreateComment(c *gin.Context) {
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from authenticated user
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	// Parse post ID
	postID, err := primitive.ObjectIDFromHex(req.PostID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid post ID", err)
		return
	}

	commentService := c.MustGet("commentService").(CommentService)

	// Check if post exists
	post, err := commentService.GetPostByID(postID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Post not found", err)
		return
	}

	// Check if comments are allowed on this post
	if !post.AllowComments {
		response.Error(c, http.StatusForbidden, "Comments are not allowed on this post", nil)
		return
	}

	// Process mentioned users
	var mentionedUsers []primitive.ObjectID
	for _, userIDStr := range req.MentionedUsers {
		mentionedUserID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			continue // Skip invalid IDs
		}
		mentionedUsers = append(mentionedUsers, mentionedUserID)
	}

	// Process media files
	var mediaFiles []models.Media
	for _, mediaIDStr := range req.MediaFiles {
		mediaID, err := primitive.ObjectIDFromHex(mediaIDStr)
		if err != nil {
			continue // Skip invalid IDs
		}
		mediaFiles = append(mediaFiles, models.Media{ID: mediaID})
	}

	// Check if this is a reply to another comment
	var parentID *primitive.ObjectID
	if req.ParentID != "" {
		parsedParentID, err := primitive.ObjectIDFromHex(req.ParentID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid parent comment ID", err)
			return
		}

		// Check if parent comment exists
		parentComment, err := commentService.GetCommentByID(parsedParentID)
		if err != nil {
			response.Error(c, http.StatusNotFound, "Parent comment not found", err)
			return
		}

		// Make sure parent comment belongs to the same post
		if parentComment.PostID != postID {
			response.Error(c, http.StatusBadRequest, "Parent comment does not belong to the specified post", nil)
			return
		}

		parentID = &parsedParentID
	}

	// Create comment model
	comment := &models.Comment{
		PostID:         postID,
		UserID:         userID.(primitive.ObjectID),
		Content:        req.Content,
		MediaFiles:     mediaFiles,
		MentionedUsers: mentionedUsers,
		ParentID:       parentID,
		LikeCount:      0,
		ReplyCount:     0,
		IsEdited:       false,
		IsPinned:       false,
		IsHidden:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Create the comment
	commentID, err := commentService.CreateComment(comment)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create comment", err)
		return
	}

	// Increment comment counts
	if parentID != nil {
		// This is a reply, increment the parent comment's reply count
		if err := commentService.IncrementCommentReplyCount(*parentID); err != nil {
			// Log the error but don't fail the request
			// TODO: Implement proper logging
			// logger.Error("Failed to increment comment reply count", err)
		}
	} else {
		// This is a top-level comment, increment the post's comment count
		if err := commentService.IncrementPostCommentCount(postID); err != nil {
			// Log the error but don't fail the request
			// TODO: Implement proper logging
			// logger.Error("Failed to increment post comment count", err)
		}
	}

	// Notify mentioned users
	if len(mentionedUsers) > 0 {
		go commentService.NotifyMentionedUsers(comment)
	}

	response.Success(c, http.StatusCreated, "Comment created successfully", gin.H{
		"comment_id": commentID.Hex(),
	})
}
