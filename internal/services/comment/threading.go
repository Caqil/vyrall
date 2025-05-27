package comment

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// ThreadingService handles comment threading functionality
type ThreadingService struct {
	commentRepo CommentRepository
	postRepo    PostRepository
	logger      logging.Logger
}

// NewThreadingService creates a new threading service
func NewThreadingService(
	commentRepo CommentRepository,
	postRepo PostRepository,
	logger logging.Logger,
) *ThreadingService {
	return &ThreadingService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		logger:      logger,
	}
}

// CreateReply creates a reply to a comment
func (s *ThreadingService) CreateReply(ctx context.Context, parentComment *models.Comment, reply *models.Comment) (*models.Comment, error) {
	// Validate parent comment
	if parentComment == nil {
		return nil, errors.New(errors.CodeInvalidArgument, "Parent comment cannot be nil")
	}

	// Check if parent comment is hidden
	if parentComment.IsHidden {
		return nil, errors.New(errors.CodeInvalidOperation, "Cannot reply to a hidden comment")
	}

	// Check if parent comment is deleted
	if parentComment.DeletedAt != nil {
		return nil, errors.New(errors.CodeInvalidOperation, "Cannot reply to a deleted comment")
	}

	// Set default values
	now := time.Now()
	reply.CreatedAt = now
	reply.UpdatedAt = now
	reply.LikeCount = 0
	reply.ReplyCount = 0
	reply.IsEdited = false
	reply.IsHidden = false
	reply.IsPinned = false

	// Set parent ID and post ID
	reply.ParentID = &parentComment.ID
	reply.PostID = parentComment.PostID

	// Create the reply
	createdReply, err := s.commentRepo.Create(ctx, reply)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create reply")
	}

	// Update parent comment's reply count
	err = s.commentRepo.IncrementReplyCount(ctx, parentComment.ID)
	if err != nil {
		s.logger.Warn("Failed to increment reply count", "parentId", parentComment.ID.Hex(), "error", err)
	}

	return createdReply, nil
}

// GetReplies retrieves replies to a comment with pagination
func (s *ThreadingService) GetReplies(ctx context.Context, parentID primitive.ObjectID, options *CommentListOptions) ([]models.Comment, int, error) {
	// Set default options if not provided
	if options == nil {
		options = &CommentListOptions{
			Page:          1,
			Limit:         10,
			SortBy:        "created_at",
			SortOrder:     "asc",
			IncludeHidden: false,
		}
	}

	// Validate options
	if options.Page < 1 {
		options.Page = 1
	}

	if options.Limit < 1 {
		options.Limit = 10
	} else if options.Limit > 100 {
		options.Limit = 100
	}

	// Set up filter
	filter := map[string]interface{}{
		"parent_id":  parentID,
		"deleted_at": nil,
	}

	if !options.IncludeHidden {
		filter["is_hidden"] = false
	}

	// Apply time filters if provided
	if options.Since != nil {
		filter["created_at"] = map[string]interface{}{"$gte": options.Since}
	}

	if options.Until != nil {
		if filter["created_at"] == nil {
			filter["created_at"] = map[string]interface{}{"$lte": options.Until}
		} else {
			filter["created_at"].(map[string]interface{})["$lte"] = options.Until
		}
	}

	// Apply additional filters
	if options.Filters != nil {
		for key, value := range options.Filters {
			filter[key] = value
		}
	}

	return s.commentRepo.FindWithFilter(ctx, filter, options.Page, options.Limit, options.SortBy, options.SortOrder)
}

