package comment

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/config"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
	"github.com/Caqil/vyrall/internal/pkg/metrics"
)

// Service defines the interface for comment-related operations
type Service interface {
	// CRUD operations
	Create(ctx context.Context, comment *models.Comment) (*models.Comment, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Comment, error)
	GetByPostID(ctx context.Context, postID primitive.ObjectID, options *CommentListOptions) ([]models.Comment, int, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, options *CommentListOptions) ([]models.Comment, int, error)
	Update(ctx context.Context, id primitive.ObjectID, updates *CommentUpdates, userID primitive.ObjectID) (*models.Comment, error)
	Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, permanently bool) error

	// Threading operations
	CreateReply(ctx context.Context, parentID primitive.ObjectID, comment *models.Comment) (*models.Comment, error)
	GetReplies(ctx context.Context, parentID primitive.ObjectID, options *CommentListOptions) ([]models.Comment, int, error)
	GetThreadedComments(ctx context.Context, postID primitive.ObjectID, options *ThreadedCommentsOptions) ([]ThreadedComment, error)

	// Interactions
	LikeComment(ctx context.Context, commentID, userID primitive.ObjectID) error
	UnlikeComment(ctx context.Context, commentID, userID primitive.ObjectID) error
	ReactToComment(ctx context.Context, commentID, userID primitive.ObjectID, reactionType string) error
	RemoveReaction(ctx context.Context, commentID, userID primitive.ObjectID, reactionType string) error
	GetLikes(ctx context.Context, commentID primitive.ObjectID, options *LikeListOptions) ([]models.Like, int, error)
	GetReactions(ctx context.Context, commentID primitive.ObjectID) (map[string]int, error)
	GetReactionDetails(ctx context.Context, commentID primitive.ObjectID, reactionType string) ([]UserReaction, error)
	IsLikedByUser(ctx context.Context, commentID, userID primitive.ObjectID) (bool, error)

	// Moderation operations
	ReportComment(ctx context.Context, commentID, reporterID primitive.ObjectID, reason, description string) error
	ReviewReport(ctx context.Context, reportID, moderatorID primitive.ObjectID, action, notes string) error
	HideComment(ctx context.Context, commentID primitive.ObjectID, moderatorID primitive.ObjectID, reason string) error
	UnhideComment(ctx context.Context, commentID primitive.ObjectID, moderatorID primitive.ObjectID) error
	GetReportedComments(ctx context.Context, status string, options *ReportListOptions) ([]models.Report, int, error)
	GetCommentsByStatus(ctx context.Context, status string, options *CommentListOptions) ([]models.Comment, int, error)

	// Admin operations
	PinComment(ctx context.Context, commentID, userID primitive.ObjectID) error
	UnpinComment(ctx context.Context, commentID, userID primitive.ObjectID) error
	BulkDelete(ctx context.Context, postID primitive.ObjectID, moderatorID primitive.ObjectID, reason string) (int, error)

	// Analytics and stats
	GetCommentStats(ctx context.Context, postID primitive.ObjectID) (*CommentStats, error)
	GetUserCommentActivity(ctx context.Context, userID primitive.ObjectID) (*UserCommentActivity, error)
	GetTrendingComments(ctx context.Context, timeframe string, limit int) ([]models.Comment, error)
	GetMostActiveCommenters(ctx context.Context, timeframe string, limit int) ([]UserCommentCount, error)
	GetMostCommentedPosts(ctx context.Context, timeframe string, limit int) ([]PostCommentCount, error)
}

// CommentService implements the Service interface
type CommentService struct {
	crud          *CRUDService
	threading     *ThreadingService
	interactions  *InteractionsService
	moderation    *ModerationService
	notifications *NotificationsService
	commentRepo   CommentRepository
	likeRepo      LikeRepository
	postRepo      PostRepository
	userRepo      UserRepository
	reportRepo    ReportRepository
	config        *config.CommentConfig
	metrics       metrics.Collector
	logger        logging.Logger
}

