package stories

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/story"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuestionsHandler handles story questions operations
type QuestionsHandler struct {
	storyService *story.Service
}

// NewQuestionsHandler creates a new questions handler
func NewQuestionsHandler(storyService *story.Service) *QuestionsHandler {
	return &QuestionsHandler{
		storyService: storyService,
	}
}

// AnswerQuestion handles the request to answer a story question
func (h *QuestionsHandler) AnswerQuestion(c *gin.Context) {
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
		Answer string `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if req.Answer == "" {
		response.ValidationError(c, "Answer cannot be empty", nil)
		return
	}

	// Answer the question
	err := h.storyService.AnswerQuestion(c.Request.Context(), storyID, userID.(primitive.ObjectID), req.Answer)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to answer question", err)
		return
	}

	// Return success response
	response.OK(c, "Question answered successfully", nil)
}

// GetQuestionAnswers handles the request to get answers to a story question
func (h *QuestionsHandler) GetQuestionAnswers(c *gin.Context) {
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

	// Get pagination parameters
	limit, offset := response.GetPaginationParams(c)

	// Get question answers
	answers, total, err := h.storyService.GetQuestionAnswers(c.Request.Context(), storyID, userID.(primitive.ObjectID), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve question answers", err)
		return
	}

	// Return paginated response
	response.SuccessWithPagination(c, http.StatusOK, "Question answers retrieved successfully", answers, limit, offset, total)
}

// DeleteQuestionAnswer handles the request to delete an answer to a story question
func (h *QuestionsHandler) DeleteQuestionAnswer(c *gin.Context) {
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

	// Delete question answer
	err := h.storyService.DeleteQuestionAnswer(c.Request.Context(), storyID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete question answer", err)
		return
	}

	// Return success response
	response.OK(c, "Question answer deleted successfully", nil)
}

// FeatureQuestionAnswer handles the request to feature an answer to a story question
func (h *QuestionsHandler) FeatureQuestionAnswer(c *gin.Context) {
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
		AnswerID string `json:"answer_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	if !validation.IsValidObjectID(req.AnswerID) {
		response.ValidationError(c, "Invalid answer ID", nil)
		return
	}
	answerID, _ := primitive.ObjectIDFromHex(req.AnswerID)

	// Feature question answer
	err := h.storyService.FeatureQuestionAnswer(c.Request.Context(), storyID, answerID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to feature question answer", err)
		return
	}

	// Return success response
	response.OK(c, "Question answer featured successfully", nil)
}
