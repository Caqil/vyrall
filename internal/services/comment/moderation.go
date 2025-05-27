package comment

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Caqil/vyrall/internal/internal/models"
	"github.com/Caqil/vyrall/internal/pkg/errors"
	"github.com/Caqil/vyrall/internal/pkg/logging"
)

// ModerationService handles comment moderation
type ModerationService struct {
	commentRepo CommentRepository
	reportRepo  ReportRepository
	userRepo    UserRepository
	logger      logging.Logger
}

// NewModerationService creates a new moderation service
func NewModerationService(
	commentRepo CommentRepository,
	reportRepo ReportRepository,
	userRepo UserRepository,
	logger logging.Logger,
) *ModerationService {
	return &ModerationService{
		commentRepo: commentRepo,
		reportRepo:  reportRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// ReportComment creates a report for a comment
func (s *ModerationService) ReportComment(ctx context.Context, comment *models.Comment, reporterID primitive.ObjectID, reason string) error {
	// Validate reporter exists
	_, err := s.userRepo.FindByID(ctx, reporterID)
	if err != nil {
		return errors.Wrap(err, "Failed to find reporter")
	}

	// Check if user already reported this comment
	existingReport, err := s.reportRepo.FindByReporterAndContent(ctx, reporterID, comment.ID, "comment")
	if err == nil && existingReport != nil {
		return errors.New(errors.CodeInvalidOperation, "You have already reported this comment")
	}

	// Create report
	report := &models.Report{
		ReporterID:  reporterID,
		ContentID:   comment.ID,
		ContentType: "comment",
		ReasonCode:  reason,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = s.reportRepo.Create(ctx, report)
	if err != nil {
		return errors.Wrap(err, "Failed to create report")
	}

	return nil
}

// ReviewReport reviews a comment report
func (s *ModerationService) ReviewReport(ctx context.Context, reportID, moderatorID primitive.ObjectID, action, notes string) error {
	// Validate moderator exists
	moderator, err := s.userRepo.FindByID(ctx, moderatorID)
	if err != nil {
		return errors.Wrap(err, "Failed to find moderator")
	}

	// Check if moderator has permission
	if moderator.Role != "admin" && moderator.Role != "moderator" {
		return errors.New(errors.CodeForbidden, "You don't have permission to review reports")
	}

	// Get report
	report, err := s.reportRepo.FindByID(ctx, reportID)
	if err != nil {
		return errors.Wrap(err, "Failed to find report")
	}

	// Check if report is already reviewed
	if report.Status != "pending" {
		return errors.New(errors.CodeInvalidOperation, "Report has already been reviewed")
	}

	// Update report
	now := time.Now()
	report.Status = "reviewed"
	report.ModeratorID = &moderatorID
	report.ModeratorNotes = notes
	report.UpdatedAt = now
	report.ResolvedAt = &now

	// Apply action
	switch action {
	case "hide":
		// Hide the comment
		err = s.HideComment(ctx, report.ContentID, moderatorID, notes)
		if err != nil {
			return errors.Wrap(err, "Failed to hide comment")
		}
		report.ActionTaken = "hidden"
	case "delete":
		// Delete the comment
		comment, err := s.commentRepo.FindByID(ctx, report.ContentID)
		if err != nil {
			return errors.Wrap(err, "Failed to find comment")
		}

		now := time.Now()
		comment.DeletedAt = &now
		err = s.commentRepo.Update(ctx, comment)
		if err != nil {
			return errors.Wrap(err, "Failed to delete comment")
		}
		report.ActionTaken = "deleted"
	case "dismiss":
		// No action needed
		report.ActionTaken = "dismissed"
	default:
		return errors.New(errors.CodeInvalidArgument, "Invalid action")
	}

	// Update report
	err = s.reportRepo.Update(ctx, report)
	if err != nil {
		return errors.Wrap(err, "Failed to update report")
	}

	return nil
}

// HideComment hides a comment
func (s *ModerationService) HideComment(ctx context.Context, commentID primitive.ObjectID, moderatorID primitive.ObjectID, reason string) error {
	// Get comment
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return errors.Wrap(err, "Failed to find comment")
	}

	// Check if already hidden
	if comment.IsHidden {
		return errors.New(errors.CodeInvalidOperation, "Comment is already hidden")
	}

	// Hide comment
	comment.IsHidden = true
	comment.UpdatedAt = time.Now()

	// Create edit record if not already initialized
	if comment.EditHistory == nil {
		comment.EditHistory = []models.EditRecord{}
	}

	// Add moderation record
	moderationRecord := models.EditRecord{
		Content:  "Comment hidden by moderator: " + reason,
		EditedAt: time.Now(),
		EditorID: moderatorID,
		Reason:   reason,
	}
	comment.EditHistory = append(comment.EditHistory, moderationRecord)

	// Update comment
	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return errors.Wrap(err, "Failed to hide comment")
	}

	return nil
}

// UnhideComment unhides a comment
func (s *ModerationService) UnhideComment(ctx context.Context, commentID primitive.ObjectID, moderatorID primitive.ObjectID) error {
	// Get comment
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return errors.Wrap(err, "Failed to find comment")
	}

	// Check if already visible
	if !comment.IsHidden {
		return errors.New(errors.CodeInvalidOperation, "Comment is already visible")
	}

	// Unhide comment
	comment.IsHidden = false
	comment.UpdatedAt = time.Now()

	// Create edit record if not already initialized
	if comment.EditHistory == nil {
		comment.EditHistory = []models.EditRecord{}
	}

	// Add moderation record
	moderationRecord := models.EditRecord{
		Content:  "Comment unhidden by moderator",
		EditedAt: time.Now(),
		EditorID: moderatorID,
	}
	comment.EditHistory = append(comment.EditHistory, moderationRecord)

	// Update comment
	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return errors.Wrap(err, "Failed to unhide comment")
	}

	return nil
}

