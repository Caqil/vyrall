package comment

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/internal/notification"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// NotificationsService handles notifications for comment actions
type NotificationsService struct {
	notifier    notification.Service
	userRepo    UserRepository
	postRepo    PostRepository
	commentRepo CommentRepository
	logger      logging.Logger
}

// NewNotificationsService creates a new notifications service
func NewNotificationsService(
	notifier notification.Service,
	userRepo UserRepository,
	postRepo PostRepository,
	commentRepo CommentRepository,
	logger logging.Logger,
) *NotificationsService {
	return &NotificationsService{
		notifier:    notifier,
		userRepo:    userRepo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		logger:      logger,
	}
}

// NotifyNewComment sends notifications for a new comment
func (s *NotificationsService) NotifyNewComment(ctx context.Context, comment *models.Comment, post *models.Post) {
	// Notify post owner if it's not their own comment
	if post.UserID != comment.UserID {
		s.notifyPostOwner(ctx, comment, post)
	}

	// Notify mentioned users
	s.notifyMentionedUsers(ctx, comment, post)
}

// NotifyNewReply sends notifications for a new reply
func (s *NotificationsService) NotifyNewReply(ctx context.Context, reply *models.Comment, parentComment *models.Comment) {
	// Get post info
	post, err := s.postRepo.FindByID(ctx, reply.PostID)
	if err != nil {
		s.logger.Warn("Failed to find post for reply notification", "postId", reply.PostID.Hex(), "error", err)
		return
	}

	// Notify parent comment author if it's not their own reply
	if parentComment.UserID != reply.UserID {
		s.notifyCommentAuthor(ctx, reply, parentComment, post)
	}

	// Notify mentioned users
	s.notifyMentionedUsers(ctx, reply, post)
}

// NotifyCommentLiked sends notification when a comment is liked
func (s *NotificationsService) NotifyCommentLiked(ctx context.Context, comment *models.Comment, likerID primitive.ObjectID) {
	// Don't notify if user liked their own comment
	if comment.UserID == likerID {
		return
	}

	// Get post info
	post, err := s.postRepo.FindByID(ctx, comment.PostID)
	if err != nil {
		s.logger.Warn("Failed to find post for like notification", "postId", comment.PostID.Hex(), "error", err)
		return
	}

	// Get liker info
	liker, err := s.userRepo.FindByID(ctx, likerID)
	if err != nil {
		s.logger.Warn("Failed to find liker for notification", "likerId", likerID.Hex(), "error", err)
		return
	}

	// Create notification
	notif := &models.Notification{
		UserID:         comment.UserID,
		Type:           "comment_like",
		Actor:          likerID,
		Subject:        "comment",
		SubjectID:      comment.ID,
		Message:        liker.Username + " liked your comment",
		SubjectPreview: truncateText(comment.Content, 50),
		ActionURL:      "/posts/" + post.ID.Hex() + "#comment-" + comment.ID.Hex(),
		IsRead:         false,
	}

	// Send notification
	err = s.notifier.Send(ctx, notif)
	if err != nil {
		s.logger.Warn("Failed to send like notification", "error", err)
	}
}

// NotifyCommentReaction sends notification when a comment receives a reaction
func (s *NotificationsService) NotifyCommentReaction(ctx context.Context, comment *models.Comment, userID primitive.ObjectID, reactionType string) {
	// Don't notify if user reacted to their own comment
	if comment.UserID == userID {
		return
	}

	// Get post info
	post, err := s.postRepo.FindByID(ctx, comment.PostID)
	if err != nil {
		s.logger.Warn("Failed to find post for reaction notification", "postId", comment.PostID.Hex(), "error", err)
		return
	}

	// Get reactor info
	reactor, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to find reactor for notification", "userId", userID.Hex(), "error", err)
		return
	}

	// Get reaction display text
	reactionText := getReactionText(reactionType)

	// Create notification
	notif := &models.Notification{
		UserID:         comment.UserID,
		Type:           "comment_reaction",
		Actor:          userID,
		Subject:        "comment",
		SubjectID:      comment.ID,
		Message:        reactor.Username + " reacted with " + reactionText + " to your comment",
		SubjectPreview: truncateText(comment.Content, 50),
		ActionURL:      "/posts/" + post.ID.Hex() + "#comment-" + comment.ID.Hex(),
		IsRead:         false,
	}

	// Send notification
	err = s.notifier.Send(ctx, notif)
	if err != nil {
		s.logger.Warn("Failed to send reaction notification", "error", err)
	}
}