// GetThreadedComments retrieves comments in a threaded structure
func (s *ThreadingService) GetThreadedComments(ctx context.Context, postID primitive.ObjectID, options *ThreadedCommentsOptions) ([]ThreadedComment, error) {
	// Set default options if not provided
	if options == nil {
		options = &ThreadedCommentsOptions{
			MaxDepth:      3,
			TopLimit:      20,
			ReplyLimit:    5,
			SortBy:        "created_at",
			SortOrder:     "desc",
			IncludeHidden: false,
		}
	}

	// Validate options
	if options.MaxDepth < 1 {
		options.MaxDepth = 3
	} else if options.MaxDepth > 10 {
		options.MaxDepth = 10
	}

	if options.TopLimit < 1 {
		options.TopLimit = 20
	} else if options.TopLimit > 100 {
		options.TopLimit = 100
	}

	if options.ReplyLimit < 1 {
		options.ReplyLimit = 5
	} else if options.ReplyLimit > 50 {
		options.ReplyLimit = 50
	}

	// Get top-level comments
	topFilter := map[string]interface{}{
		"post_id":    postID,
		"parent_id":  nil,
		"deleted_at": nil,
	}

	if !options.IncludeHidden {
		topFilter["is_hidden"] = false
	}

	// Apply time filters if provided
	if options.Since != nil {
		topFilter["created_at"] = map[string]interface{}{"$gte": options.Since}
	}

	if options.Until != nil {
		if topFilter["created_at"] == nil {
			topFilter["created_at"] = map[string]interface{}{"$lte": options.Until}
		} else {
			topFilter["created_at"].(map[string]interface{})["$lte"] = options.Until
		}
	}

	// Sort options
	sortBy := options.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := options.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Get pinned comments first
	pinnedFilter := map[string]interface{}{
		"post_id":    postID,
		"parent_id":  nil,
		"is_pinned":  true,
		"deleted_at": nil,
	}

	if !options.IncludeHidden {
		pinnedFilter["is_hidden"] = false
	}

	pinnedComments, _, err := s.commentRepo.FindWithFilter(ctx, pinnedFilter, 1, 10, sortBy, sortOrder)
	if err != nil {
		s.logger.Warn("Failed to get pinned comments", "error", err)
		pinnedComments = []models.Comment{}
	}

	// Get regular top-level comments
	topFilter["is_pinned"] = false
	topComments, total, err := s.commentRepo.FindWithFilter(ctx, topFilter, 1, options.TopLimit, sortBy, sortOrder)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get top-level comments")
	}

	// Combine pinned and regular comments
	allTopComments := append(pinnedComments, topComments...)

	// Build threaded comments
	threadedComments := make([]ThreadedComment, 0, len(allTopComments))

	for _, comment := range allTopComments {
		// Get child count
		childFilter := map[string]interface{}{
			"parent_id":  comment.ID,
			"deleted_at": nil,
		}

		if !options.IncludeHidden {
			childFilter["is_hidden"] = false
		}

		totalReplies, err := s.commentRepo.Count(ctx, childFilter)
		if err != nil {
			s.logger.Warn("Failed to count replies", "commentId", comment.ID.Hex(), "error", err)
			totalReplies = 0
		}

		// Get replies
		replies, err := s.getChildComments(ctx, comment.ID, options, 1)
		if err != nil {
			s.logger.Warn("Failed to get child comments", "commentId", comment.ID.Hex(), "error", err)
			replies = []ThreadedComment{}
		}

		threadedComment := ThreadedComment{
			Comment:      comment,
			Replies:      replies,
			HasMore:      len(replies) < totalReplies,
			TotalReplies: totalReplies,
		}

		threadedComments = append(threadedComments, threadedComment)
	}

	return threadedComments, nil
}

// Helper methods

// getChildComments recursively fetches child comments up to max depth
func (s *ThreadingService) getChildComments(ctx context.Context, parentID primitive.ObjectID, options *ThreadedCommentsOptions, currentDepth int) ([]ThreadedComment, error) {
	// Stop if we've reached max depth
	if currentDepth >= options.MaxDepth {
		return []ThreadedComment{}, nil
	}

	// Set up filter
	filter := map[string]interface{}{
		"parent_id":  parentID,
		"deleted_at": nil,
	}

	if !options.IncludeHidden {
		filter["is_hidden"] = false
	}

	// Get replies with pagination
	replies, _, err := s.commentRepo.FindWithFilter(ctx, filter, 1, options.ReplyLimit, options.SortBy, options.SortOrder)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get replies")
	}

	// Build threaded replies
	threadedReplies := make([]ThreadedComment, 0, len(replies))

	for _, reply := range replies {
		// Get child count
		childFilter := map[string]interface{}{
			"parent_id":  reply.ID,
			"deleted_at": nil,
		}

		if !options.IncludeHidden {
			childFilter["is_hidden"] = false
		}

		totalReplies, err := s.commentRepo.Count(ctx, childFilter)
		if err != nil {
			s.logger.Warn("Failed to count replies", "commentId", reply.ID.Hex(), "error", err)
			totalReplies = 0
		}

		// Recursively get child comments
		childReplies, err := s.getChildComments(ctx, reply.ID, options, currentDepth+1)
		if err != nil {
			s.logger.Warn("Failed to get child comments", "commentId", reply.ID.Hex(), "error", err)
			childReplies = []ThreadedComment{}
		}

		threadedReply := ThreadedComment{
			Comment:      reply,
			Replies:      childReplies,
			HasMore:      len(childReplies) < totalReplies,
			TotalReplies: totalReplies,
		}

		threadedReplies = append(threadedReplies, threadedReply)
	}

	return threadedReplies, nil
}

// GetRepliesFlatMap gets all replies to a comment in a flat map structure
func (s *ThreadingService) GetRepliesFlatMap(ctx context.Context, commentID primitive.ObjectID) (map[string][]models.Comment, error) {
	// Set up result map (parentId -> replies)
	repliesMap := make(map[string][]models.Comment)

	// Get all replies recursively
	filter := map[string]interface{}{
		"parent_id":  commentID,
		"deleted_at": nil,
		"is_hidden":  false,
	}

	replies, _, err := s.commentRepo.FindWithFilter(ctx, filter, 1, 1000, "created_at", "asc")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get replies")
	}

	// Add top-level replies to map
	repliesMap[commentID.Hex()] = replies

	// Recursively get nested replies
	for _, reply := range replies {
		childMap, err := s.GetRepliesFlatMap(ctx, reply.ID)
		if err != nil {
			s.logger.Warn("Failed to get child replies", "replyId", reply.ID.Hex(), "error", err)
			continue
		}

		// Merge child map into result map
		for k, v := range childMap {
			repliesMap[k] = v
		}
	}

	return repliesMap, nil
}