// GetReportedComments retrieves reported comments with status
func (s *ModerationService) GetReportedComments(ctx context.Context, status string, page, limit int) ([]models.Report, int, error) {
	// Set up filter
	filter := map[string]interface{}{
		"content_type": "comment",
	}

	if status != "" {
		filter["status"] = status
	}

	// Get reports
	reports, total, err := s.reportRepo.FindWithFilter(ctx, filter, page, limit, "created_at", "desc")
	if err != nil {
		return nil, 0, errors.Wrap(err, "Failed to find reports")
	}

	return reports, total, nil
}

// BulkDeleteComments deletes multiple comments
func (s *ModerationService) BulkDeleteComments(ctx context.Context, postID primitive.ObjectID, moderatorID primitive.ObjectID, reason string) (int, error) {
	// Validate moderator exists and has permission
	moderator, err := s.userRepo.FindByID(ctx, moderatorID)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to find moderator")
	}

	if moderator.Role != "admin" && moderator.Role != "moderator" {
		return 0, errors.New(errors.CodeForbidden, "You don't have permission to perform bulk deletion")
	}

	// Get all comments for the post
	filter := map[string]interface{}{
		"post_id":    postID,
		"deleted_at": nil,
	}

	comments, _, err := s.commentRepo.FindWithFilter(ctx, filter, 1, 1000, "created_at", "desc")
	if err != nil {
		return 0, errors.Wrap(err, "Failed to find comments")
	}

	// Soft delete all comments
	now := time.Now()
	count := 0

	for _, comment := range comments {
		comment.DeletedAt = &now
		err = s.commentRepo.Update(ctx, &comment)
		if err != nil {
			s.logger.Warn("Failed to delete comment during bulk delete", "commentId", comment.ID.Hex(), "error", err)
			continue
		}
		count++
	}

	return count, nil
}

// PinComment pins a comment to the top
func (s *ModerationService) PinComment(ctx context.Context, commentID primitive.ObjectID, userID primitive.ObjectID) error {
	// Get comment
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return errors.Wrap(err, "Failed to find comment")
	}

	// Check if already pinned
	if comment.IsPinned {
		return errors.New(errors.CodeInvalidOperation, "Comment is already pinned")
	}

	// Check permission (only post owner or moderator can pin)
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to find user")
	}

	// Pin comment
	comment.IsPinned = true
	comment.UpdatedAt = time.Now()

	// Update comment
	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return errors.Wrap(err, "Failed to pin comment")
	}

	return nil
}

// UnpinComment unpins a comment
func (s *ModerationService) UnpinComment(ctx context.Context, commentID primitive.ObjectID, userID primitive.ObjectID) error {
	// Get comment
	comment, err := s.commentRepo.FindByID(ctx, commentID)
	if err != nil {
		return errors.Wrap(err, "Failed to find comment")
	}

	// Check if already unpinned
	if !comment.IsPinned {
		return errors.New(errors.CodeInvalidOperation, "Comment is not pinned")
	}

	// Check permission (only post owner or moderator can unpin)
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Failed to find user")
	}

	// Unpin comment
	comment.IsPinned = false
	comment.UpdatedAt = time.Now()

	// Update comment
	err = s.commentRepo.Update(ctx, comment)
	if err != nil {
		return errors.Wrap(err, "Failed to unpin comment")
	}

	return nil
}
