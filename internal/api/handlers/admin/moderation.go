package admin

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ModerationService interface for moderation operations
type ModerationService interface {
	GetModerationQueue(contentType, status string, limit, offset int) ([]map[string]interface{}, int, error)
	GetModerationItem(contentType string, id primitive.ObjectID) (map[string]interface{}, error)
	ModerationAction(contentType string, id primitive.ObjectID, action, reason string) error
	GetModerationHistory(contentType string, id primitive.ObjectID) ([]map[string]interface{}, error)
	GetModerationStats() (map[string]interface{}, error)
	GetModerationSettings() (map[string]interface{}, error)
	UpdateModerationSettings(settings map[string]interface{}) error
	GetAutoModerationRules() ([]map[string]interface{}, error)
	CreateAutoModerationRule(rule map[string]interface{}) (primitive.ObjectID, error)
	UpdateAutoModerationRule(id primitive.ObjectID, rule map[string]interface{}) error
	DeleteAutoModerationRule(id primitive.ObjectID) error
	GetModerationLog(limit, offset int) ([]map[string]interface{}, int, error)
}

// GetModerationQueue returns items pending moderation
func GetModerationQueue(c *gin.Context) {
	contentType := c.DefaultQuery("type", "all")
	status := c.DefaultQuery("status", "pending")
	limit, offset := getPaginationParams(c)

	moderationService := c.MustGet("moderationService").(ModerationService)

	// Validate status
	validStatuses := map[string]bool{
		"pending":  true,
		"approved": true,
		"rejected": true,
		"flagged":  true,
		"all":      true,
	}

	if !validStatuses[status] {
		response.Error(c, http.StatusBadRequest, "Invalid status", nil)
		return
	}

	queue, total, err := moderationService.GetModerationQueue(contentType, status, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve moderation queue", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Moderation queue retrieved successfully", queue, limit, offset, total)
}

// GetModerationItem returns details about a specific item pending moderation
func GetModerationItem(c *gin.Context) {
	contentType := c.Param("type")
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	moderationService := c.MustGet("moderationService").(ModerationService)

	item, err := moderationService.GetModerationItem(contentType, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve moderation item", err)
		return
	}

	response.Success(c, http.StatusOK, "Moderation item retrieved successfully", item)
}

// ModerationAction performs an action on a moderation item
func ModerationAction(c *gin.Context) {
	contentType := c.Param("type")
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	moderationService := c.MustGet("moderationService").(ModerationService)

	// Validate action
	validActions := map[string]bool{
		"approve":          true,
		"reject":           true,
		"hide":             true,
		"delete":           true,
		"flag":             true,
		"unflag":           true,
		"escalate":         true,
		"warn_user":        true,
		"suspend_user":     true,
		"restrict_content": true,
	}

	if !validActions[req.Action] {
		response.Error(c, http.StatusBadRequest, "Invalid action", nil)
		return
	}

	if err := moderationService.ModerationAction(contentType, id, req.Action, req.Reason); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to perform moderation action", err)
		return
	}

	response.Success(c, http.StatusOK, "Moderation action performed successfully", nil)
}

// GetModerationHistory returns the history of moderation actions for an item
func GetModerationHistory(c *gin.Context) {
	contentType := c.Param("type")
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid item ID", err)
		return
	}

	moderationService := c.MustGet("moderationService").(ModerationService)

	history, err := moderationService.GetModerationHistory(contentType, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve moderation history", err)
		return
	}

	response.Success(c, http.StatusOK, "Moderation history retrieved successfully", history)
}

// GetModerationStats returns statistics about moderation activities
func GetModerationStats(c *gin.Context) {
	moderationService := c.MustGet("moderationService").(ModerationService)

	stats, err := moderationService.GetModerationStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve moderation statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Moderation statistics retrieved successfully", stats)
}

// GetModerationSettings returns the current moderation settings
func GetModerationSettings(c *gin.Context) {
	moderationService := c.MustGet("moderationService").(ModerationService)

	settings, err := moderationService.GetModerationSettings()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve moderation settings", err)
		return
	}

	response.Success(c, http.StatusOK, "Moderation settings retrieved successfully", settings)
}

// UpdateModerationSettings updates the moderation settings
func UpdateModerationSettings(c *gin.Context) {
	var settings map[string]interface{}

	if err := c.ShouldBindJSON(&settings); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	moderationService := c.MustGet("moderationService").(ModerationService)

	if err := moderationService.UpdateModerationSettings(settings); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update moderation settings", err)
		return
	}

	response.Success(c, http.StatusOK, "Moderation settings updated successfully", nil)
}

// GetAutoModerationRules returns the auto-moderation rules
func GetAutoModerationRules(c *gin.Context) {
	moderationService := c.MustGet("moderationService").(ModerationService)

	rules, err := moderationService.GetAutoModerationRules()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve auto-moderation rules", err)
		return
	}

	response.Success(c, http.StatusOK, "Auto-moderation rules retrieved successfully", rules)
}

// CreateAutoModerationRule creates a new auto-moderation rule
func CreateAutoModerationRule(c *gin.Context) {
	var rule map[string]interface{}

	if err := c.ShouldBindJSON(&rule); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	moderationService := c.MustGet("moderationService").(ModerationService)

	id, err := moderationService.CreateAutoModerationRule(rule)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create auto-moderation rule", err)
		return
	}

	response.Success(c, http.StatusCreated, "Auto-moderation rule created successfully", gin.H{
		"rule_id": id.Hex(),
	})
}

// UpdateAutoModerationRule updates an auto-moderation rule
func UpdateAutoModerationRule(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid rule ID", err)
		return
	}

	var rule map[string]interface{}

	if err := c.ShouldBindJSON(&rule); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	moderationService := c.MustGet("moderationService").(ModerationService)

	if err := moderationService.UpdateAutoModerationRule(id, rule); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update auto-moderation rule", err)
		return
	}

	response.Success(c, http.StatusOK, "Auto-moderation rule updated successfully", nil)
}

// DeleteAutoModerationRule deletes an auto-moderation rule
func DeleteAutoModerationRule(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid rule ID", err)
		return
	}

	moderationService := c.MustGet("moderationService").(ModerationService)

	if err := moderationService.DeleteAutoModerationRule(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete auto-moderation rule", err)
		return
	}

	response.Success(c, http.StatusOK, "Auto-moderation rule deleted successfully", nil)
}

// GetModerationLog returns the moderation activity log
func GetModerationLog(c *gin.Context) {
	limit, offset := getPaginationParams(c)

	moderationService := c.MustGet("moderationService").(ModerationService)

	log, total, err := moderationService.GetModerationLog(limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve moderation log", err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "Moderation log retrieved successfully", log, limit, offset, total)
}