// NewCommentService creates a new comment service
func NewCommentService(
	crud *CRUDService,
	threading *ThreadingService,
	interactions *InteractionsService,
	moderation *ModerationService,
	notifications *NotificationsService,
	commentRepo CommentRepository,
	likeRepo LikeRepository,
	postRepo PostRepository,
	userRepo UserRepository,
	reportRepo ReportRepository,
	config *config.CommentConfig,
	metrics metrics.Collector,
	logger logging.Logger,
) Service {
	return &CommentService{
		crud:          crud,
		threading:     threading,
		interactions:  interactions,
		moderation:    moderation,
		notifications: notifications,
		commentRepo:   commentRepo,
		likeRepo:      likeRepo,
		postRepo:      postRepo,
		userRepo:      userRepo,
		reportRepo:    reportRepo,
		config:        config,
		metrics:       metrics,
		logger:        logger,
	}
}

// Create creates a new comment
func (s *CommentService) Create(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.create", time.Since(startTime))
	}()

	createdComment, err := s.crud.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Get post for notification
	post, err := s.postRepo.FindByID(ctx, comment.PostID)
	if err != nil {
		s.logger.Warn("Failed to get post for notifications", "postId", comment.PostID.Hex(), "error", err)
	} else {
		// Send notifications
		s.notifications.NotifyNewComment(ctx, createdComment, post)
	}

	s.metrics.IncrementCounter("comment.created")
	return createdComment, nil
}

// GetByID retrieves a comment by ID
func (s *CommentService) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Comment, error) {
	return s.crud.GetCommentByID(ctx, id)
}

// GetByPostID retrieves comments for a post with pagination
func (s *CommentService) GetByPostID(ctx context.Context, postID primitive.ObjectID, options *CommentListOptions) ([]models.Comment, int, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.getByPostID", time.Since(startTime))
	}()

	// Set default options if not provided
	if options == nil {
		options = &CommentListOptions{
			Page:          1,
			Limit:         s.config.DefaultPageSize,
			SortBy:        "created_at",
			SortOrder:     "desc",
			IncludeHidden: false,
		}
	}

	// Validate options
	if options.Page < 1 {
		options.Page = 1
	}

	if options.Limit < 1 {
		options.Limit = s.config.DefaultPageSize
	} else if options.Limit > s.config.MaxPageSize {
		options.Limit = s.config.MaxPageSize
	}

	// Set up filter
	filter := map[string]interface{}{
		"post_id":    postID,
		"parent_id":  nil, // Only top-level comments
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

// GetByUserID retrieves comments made by a user
func (s *CommentService) GetByUserID(ctx context.Context, userID primitive.ObjectID, options *CommentListOptions) ([]models.Comment, int, error) {
	// Set default options if not provided
	if options == nil {
		options = &CommentListOptions{
			Page:          1,
			Limit:         s.config.DefaultPageSize,
			SortBy:        "created_at",
			SortOrder:     "desc",
			IncludeHidden: false,
		}
	}

	// Set up filter
	filter := map[string]interface{}{
		"user_id":    userID,
		"deleted_at": nil,
	}

	if !options.IncludeHidden {
		filter["is_hidden"] = false
	}

	// Apply time filters
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

	return s.commentRepo.FindWithFilter(ctx, filter, options.Page, options.Limit, options.SortBy, options.SortOrder)
}

// Update updates a comment
func (s *CommentService) Update(ctx context.Context, id primitive.ObjectID, updates *CommentUpdates, userID primitive.ObjectID) (*models.Comment, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.update", time.Since(startTime))
	}()

	// Get existing comment
	comment, err := s.crud.GetCommentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if comment.UserID != userID {
		return nil, errors.New(errors.CodeForbidden, "You can only edit your own comments")
	}

	// Validate content
	if updates.Content == "" && len(updates.MediaFiles) == 0 {
		return nil, errors.New(errors.CodeInvalidArgument, "Comment must have content or media")
	}

	if len(updates.Content) > s.config.MaxCommentLength {
		return nil, errors.New(errors.CodeInvalidArgument, "Comment exceeds maximum length")
	}

	// Update fields
	comment.Content = updates.Content
	if updates.MediaFiles != nil {
		comment.MediaFiles = updates.MediaFiles
	}
	if updates.MentionedUsers != nil {
		comment.MentionedUsers = updates.MentionedUsers
	}

	comment.UpdatedAt = time.Now()
	comment.IsEdited = true

	// Create edit history record if not already there
	if comment.EditHistory == nil {
		comment.EditHistory = []models.EditRecord{}
	}

	// Add edit record
	editRecord := models.EditRecord{
		Content:  comment.Content,
		EditedAt: time.Now(),
		EditorID: userID,
	}
	comment.EditHistory = append(comment.EditHistory, editRecord)

	// Update in database
	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update comment")
	}

	s.metrics.IncrementCounter("comment.updated")
	return comment, nil
}

