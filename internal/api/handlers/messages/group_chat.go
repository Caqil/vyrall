package messages

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/models"
	"github.com/Caqil/vyrall/internal/services/message"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/Caqil/vyrall/internal/utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupChatHandler handles group chat operations
type GroupChatHandler struct {
	messageService *message.Service
}

// NewGroupChatHandler creates a new group chat handler
func NewGroupChatHandler(messageService *message.Service) *GroupChatHandler {
	return &GroupChatHandler{
		messageService: messageService,
	}
}

// CreateGroupChat handles the request to create a new group chat
func (h *GroupChatHandler) CreateGroupChat(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Title          string   `json:"title" binding:"required"`
		ParticipantIDs []string `json:"participant_ids" binding:"required"`
		Avatar         string   `json:"avatar,omitempty"`
		Description    string   `json:"description,omitempty"`
		IsEncrypted    bool     `json:"is_encrypted,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate participants
	if len(req.ParticipantIDs) == 0 {
		response.ValidationError(c, "At least one participant is required", nil)
		return
	}

	// Convert participant IDs to ObjectIDs
	participantIDs := make([]primitive.ObjectID, 0, len(req.ParticipantIDs)+1) // +1 for the creator
	participantIDs = append(participantIDs, userID.(primitive.ObjectID))       // Add creator

	for _, idStr := range req.ParticipantIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		partID, _ := primitive.ObjectIDFromHex(idStr)
		// Avoid duplicates
		if partID != userID.(primitive.ObjectID) {
			participantIDs = append(participantIDs, partID)
		}
	}

	if len(participantIDs) < 2 {
		response.ValidationError(c, "At least one valid participant is required", nil)
		return
	}

	// Create group info
	groupInfo := &models.GroupChatInfo{
		Description: req.Description,
		CreatorID:   userID.(primitive.ObjectID),
		Admins:      []primitive.ObjectID{userID.(primitive.ObjectID)},
		IsPublic:    false,
		MaxMembers:  100, // Default max members
	}

	// Create the group chat conversation
	conversation, err := h.messageService.CreateGroupChat(c.Request.Context(), req.Title, req.Avatar, participantIDs, req.IsEncrypted, groupInfo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create group chat", err)
		return
	}

	// Return success response
	response.Created(c, "Group chat created successfully", conversation)
}

// AddParticipants handles the request to add participants to a group chat
func (h *GroupChatHandler) AddParticipants(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	if !validation.IsValidObjectID(conversationIDStr) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	// Parse request body
	var req struct {
		ParticipantIDs []string `json:"participant_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate participants
	if len(req.ParticipantIDs) == 0 {
		response.ValidationError(c, "At least one participant is required", nil)
		return
	}

	// Check if user is an admin in the group
	isAdmin, err := h.messageService.IsGroupAdmin(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify admin status", err)
		return
	}

	if !isAdmin {
		response.ForbiddenError(c, "Only group admins can add participants")
		return
	}

	// Convert participant IDs to ObjectIDs
	participantIDs := make([]primitive.ObjectID, 0, len(req.ParticipantIDs))
	for _, idStr := range req.ParticipantIDs {
		if !validation.IsValidObjectID(idStr) {
			continue // Skip invalid IDs
		}
		partID, _ := primitive.ObjectIDFromHex(idStr)
		participantIDs = append(participantIDs, partID)
	}

	if len(participantIDs) == 0 {
		response.ValidationError(c, "No valid participant IDs provided", nil)
		return
	}

	// Add participants to the group chat
	addedParticipants, err := h.messageService.AddParticipantsToGroup(c.Request.Context(), conversationID, participantIDs, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to add participants", err)
		return
	}

	// Return success response
	response.OK(c, "Participants added successfully", gin.H{
		"added_participants": addedParticipants,
		"added_count":        len(addedParticipants),
	})
}

// RemoveParticipant handles the request to remove a participant from a group chat
func (h *GroupChatHandler) RemoveParticipant(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	if !validation.IsValidObjectID(conversationIDStr) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	// Get participant ID from URL parameter
	participantIDStr := c.Param("participantId")
	if !validation.IsValidObjectID(participantIDStr) {
		response.ValidationError(c, "Invalid participant ID", nil)
		return
	}
	participantID, _ := primitive.ObjectIDFromHex(participantIDStr)

	// Check if user is trying to remove themselves
	isSelf := participantID == userID.(primitive.ObjectID)

	// Check if user is an admin in the group (not needed if removing self)
	if !isSelf {
		isAdmin, err := h.messageService.IsGroupAdmin(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to verify admin status", err)
			return
		}

		if !isAdmin {
			response.ForbiddenError(c, "Only group admins can remove participants")
			return
		}
	}

	// Remove participant from the group chat
	err := h.messageService.RemoveParticipantFromGroup(c.Request.Context(), conversationID, participantID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to remove participant", err)
		return
	}

	// Return success response
	response.OK(c, "Participant removed successfully", nil)
}

