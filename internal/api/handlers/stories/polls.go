package stories

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/story"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PollsHandler handles story polls operations
type PollsHandler struct {
	storyService *story.Service
}

// NewPollsHandler creates a new polls handler
func NewPollsHandler(storyService *story.Service) *PollsHandler {
	return &PollsHandler{
		storyService: storyService,
	}
}

// VoteOnPoll handles the request to vote on a story poll
func (h *PollsHandler) VoteOnPoll(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get story ID from URL parameter
	storyIDStr := c.Param("id")
	if !validation.IsValidObjectID(storyIDStr) {
		response.ValidationError(c, "Invalid story ID", nil)
		return
	}
	storyID, _ := primitive.ObjectIDFromHex(storyIDStr)

	// Parse request body
	var req struct {
		OptionIDs []string `json:"option_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if len(req.OptionIDs) == 0 {
		response.ValidationError(c, "At least one option ID is required", nil)
		return
	}

	// Convert option IDs to ObjectIDs
	optionIDs := make([]primitive.ObjectID, 0, len(req.OptionIDs))
	for _, idStr := range req.OptionIDs {
		if !validation.IsValidObjectID(idStr) {
			response.ValidationError(c, "Invalid option ID: "+idStr, nil)
			return
		}
		optionID, _ := primitive.ObjectIDFromHex(idStr)
		optionIDs = append(optionIDs, optionID)
	}

	// Vote on poll
	updatedPoll, err := h.storyService.VoteOnPoll(c.Request.Context(), storyID, userID.(primitive.ObjectID), optionIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to vote on poll", err)
		return
	}

	// Return success response
	response.OK(c, "Vote submitted successfully", updatedPoll)
}

// GetPollResults handles the request to get poll results
func (h *PollsHandler) GetPollResults(c *gin.Context) {
	// Get story ID from URL parameter
	storyIDStr := c.Param("id")
	if !validation.IsValidObjectID(storyIDStr) {
		response.ValidationError(c, "Invalid story ID", nil)
		return
	}
	storyID, _ := primitive.ObjectIDFromHex(storyIDStr)

	// Get authenticated user ID (may be nil for unauthenticated users)
	var userID primitive.ObjectID
	if id, exists := c.Get("userID"); exists {
		userID = id.(primitive.ObjectID)
	}

	// Get poll results
	results, err := h.storyService.GetPollResults(c.Request.Context(), storyID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get poll results", err)
		return
	}

	// Return success response
	response.OK(c, "Poll results retrieved successfully", results)
}

// GetMyPollVotes handles the request to get user's poll votes
func (h *PollsHandler) GetMyPollVotes(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get user's poll votes
	votes, total, err := h.storyService.GetUserPollVotes(c.Request.Context(), userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve poll votes", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Poll votes retrieved successfully", votes, limit, offset, total)
}

// RemovePollVote handles the request to remove a vote from a poll
func (h *PollsHandler) RemovePollVote(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get story ID from URL parameter
	storyIDStr := c.Param("id")
	if !validation.IsValidObjectID(storyIDStr) {
		response.ValidationError(c, "Invalid story ID", nil)
		return
	}
	storyID, _ := primitive.ObjectIDFromHex(storyIDStr)

	// Remove poll vote
	err := h.storyService.RemovePollVote(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to remove poll vote", err)
		return
	}

	// Return success response
	response.OK(c, "Poll vote removed successfully", nil)
}