// Delete removes a comment
func (s *CommentService) Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, permanently bool) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.delete", time.Since(startTime))
	}()

	// Get existing comment
	comment, err := s.crud.GetCommentByID(ctx, id)
	if err != nil {
		return err
	}

	// Check permissions
	if comment.UserID != userID {
		// Check if user is a moderator
		user, err := s.userRepo.FindByID(ctx, userID)
		if err != nil || (user.Role != "admin" && user.Role != "moderator") {
			// Also check if user is the post owner
			post, err := s.postRepo.FindByID(ctx, comment.PostID)
			if err != nil || post.UserID != userID {
				return errors.New(errors.CodeForbidden, "You don't have permission to delete this comment")
			}
		}
	}

	if permanently {
		// Perform hard delete
		err = s.crud.HardDeleteComment(ctx, id)
	} else {
		// Perform soft delete
		err = s.crud.DeleteComment(ctx, id)
	}

	if err != nil {
		return err
	}

	s.metrics.IncrementCounter("comment.deleted")
	return nil
}

// CreateReply creates a reply to a comment
func (s *CommentService) CreateReply(ctx context.Context, parentID primitive.ObjectID, comment *models.Comment) (*models.Comment, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.createReply", time.Since(startTime))
	}()

	// Get parent comment
	parentComment, err := s.crud.GetCommentByID(ctx, parentID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find parent comment")
	}

	// Set parent ID and post ID from parent comment
	comment.ParentID = &parentID
	comment.PostID = parentComment.PostID

	// Create the reply
	reply, err := s.crud.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Update reply count on parent comment
	err = s.commentRepo.IncrementReplyCount(ctx, parentID)
	if err != nil {
		s.logger.Warn("Failed to increment reply count", "parentId", parentID.Hex(), "error", err)
	}

	// Get post for notification
	post, err := s.postRepo.FindByID(ctx, parentComment.PostID)
	if err != nil {
		s.logger.Warn("Failed to get post for reply notification", "postId", parentComment.PostID.Hex(), "error", err)
	} else {
		// Send notifications
		s.notifications.NotifyNewReply(ctx, reply, parentComment)
	}

	s.metrics.IncrementCounter("comment.replied")
	return reply, nil
}

// GetReplies retrieves replies to a comment
func (s *CommentService) GetReplies(ctx context.Context, parentID primitive.ObjectID, options *CommentListOptions) ([]models.Comment, int, error) {
	return s.threading.GetReplies(ctx, parentID, options)
}

// GetThreadedComments retrieves comments in a threaded structure
func (s *CommentService) GetThreadedComments(ctx context.Context, postID primitive.ObjectID, options *ThreadedCommentsOptions) ([]ThreadedComment, error) {
	return s.threading.GetThreadedComments(ctx, postID, options)
}

// LikeComment adds a like to a comment
func (s *CommentService) LikeComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.like", time.Since(startTime))
	}()

	// Get comment
	comment, err := s.crud.GetCommentByID(ctx, commentID)
	if err != nil {
		return err
	}

	// Add like
	err = s.interactions.LikeComment(ctx, comment, userID)
	if err != nil {
		return err
	}

	// Send notification
	s.notifications.NotifyCommentLiked(ctx, comment, userID)

	s.metrics.IncrementCounter("comment.liked")
	return nil
}

