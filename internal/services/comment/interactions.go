package comment

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// InteractionsService handles interactions with comments (likes, reactions)
type InteractionsService struct {
	commentRepo CommentRepository
	likeRepo    LikeRepository
	userRepo    UserRepository
	logger      logging.Logger
}

// NewInteractionsService creates a new interactions service
func NewInteractionsService(
	commentRepo CommentRepository,
	likeRepo LikeRepository,
	userRepo UserRepository,
	logger logging.Logger,
) *InteractionsService {
	return &InteractionsService{
		commentRepo: commentRepo,
		likeRepo:    likeRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// LikeComment adds a like to a comment
func (s *InteractionsService) LikeComment(ctx context.Context, comment *models.Comment, userID primitive.ObjectID) error {
	// Check if already liked
	existingLike, err := s.likeRepo.FindByUserAndContent(ctx, userID, comment.ID, "comment")
	if err == nil && existingLike != nil {
		// Already liked
		return errors.New(errors.CodeInvalidOperation, "Comment already liked")
	}

	// Create like
	like := &models.Like{
		UserID:       userID,
		ContentID:    comment.ID,
		ContentType:  "comment",
		ReactionType: "like",
		CreatedAt:    time.Now(),
	}

	_, err = s.likeRepo.Create(ctx, like)
	if err != nil {
		return errors.Wrap(err, "Failed to create like")
	}

	// Increment like count
	err = s.commentRepo.IncrementLikeCount(ctx, comment.ID)
	if err != nil {
		s.logger.Warn("Failed to increment like count", "commentId", comment.ID.Hex(), "error", err)
	}

	return nil
}

// UnlikeComment removes a like from a comment
func (s *InteractionsService) UnlikeComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	// Find existing like
	like, err := s.likeRepo.FindByUserAndContent(ctx, userID, commentID, "comment")
	if err != nil || like == nil {
		return errors.New(errors.CodeNotFound, "Like not found")
	}

	// Delete like
	err = s.likeRepo.Delete(ctx, like.ID)
	if err != nil {
		return errors.Wrap(err, "Failed to delete like")
	}

	// Decrement like count
	err = s.commentRepo.DecrementLikeCount(ctx, commentID)
	if err != nil {
		s.logger.Warn("Failed to decrement like count", "commentId", commentID.Hex(), "error", err)
	}

	return nil
}

// ReactToComment adds a reaction to a comment
func (s *InteractionsService) ReactToComment(ctx context.Context, comment *models.Comment, userID primitive.ObjectID, reactionType string) error {
	// Check for existing reaction
	existingReaction, err := s.likeRepo.FindByUserAndContent(ctx, userID, comment.ID, "comment")

	// If reaction exists and is the same type, return error
	if err == nil && existingReaction != nil && existingReaction.ReactionType == reactionType {
		return errors.New(errors.CodeInvalidOperation, "Reaction already exists")
	}

	// If reaction exists but is a different type, update it
	if err == nil && existingReaction != nil {
		// Get old reaction type for updating counts
		oldReactionType := existingReaction.ReactionType

		// Update reaction type
		existingReaction.ReactionType = reactionType
		err = s.likeRepo.Update(ctx, existingReaction)
		if err != nil {
			return errors.Wrap(err, "Failed to update reaction")
		}

		// Update reaction counts
		s.updateReactionCounts(ctx, comment, oldReactionType, reactionType)
		return nil
	}

	// Create new reaction
	reaction := &models.Like{
		UserID:       userID,
		ContentID:    comment.ID,
		ContentType:  "comment",
		ReactionType: reactionType,
		CreatedAt:    time.Now(),
	}

	_, err = s.likeRepo.Create(ctx, reaction)
	if err != nil {
		return errors.Wrap(err, "Failed to create reaction")
	}

	// Initialize reaction counts map if nil
	if comment.ReactionCounts == nil {
		comment.ReactionCounts = make(map[string]int)
	}

	// Increment reaction count
	comment.ReactionCounts[reactionType]++

	// Update comment with new counts
	err = s.commentRepo.UpdateReactionCounts(ctx, comment.ID, comment.ReactionCounts)
	if err != nil {
		s.logger.Warn("Failed to update reaction counts", "commentId", comment.ID.Hex(), "error", err)
	}

	return nil
}

// RemoveReaction removes a reaction from a comment
func (s *InteractionsService) RemoveReaction(ctx context.Context, commentID, userID primitive.ObjectID) error {
	// Find existing reaction
	reaction, err := s.likeRepo.FindByUserAndContent(ctx, userID, commentID, "comment")
	if err != nil || reaction == nil {
		return errors.New(errors.CodeNotFound, "Reaction not found")
	}

	// Get reaction type for updating counts
	reactionType := reaction.ReactionType

	// Delete reaction
	err = s.likeRepo.Delete(ctx, reaction.ID)
	if err != nil {
		return errors.Wrap(err, "Failed to delete reaction")
	}

	// Get comment to update reaction counts
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		s.logger.Warn("Failed to find comment for reaction count update", "commentId", commentID.Hex(), "error", err)
		return nil
	}

	// Update reaction counts
	if comment.ReactionCounts != nil && comment.ReactionCounts[reactionType] > 0 {
		comment.ReactionCounts[reactionType]--

		// Update comment with new counts
		err = s.commentRepo.UpdateReactionCounts(ctx, comment.ID, comment.ReactionCounts)
		if err != nil {
			s.logger.Warn("Failed to update reaction counts", "commentId", comment.ID.Hex(), "error", err)
		}
	}

	return nil
}

