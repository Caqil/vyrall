package comment

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// CRUDService handles basic CRUD operations for comments
type CRUDService struct {
	commentRepo CommentRepository
	postRepo    PostRepository
	userRepo    UserRepository
	logger      logging.Logger
}

// NewCRUDService creates a new CRUD service for comments
func NewCRUDService(
	commentRepo CommentRepository,
	postRepo PostRepository,
	userRepo UserRepository,
	logger logging.Logger,
) *CRUDService {
	return &CRUDService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// CreateComment creates a new comment
func (s *CRUDService) CreateComment(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	// Set default values
	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now
	comment.LikeCount = 0
	comment.ReplyCount = 0
	comment.IsEdited = false
	comment.IsHidden = false

	// Create the comment
	createdComment, err := s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create comment")
	}

	// Update comment count on post
	err = s.postRepo.IncrementCommentCount(ctx, comment.PostID)
	if err != nil {
		s.logger.Warn("Failed to increment comment count", "postId", comment.PostID.Hex(), "error", err)
	}

	return createdComment, nil
}

// GetCommentByID retrieves a comment by ID
func (s *CRUDService) GetCommentByID(ctx context.Context, id primitive.ObjectID) (*models.Comment, error) {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			return nil, errors.New(errors.CodeNotFound, "Comment not found")
		}
		return nil, errors.Wrap(err, "Failed to find comment")
	}

	// Check if comment is deleted
	if comment.DeletedAt != nil {
		return nil, errors.New(errors.CodeNotFound, "Comment has been deleted")
	}

	return comment, nil
}

// GetCommentsByPostID retrieves comments for a post with pagination
func (s *CRUDService) GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, page, limit int) ([]models.Comment, int, error) {
	// Only get top-level comments (no parent ID)
	filter := map[string]interface{}{
		"post_id":    postID,
		"parent_id":  nil,
		"deleted_at": nil,
		"is_hidden":  false,
	}

	return s.commentRepo.FindWithFilter(ctx, filter, page, limit, "created_at", "desc")
}

// GetCommentsByUserID retrieves comments created by a user
func (s *CRUDService) GetCommentsByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]models.Comment, int, error) {
	filter := map[string]interface{}{
		"user_id":    userID,
		"deleted_at": nil,
		"is_hidden":  false,
	}

	return s.commentRepo.FindWithFilter(ctx, filter, page, limit, "created_at", "desc")
}

// UpdateComment updates a comment's content
func (s *CRUDService) UpdateComment(ctx context.Context, id primitive.ObjectID, content string) (*models.Comment, error) {
	// Get existing comment
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find comment")
	}

	// Update fields
	comment.Content = content
	comment.UpdatedAt = time.Now()
	comment.IsEdited = true

	// Create edit history record if not already there
	if comment.EditHistory == nil {
		comment.EditHistory = []models.EditRecord{}
	}

	// Add edit record
	editRecord := models.EditRecord{
		Content:  content,
		EditedAt: time.Now(),
		EditorID: comment.UserID, // Assuming the owner is editing
	}
	comment.EditHistory = append(comment.EditHistory, editRecord)

	// Update in database
	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update comment")
	}

	return comment, nil
}

// DeleteComment soft deletes a comment
func (s *CRUDService) DeleteComment(ctx context.Context, id primitive.ObjectID) error {
	// Get comment
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "Failed to find comment")
	}

	// Set deletion time
	now := time.Now()
	comment.DeletedAt = &now

	// Update comment
	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return errors.Wrap(err, "Failed to delete comment")
	}

	return nil
}

// HardDeleteComment permanently deletes a comment
func (s *CRUDService) HardDeleteComment(ctx context.Context, id primitive.ObjectID) error {
	// Get comment for post ID
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "Failed to find comment")
	}

	// Delete comment
	err = s.commentRepo.Delete(ctx, id)
	if err != nil {
		return errors.Wrap(err, "Failed to delete comment")
	}

	// Update comment count on post
	err = s.postRepo.DecrementCommentCount(ctx, comment.PostID)
	if err != nil {
		s.logger.Warn("Failed to decrement comment count", "postId", comment.PostID.Hex(), "error", err)
	}

	return nil
}

// GetCommentsByStatus retrieves comments with a specific status
func (s *CRUDService) GetCommentsByStatus(ctx context.Context, isHidden bool, page, limit int) ([]models.Comment, int, error) {
	filter := map[string]interface{}{
		"is_hidden":  isHidden,
		"deleted_at": nil,
	}

	return s.commentRepo.FindWithFilter(ctx, filter, page, limit, "created_at", "desc")
}

// GetTotalCommentCount gets the total number of comments for a post
func (s *CRUDService) GetTotalCommentCount(ctx context.Context, postID primitive.ObjectID) (int, error) {
	filter := map[string]interface{}{
		"post_id":    postID,
		"deleted_at": nil,
	}

	return s.commentRepo.Count(ctx, filter)
}