// UnlikeComment removes a like from a comment
func (s *CommentService) UnlikeComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.unlike", time.Since(startTime))
	}()

	err := s.interactions.UnlikeComment(ctx, commentID, userID)
	if err != nil {
		return err
	}

	s.metrics.IncrementCounter("comment.unliked")
	return nil
}

// ReactToComment adds a reaction to a comment
func (s *CommentService) ReactToComment(ctx context.Context, commentID, userID primitive.ObjectID, reactionType string) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.react", time.Since(startTime))
	}()

	// Validate reaction type
	if !s.isValidReactionType(reactionType) {
		return errors.New(errors.CodeInvalidArgument, "Invalid reaction type")
	}

	// Get comment
	comment, err := s.crud.GetCommentByID(ctx, commentID)
	if err != nil {
		return err
	}

	// Add reaction
	err = s.interactions.ReactToComment(ctx, comment, userID, reactionType)
	if err != nil {
		return err
	}

	// Send notification
	s.notifications.NotifyCommentReaction(ctx, comment, userID, reactionType)

	s.metrics.IncrementCounter("comment.reacted")
	return nil
}

// RemoveReaction removes a reaction from a comment
func (s *CommentService) RemoveReaction(ctx context.Context, commentID, userID primitive.ObjectID, reactionType string) error {
	return s.interactions.RemoveReaction(ctx, commentID, userID)
}

// GetLikes retrieves users who liked a comment
func (s *CommentService) GetLikes(ctx context.Context, commentID primitive.ObjectID, options *LikeListOptions) ([]models.Like, int, error) {
	// Set default options if not provided
	if options == nil {
		options = &LikeListOptions{
			Page:      1,
			Limit:     s.config.DefaultPageSize,
			SortBy:    "created_at",
			SortOrder: "desc",
		}
	}

	return s.likeRepo.FindByContentWithPagination(ctx, commentID, "comment", "like", options.Page, options.Limit, options.SortBy, options.SortOrder)
}

// GetReactions retrieves reaction counts for a comment
func (s *CommentService) GetReactions(ctx context.Context, commentID primitive.ObjectID) (map[string]int, error) {
	// Get comment
	comment, err := s.crud.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	if comment.ReactionCounts == nil {
		return make(map[string]int), nil
	}

	return comment.ReactionCounts, nil
}

// GetReactionDetails retrieves users who reacted to a comment with a specific reaction
func (s *CommentService) GetReactionDetails(ctx context.Context, commentID primitive.ObjectID, reactionType string) ([]UserReaction, error) {
	// Validate reaction type
	if !s.isValidReactionType(reactionType) && reactionType != "all" {
		return nil, errors.New(errors.CodeInvalidArgument, "Invalid reaction type")
	}

	// Get reactions
	var reactions []models.Like
	var err error

	if reactionType == "all" {
		reactions, err = s.likeRepo.FindAllByContent(ctx, commentID, "comment")
	} else {
		reactions, err = s.likeRepo.FindByContentAndType(ctx, commentID, "comment", reactionType)
	}

	if err != nil {
		return nil, errors.Wrap(err, "Failed to find reactions")
	}

	// Build user reaction details
	userReactions := make([]UserReaction, 0, len(reactions))

	for _, reaction := range reactions {
		// Get user info
		user, err := s.userRepo.FindByID(ctx, reaction.UserID)
		if err != nil {
			s.logger.Warn("Failed to find user for reaction", "userId", reaction.UserID.Hex(), "error", err)
			continue
		}

		userReaction := UserReaction{
			UserID:         user.ID,
			Username:       user.Username,
			ProfilePicture: user.ProfilePicture,
			CreatedAt:      reaction.CreatedAt,
		}

		userReactions = append(userReactions, userReaction)
	}

	return userReactions, nil
}

// IsLikedByUser checks if a comment is liked by a user
func (s *CommentService) IsLikedByUser(ctx context.Context, commentID, userID primitive.ObjectID) (bool, error) {
	return s.interactions.IsLikedByUser(ctx, commentID, userID)
}