// UpdateGroupInfo handles the request to update group chat information
func (h *GroupChatHandler) UpdateGroupInfo(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	if !validation.IsValidObjectID(conversationIDStr) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	// Parse request body
	var req struct {
		Title        string   `json:"title,omitempty"`
		Avatar       string   `json:"avatar,omitempty"`
		Description  string   `json:"description,omitempty"`
		IsPublic     *bool    `json:"is_public,omitempty"`
		JoinApproval *bool    `json:"join_approval,omitempty"`
		Rules        []string `json:"rules,omitempty"`
		MaxMembers   *int     `json:"max_members,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Check if user is an admin in the group
	isAdmin, err := h.messageService.IsGroupAdmin(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify admin status", err)
		return
	}

	if !isAdmin {
		response.ForbiddenError(c, "Only group admins can update group information")
		return
	}

	// Create update map
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Description != "" {
		updates["group_info.description"] = req.Description
	}
	if req.IsPublic != nil {
		updates["group_info.is_public"] = *req.IsPublic
	}
	if req.JoinApproval != nil {
		updates["group_info.join_approval"] = *req.JoinApproval
	}
	if len(req.Rules) > 0 {
		groupRules := make([]models.GroupRule, 0, len(req.Rules))
		for _, rule := range req.Rules {
			groupRules = append(groupRules, models.GroupRule{
				Title:       rule,
				Description: "",
			})
		}
		updates["group_info.rules"] = groupRules
	}
	if req.MaxMembers != nil && *req.MaxMembers > 0 {
		updates["group_info.max_members"] = *req.MaxMembers
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update group info
	updatedConversation, err := h.messageService.UpdateGroupInfo(c.Request.Context(), conversationID, updates, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update group information", err)
		return
	}

	// Return success response
	response.OK(c, "Group information updated successfully", updatedConversation)
}

// MakeAdmin handles the request to promote a participant to admin
func (h *GroupChatHandler) MakeAdmin(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	if !validation.IsValidObjectID(conversationIDStr) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	// Get participant ID from URL parameter
	participantIDStr := c.Param("participantId")
	if !validation.IsValidObjectID(participantIDStr) {
		response.ValidationError(c, "Invalid participant ID", nil)
		return
	}
	participantID, _ := primitive.ObjectIDFromHex(participantIDStr)

	// Check if user is an admin in the group
	isAdmin, err := h.messageService.IsGroupAdmin(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify admin status", err)
		return
	}

	if !isAdmin {
		response.ForbiddenError(c, "Only group admins can promote participants to admin")
		return
	}

	// Promote participant to admin
	err = h.messageService.MakeGroupAdmin(c.Request.Context(), conversationID, participantID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to promote participant to admin", err)
		return
	}

	// Return success response
	response.OK(c, "Participant promoted to admin successfully", nil)
}

// RemoveAdmin handles the request to demote an admin to regular participant
func (h *GroupChatHandler) RemoveAdmin(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	if !validation.IsValidObjectID(conversationIDStr) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	// Get admin ID from URL parameter
	adminIDStr := c.Param("adminId")
	if !validation.IsValidObjectID(adminIDStr) {
		response.ValidationError(c, "Invalid admin ID", nil)
		return
	}
	adminID, _ := primitive.ObjectIDFromHex(adminIDStr)

	// Check if trying to demote the creator
	isCreator, err := h.messageService.IsGroupCreator(c.Request.Context(), conversationID, adminID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify creator status", err)
		return
	}

	if isCreator {
		response.ValidationError(c, "Group creator cannot be demoted", nil)
		return
	}

	// Check if user is the creator or trying to demote themselves
	isSelfDemotion := adminID == userID.(primitive.ObjectID)
	if !isSelfDemotion {
		isCreator, err = h.messageService.IsGroupCreator(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to verify creator status", err)
			return
		}

		if !isCreator {
			response.ForbiddenError(c, "Only the group creator can demote admins")
			return
		}
	}

	// Demote admin to regular participant
	err = h.messageService.RemoveGroupAdmin(c.Request.Context(), conversationID, adminID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to demote admin", err)
		return
	}

	// Return success response
	response.OK(c, "Admin demoted successfully", nil)
}

// GenerateInviteLink handles the request to generate a group chat invite link
func (h *GroupChatHandler) GenerateInviteLink(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get conversation ID from URL parameter
	conversationIDStr := c.Param("id")
	if !validation.IsValidObjectID(conversationIDStr) {
		response.ValidationError(c, "Invalid conversation ID", nil)
		return
	}
	conversationID, _ := primitive.ObjectIDFromHex(conversationIDStr)

	// Check if user is an admin in the group
	isAdmin, err := h.messageService.IsGroupAdmin(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to verify admin status", err)
		return
	}

	if !isAdmin {
		response.ForbiddenError(c, "Only group admins can generate invite links")
		return
	}

	// Generate invite link
	inviteLink, err := h.messageService.GenerateGroupInviteLink(c.Request.Context(), conversationID, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate invite link", err)
		return
	}

	// Return success response
	response.OK(c, "Invite link generated successfully", gin.H{
		"invite_link": inviteLink,
	})
}

// JoinViaInviteLink handles the request to join a group chat via invite link
func (h *GroupChatHandler) JoinViaInviteLink(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		InviteLink string `json:"invite_link" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Join group via invite link
	conversation, err := h.messageService.JoinGroupViaInviteLink(c.Request.Context(), req.InviteLink, userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to join group chat", err)
		return
	}

	// Return success response
	response.OK(c, "Joined group chat successfully", conversation)
}