// NotifyCommentHidden sends notification when a comment is hidden by a moderator
func (s *NotificationsService) NotifyCommentHidden(ctx context.Context, comment *models.Comment, moderatorID primitive.ObjectID, reason string) {
	// Get moderator info
	moderator, err := s.userRepo.FindByID(ctx, moderatorID)
	if err != nil {
		s.logger.Warn("Failed to find moderator for notification", "moderatorId", moderatorID.Hex(), "error", err)
		return
	}

	// Create notification
	notif := &models.Notification{
		UserID:         comment.UserID,
		Type:           "comment_moderated",
		Actor:          moderatorID,
		Subject:        "comment",
		SubjectID:      comment.ID,
		Message:        "Your comment has been hidden by a moderator: " + reason,
		SubjectPreview: truncateText(comment.Content, 50),
		ActionURL:      "/support/moderation",
		IsRead:         false,
		Priority:       "high",
	}

	// Send notification
	err = s.notifier.Send(ctx, notif)
	if err != nil {
		s.logger.Warn("Failed to send moderation notification", "error", err)
	}
}

// Helper methods

// notifyPostOwner notifies the post owner about a new comment
func (s *NotificationsService) notifyPostOwner(ctx context.Context, comment *models.Comment, post *models.Post) {
	// Get commenter info
	commenter, err := s.userRepo.FindByID(ctx, comment.UserID)
	if err != nil {
		s.logger.Warn("Failed to find commenter for notification", "userId", comment.UserID.Hex(), "error", err)
		return
	}

	// Create notification
	notif := &models.Notification{
		UserID:         post.UserID,
		Type:           "new_comment",
		Actor:          comment.UserID,
		Subject:        "post",
		SubjectID:      post.ID,
		Message:        commenter.Username + " commented on your post",
		SubjectPreview: truncateText(comment.Content, 50),
		ActionURL:      "/posts/" + post.ID.Hex() + "#comment-" + comment.ID.Hex(),
		IsRead:         false,
	}

	// Send notification
	err = s.notifier.Send(ctx, notif)
	if err != nil {
		s.logger.Warn("Failed to send new comment notification", "error", err)
	}
}

// notifyCommentAuthor notifies the parent comment author about a reply
func (s *NotificationsService) notifyCommentAuthor(ctx context.Context, reply *models.Comment, parentComment *models.Comment, post *models.Post) {
	// Get replier info
	replier, err := s.userRepo.FindByID(ctx, reply.UserID)
	if err != nil {
		s.logger.Warn("Failed to find replier for notification", "userId", reply.UserID.Hex(), "error", err)
		return
	}

	// Create notification
	notif := &models.Notification{
		UserID:         parentComment.UserID,
		Type:           "comment_reply",
		Actor:          reply.UserID,
		Subject:        "comment",
		SubjectID:      parentComment.ID,
		Message:        replier.Username + " replied to your comment",
		SubjectPreview: truncateText(reply.Content, 50),
		ActionURL:      "/posts/" + post.ID.Hex() + "#comment-" + reply.ID.Hex(),
		IsRead:         false,
	}

	// Send notification
	err = s.notifier.Send(ctx, notif)
	if err != nil {
		s.logger.Warn("Failed to send reply notification", "error", err)
	}
}

// notifyMentionedUsers notifies users mentioned in a comment
func (s *NotificationsService) notifyMentionedUsers(ctx context.Context, comment *models.Comment, post *models.Post) {
	if len(comment.MentionedUsers) == 0 {
		return
	}

	// Get commenter info
	commenter, err := s.userRepo.FindByID(ctx, comment.UserID)
	if err != nil {
		s.logger.Warn("Failed to find commenter for mention notification", "userId", comment.UserID.Hex(), "error", err)
		return
	}

	// Notify each mentioned user
	for _, mentionedUserID := range comment.MentionedUsers {
		// Skip if mentioned user is the commenter
		if mentionedUserID == comment.UserID {
			continue
		}

		// Create notification
		notif := &models.Notification{
			UserID:         mentionedUserID,
			Type:           "mention",
			Actor:          comment.UserID,
			Subject:        "comment",
			SubjectID:      comment.ID,
			Message:        commenter.Username + " mentioned you in a comment",
			SubjectPreview: truncateText(comment.Content, 50),
			ActionURL:      "/posts/" + post.ID.Hex() + "#comment-" + comment.ID.Hex(),
			IsRead:         false,
		}

		// Send notification
		err = s.notifier.Send(ctx, notif)
		if err != nil {
			s.logger.Warn("Failed to send mention notification", "error", err, "mentionedUserId", mentionedUserID.Hex())
		}
	}
}

// truncateText truncates text to a specified length
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

// getReactionText returns a human-readable text for a reaction type
func getReactionText(reactionType string) string {
	switch reactionType {
	case "like":
		return "👍"
	case "love":
		return "❤️"
	case "haha":
		return "😂"
	case "wow":
		return "😮"
	case "sad":
		return "😢"
	case "angry":
		return "😡"
	default:
		return reactionType
	}
}