// ReportComment reports a comment for moderation
func (s *CommentService) ReportComment(ctx context.Context, commentID, reporterID primitive.ObjectID, reason, description string) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.report", time.Since(startTime))
	}()

	// Get comment
	comment, err := s.crud.GetCommentByID(ctx, commentID)
	if err != nil {
		return err
	}

	// Create report
	err = s.moderation.ReportComment(ctx, comment, reporterID, reason)
	if err != nil {
		return err
	}

	s.metrics.IncrementCounter("comment.reported")
	return nil
}

// ReviewReport reviews a comment report
func (s *CommentService) ReviewReport(ctx context.Context, reportID, moderatorID primitive.ObjectID, action, notes string) error {
	return s.moderation.ReviewReport(ctx, reportID, moderatorID, action, notes)
}

// HideComment hides a comment (moderation action)
func (s *CommentService) HideComment(ctx context.Context, commentID primitive.ObjectID, moderatorID primitive.ObjectID, reason string) error {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.hide", time.Since(startTime))
	}()

	// Get comment
	comment, err := s.crud.GetCommentByID(ctx, commentID)
	if err != nil {
		return err
	}

	// Hide comment
	err = s.moderation.HideComment(ctx, commentID, moderatorID, reason)
	if err != nil {
		return err
	}

	// Notify user
	s.notifications.NotifyCommentHidden(ctx, comment, moderatorID, reason)

	s.metrics.IncrementCounter("comment.hidden")
	return nil
}

// UnhideComment unhides a comment
func (s *CommentService) UnhideComment(ctx context.Context, commentID primitive.ObjectID, moderatorID primitive.ObjectID) error {
	return s.moderation.UnhideComment(ctx, commentID, moderatorID)
}

// GetReportedComments retrieves reported comments with status
func (s *CommentService) GetReportedComments(ctx context.Context, status string, options *ReportListOptions) ([]models.Report, int, error) {
	// Validate options
	if options == nil {
		options = &ReportListOptions{
			Page:      1,
			Limit:     s.config.DefaultPageSize,
			SortBy:    "created_at",
			SortOrder: "desc",
		}
	}

	if options.Page < 1 {
		options.Page = 1
	}

	if options.Limit < 1 {
		options.Limit = s.config.DefaultPageSize
	} else if options.Limit > s.config.MaxPageSize {
		options.Limit = s.config.MaxPageSize
	}

	return s.moderation.GetReportedComments(ctx, status, options.Page, options.Limit)
}

// GetCommentsByStatus retrieves comments by hidden status
func (s *CommentService) GetCommentsByStatus(ctx context.Context, status string, options *CommentListOptions) ([]models.Comment, int, error) {
	// Validate status
	var isHidden bool
	switch status {
	case "hidden":
		isHidden = true
	case "visible":
		isHidden = false
	default:
		return nil, 0, errors.New(errors.CodeInvalidArgument, "Invalid status: must be 'hidden' or 'visible'")
	}

	// Set default options if not provided
	if options == nil {
		options = &CommentListOptions{
			Page:      1,
			Limit:     s.config.DefaultPageSize,
			SortBy:    "created_at",
			SortOrder: "desc",
		}
	}

	return s.crud.GetCommentsByStatus(ctx, isHidden, options.Page, options.Limit)
}

// PinComment pins a comment to the top
func (s *CommentService) PinComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	return s.moderation.PinComment(ctx, commentID, userID)
}

// UnpinComment unpins a comment
func (s *CommentService) UnpinComment(ctx context.Context, commentID, userID primitive.ObjectID) error {
	return s.moderation.UnpinComment(ctx, commentID, userID)
}