// GetCommentLikes retrieves users who liked a comment
func (s *InteractionsService) GetCommentLikes(ctx context.Context, commentID primitive.ObjectID) ([]primitive.ObjectID, error) {
	// Get likes for comment
	likes, err := s.likeRepo.FindByContent(ctx, commentID, "comment", "like")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find likes")
	}

	// Extract user IDs
	userIDs := make([]primitive.ObjectID, 0, len(likes))
	for _, like := range likes {
		userIDs = append(userIDs, like.UserID)
	}

	return userIDs, nil
}

// GetCommentReactions retrieves reactions to a comment
func (s *InteractionsService) GetCommentReactions(ctx context.Context, commentID primitive.ObjectID) (map[string][]primitive.ObjectID, error) {
	// Get all reactions for comment
	reactions, err := s.likeRepo.FindAllByContent(ctx, commentID, "comment")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find reactions")
	}

	// Group by reaction type
	result := make(map[string][]primitive.ObjectID)
	for _, reaction := range reactions {
		if result[reaction.ReactionType] == nil {
			result[reaction.ReactionType] = make([]primitive.ObjectID, 0)
		}
		result[reaction.ReactionType] = append(result[reaction.ReactionType], reaction.UserID)
	}

	return result, nil
}

// IsLikedByUser checks if a comment is liked by a user
func (s *InteractionsService) IsLikedByUser(ctx context.Context, commentID, userID primitive.ObjectID) (bool, error) {
	like, err := s.likeRepo.FindByUserAndContent(ctx, userID, commentID, "comment")
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			return false, nil
		}
		return false, errors.Wrap(err, "Failed to check like status")
	}

	return like != nil, nil
}

// GetUserReaction gets a user's reaction to a comment
func (s *InteractionsService) GetUserReaction(ctx context.Context, commentID, userID primitive.ObjectID) (string, error) {
	reaction, err := s.likeRepo.FindByUserAndContent(ctx, userID, commentID, "comment")
	if err != nil {
		if errors.Code(err) == errors.CodeNotFound {
			return "", nil
		}
		return "", errors.Wrap(err, "Failed to get user reaction")
	}

	if reaction == nil {
		return "", nil
	}

	return reaction.ReactionType, nil
}

// Helper methods

// updateReactionCounts updates reaction counts when changing reaction types
func (s *InteractionsService) updateReactionCounts(ctx context.Context, comment *models.Comment, oldType, newType string) {
	// Initialize reaction counts map if nil
	if comment.ReactionCounts == nil {
		comment.ReactionCounts = make(map[string]int)
	}

	// Decrement old type count
	if comment.ReactionCounts[oldType] > 0 {
		comment.ReactionCounts[oldType]--
	}

	// Increment new type count
	comment.ReactionCounts[newType]++

	// Update comment with new counts
	err := s.commentRepo.UpdateReactionCounts(ctx, comment.ID, comment.ReactionCounts)
	if err != nil {
		s.logger.Warn("Failed to update reaction counts", "commentId", comment.ID.Hex(), "error", err)
	}
}
