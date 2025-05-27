package users

import (
	"net/http"

	"github.com/Caqil/vyrall/internal/services/user"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SettingsHandler handles user settings operations
type SettingsHandler struct {
	userService *user.Service
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(userService *user.Service) *SettingsHandler {
	return &SettingsHandler{
		userService: userService,
	}
}

// GetSettings handles the request to get user settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Get user settings
	settings, err := h.userService.GetSettings(c.Request.Context(), userID.(primitive.ObjectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve settings", err)
		return
	}

	// Return success response
	response.OK(c, "Settings retrieved successfully", settings)
}

// UpdateSettings handles the request to update user settings
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		NotificationPreferences map[string]interface{} `json:"notification_preferences,omitempty"`
		PrivacySettings         map[string]interface{} `json:"privacy_settings,omitempty"`
		LanguagePreference      string                 `json:"language_preference,omitempty"`
		ThemePreference         string                 `json:"theme_preference,omitempty"`
		AutoPlayVideos          *bool                  `json:"auto_play_videos,omitempty"`
		ShowOnlineStatus        *bool                  `json:"show_online_status,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Create updates map
	updates := make(map[string]interface{})

	if req.NotificationPreferences != nil {
		for key, value := range req.NotificationPreferences {
			updates["notification_preferences."+key] = value
		}
	}

	if req.PrivacySettings != nil {
		for key, value := range req.PrivacySettings {
			updates["privacy_settings."+key] = value
		}
	}

	if req.LanguagePreference != "" {
		updates["language_preference"] = req.LanguagePreference
	}

	if req.ThemePreference != "" {
		updates["theme_preference"] = req.ThemePreference
	}

	if req.AutoPlayVideos != nil {
		updates["auto_play_videos"] = *req.AutoPlayVideos
	}

	if req.ShowOnlineStatus != nil {
		updates["show_online_status"] = *req.ShowOnlineStatus
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update settings
	updatedSettings, err := h.userService.UpdateSettings(c.Request.Context(), userID.(primitive.ObjectID), updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update settings", err)
		return
	}

	// Return success response
	response.OK(c, "Settings updated successfully", updatedSettings)
}

// UpdateNotificationSettings handles the request to update notification settings
func (h *SettingsHandler) UpdateNotificationSettings(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		Push            map[string]bool        `json:"push,omitempty"`
		Email           map[string]bool        `json:"email,omitempty"`
		InApp           map[string]bool        `json:"in_app,omitempty"`
		MessagePreviews *bool                  `json:"message_previews,omitempty"`
		DoNotDisturb    map[string]interface{} `json:"do_not_disturb,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Create updates map
	updates := make(map[string]interface{})

	if req.Push != nil {
		for key, value := range req.Push {
			updates["notification_preferences.push."+key] = value
		}
	}

	if req.Email != nil {
		for key, value := range req.Email {
			updates["notification_preferences.email."+key] = value
		}
	}

	if req.InApp != nil {
		for key, value := range req.InApp {
			updates["notification_preferences.in_app."+key] = value
		}
	}

	if req.MessagePreviews != nil {
		updates["notification_preferences.message_previews"] = *req.MessagePreviews
	}

	if req.DoNotDisturb != nil {
		for key, value := range req.DoNotDisturb {
			updates["notification_preferences.do_not_disturb."+key] = value
		}
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update notification settings
	updatedSettings, err := h.userService.UpdateSettings(c.Request.Context(), userID.(primitive.ObjectID), updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update notification settings", err)
		return
	}

	// Return success response
	response.OK(c, "Notification settings updated successfully", updatedSettings)
}

// UpdatePrivacySettings handles the request to update privacy settings
func (h *SettingsHandler) UpdatePrivacySettings(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		response.UnauthorizedError(c, "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		WhoCanSeeMyPosts        string `json:"who_can_see_my_posts,omitempty"`
		WhoCanSendMeMessages    string `json:"who_can_send_me_messages,omitempty"`
		WhoCanSeeMyFriends      string `json:"who_can_see_my_friends,omitempty"`
		WhoCanTagMe             string `json:"who_can_tag_me,omitempty"`
		WhoCanSeeMyStories      string `json:"who_can_see_my_stories,omitempty"`
		HideMyOnlineStatus      *bool  `json:"hide_my_online_status,omitempty"`
		HideMyLastSeen          *bool  `json:"hide_my_last_seen,omitempty"`
		HideMyProfileFromSearch *bool  `json:"hide_my_profile_from_search,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Create updates map
	updates := make(map[string]interface{})

	if req.WhoCanSeeMyPosts != "" {
		updates["privacy_settings.who_can_see_my_posts"] = req.WhoCanSeeMyPosts
	}

	if req.WhoCanSendMeMessages != "" {
		updates["privacy_settings.who_can_send_me_messages"] = req.WhoCanSendMeMessages
	}

	if req.WhoCanSeeMyFriends != "" {
		updates["privacy_settings.who_can_see_my_friends"] = req.WhoCanSeeMyFriends
	}

	if req.WhoCanTagMe != "" {
		updates["privacy_settings.who_can_tag_me"] = req.WhoCanTagMe
	}

	if req.WhoCanSeeMyStories != "" {
		updates["privacy_settings.who_can_see_my_stories"] = req.WhoCanSeeMyStories
	}

	if req.HideMyOnlineStatus != nil {
		updates["privacy_settings.hide_my_online_status"] = *req.HideMyOnlineStatus
	}

	if req.HideMyLastSeen != nil {
		updates["privacy_settings.hide_my_last_seen"] = *req.HideMyLastSeen
	}

	if req.HideMyProfileFromSearch != nil {
		updates["privacy_settings.hide_my_profile_from_search"] = *req.HideMyProfileFromSearch
	}

	if len(updates) == 0 {
		response.ValidationError(c, "No updates provided", nil)
		return
	}

	// Update privacy settings
	updatedSettings, err := h.userService.UpdateSettings(c.Request.Context(), userID.(primitive.ObjectID), updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update privacy settings", err)
		return
	}

	// Return success response
	response.OK(c, "Privacy settings updated successfully", updatedSettings)
}