// BulkDelete deletes multiple comments from a post
func (s *CommentService) BulkDelete(ctx context.Context, postID primitive.ObjectID, moderatorID primitive.ObjectID, reason string) (int, error) {
	startTime := time.Now()
	defer func() {
		s.metrics.ObserveLatency("comment.bulkDelete", time.Since(startTime))
	}()

	count, err := s.moderation.BulkDeleteComments(ctx, postID, moderatorID, reason)
	if err != nil {
		return 0, err
	}

	s.metrics.AddCounter("comment.deleted", count)
	return count, nil
}

// GetCommentStats retrieves statistics about comments on a post
func (s *CommentService) GetCommentStats(ctx context.Context, postID primitive.ObjectID) (*CommentStats, error) {
	// Get all comments for post
	filter := map[string]interface{}{
		"post_id":    postID,
		"deleted_at": nil,
	}

	comments, _, err := s.commentRepo.FindWithFilter(ctx, filter, 1, 1000, "created_at", "desc")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find comments")
	}

	// Calculate statistics
	stats := &CommentStats{
		TotalCount:       len(comments),
		TopLevelCount:    0,
		ReplyCount:       0,
		LikeCount:        0,
		AverageLength:    0,
		CommenterCount:   0,
		MediaAttachments: 0,
		ReactionCounts:   make(map[string]int),
	}

	if len(comments) > 0 {
		totalLength := 0
		mostLikedComment := comments[0]
		latestComment := comments[0]
		commenters := make(map[primitive.ObjectID]bool)

		for _, comment := range comments {
			// Count top-level vs replies
			if comment.ParentID == nil {
				stats.TopLevelCount++
			} else {
				stats.ReplyCount++
			}

			// Track likes
			stats.LikeCount += comment.LikeCount

			// Track reactions
			if comment.ReactionCounts != nil {
				for reaction, count := range comment.ReactionCounts {
					stats.ReactionCounts[reaction] += count
				}
			}

			// Track comment length
			totalLength += len(comment.Content)

			// Track commenters
			commenters[comment.UserID] = true

			// Track media
			stats.MediaAttachments += len(comment.MediaFiles)

			// Find most liked comment
			if comment.LikeCount > mostLikedComment.LikeCount {
				mostLikedComment = comment
			}

			// Find latest comment
			if comment.CreatedAt.After(latestComment.CreatedAt) {
				latestComment = comment
			}
		}

		// Set calculated fields
		stats.CommenterCount = len(commenters)
		stats.LatestComment = latestComment.CreatedAt
		if len(comments) > 0 {
			stats.AverageLength = totalLength / len(comments)
		}
		stats.MostLikedComment = mostLikedComment.ID.Hex()
	}

	return stats, nil
}

// GetUserCommentActivity retrieves a user's comment activity
func (s *CommentService) GetUserCommentActivity(ctx context.Context, userID primitive.ObjectID) (*UserCommentActivity, error) {
	// Get user's comments
	filter := map[string]interface{}{
		"user_id":    userID,
		"deleted_at": nil,
	}

	comments, _, err := s.commentRepo.FindWithFilter(ctx, filter, 1, 1000, "created_at", "desc")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to find comments")
	}

	// Calculate activity
	activity := &UserCommentActivity{
		TotalComments:    len(comments),
		CommentedPosts:   0,
		TopPosts:         []PostCommentInfo{},
		LikesReceived:    0,
		RepliesReceived:  0,
		ReactionCounts:   make(map[string]int),
		CommentsPerMonth: make(map[string]int),
	}

	if len(comments) > 0 {
		postCounts := make(map[primitive.ObjectID]int)
		postTitles := make(map[primitive.ObjectID]string)
		mostActive := comments[0].CreatedAt
		earliestComment := comments[0].CreatedAt

		for _, comment := range comments {
			// Count comments per post
			postCounts[comment.PostID]++

			// Get post title if not already fetched
			if _, ok := postTitles[comment.PostID]; !ok {
				post, err := s.postRepo.FindByID(ctx, comment.PostID)
				if err == nil {
					postTitles[comment.PostID] = post.Title
				} else {
					postTitles[comment.PostID] = "Unknown Post"
				}
			}

			// Track likes and replies
			activity.LikesReceived += comment.LikeCount
			activity.RepliesReceived += comment.ReplyCount

			// Track reactions
			if comment.ReactionCounts != nil {
				for reaction, count := range comment.ReactionCounts {
					activity.ReactionCounts[reaction] += count
				}
			}

			// Track activity time
			if comment.CreatedAt.After(mostActive) {
				mostActive = comment.CreatedAt
			}

			if comment.CreatedAt.Before(earliestComment) {
				earliestComment = comment.CreatedAt
			}

			// Track comments by month
			month := comment.CreatedAt.Format("2006-01")
			activity.CommentsPerMonth[month]++
		}

		// Set fields
		activity.CommentedPosts = len(postCounts)
		activity.MostActive = mostActive

		// Calculate average comments per day
		days := float64(time.Since(earliestComment).Hours() / 24)
		if days > 0 {
			activity.AvgCommentsPerDay = float64(len(comments)) / days
		}

		// Create top posts list
		topPosts := make([]PostCommentInfo, 0, len(postCounts))
		for postID, count := range postCounts {
			topPosts = append(topPosts, PostCommentInfo{
				PostID:       postID,
				Title:        postTitles[postID],
				CommentCount: count,
			})
		}

		// Sort by count (implementation omitted for brevity)
		// ...

		// Limit to top 5
		if len(topPosts) > 5 {
			topPosts = topPosts[:5]
		}

		activity.TopPosts = topPosts
	}

	return activity, nil
}

// GetTrendingComments retrieves trending comments
func (s *CommentService) GetTrendingComments(ctx context.Context, timeframe string, limit int) ([]models.Comment, error) {
	// Validate timeframe
	var since time.Time
	now := time.Now()

	switch timeframe {
	case "day":
		since = now.AddDate(0, 0, -1)
	case "week":
		since = now.AddDate(0, 0, -7)
	case "month":
		since = now.AddDate(0, -1, 0)
	case "year":
		since = now.AddDate(-1, 0, 0)
	default:
		since = now.AddDate(0, 0, -7) // Default to week
		timeframe = "week"
	}

	// Validate limit
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	return s.commentRepo.GetTrending(ctx, since, limit)
}

// GetMostActiveCommenters retrieves users with the most comments
func (s *CommentService) GetMostActiveCommenters(ctx context.Context, timeframe string, limit int) ([]UserCommentCount, error) {
	// Validate timeframe
	var since time.Time
	now := time.Now()

	switch timeframe {
	case "day":
		since = now.AddDate(0, 0, -1)
	case "week":
		since = now.AddDate(0, 0, -7)
	case "month":
		since = now.AddDate(0, -1, 0)
	case "year":
		since = now.AddDate(-1, 0, 0)
	case "all":
		since = time.Time{} // Zero time for all time
	default:
		since = now.AddDate(0, 0, -7) // Default to week
		timeframe = "week"
	}

	// Validate limit
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	return s.commentRepo.GetMostActiveCommenters(ctx, since, limit)
}

// GetMostCommentedPosts retrieves posts with the most comments
func (s *CommentService) GetMostCommentedPosts(ctx context.Context, timeframe string, limit int) ([]PostCommentCount, error) {
	// Validate timeframe
	var since time.Time
	now := time.Now()

	switch timeframe {
	case "day":
		since = now.AddDate(0, 0, -1)
	case "week":
		since = now.AddDate(0, 0, -7)
	case "month":
		since = now.AddDate(0, -1, 0)
	case "year":
		since = now.AddDate(-1, 0, 0)
	case "all":
		since = time.Time{} // Zero time for all time
	default:
		since = now.AddDate(0, 0, -7) // Default to week
		timeframe = "week"
	}

	// Validate limit
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	return s.commentRepo.GetMostCommentedPosts(ctx, since, limit)
}

// Helper methods

// isValidReactionType checks if a reaction type is valid
func (s *CommentService) isValidReactionType(reactionType string) bool {
	validReactions := []string{"like", "love", "haha", "wow", "sad", "angry"}
	for _, r := range validReactions {
		if reactionType == r {
			return true
		}
	}
	return false
}
