package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/users"
	"github.com/gin-gonic/gin"
)

// SetupUserRoutes configures the user routes
func SetupUserRoutes(router *gin.Engine, userHandler *users.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Users routes group
	userGroup := router.Group("/api/users")

	// Public user endpoints
	userGroup.Use(optionalAuth)
	userGroup.GET("/:identifier", userHandler.GetProfile)
	userGroup.GET("/search", userHandler.SearchUsers)
	userGroup.GET("/:id/posts", userHandler.GetUserPosts)

	// Protected user endpoints (require authentication)
	protectedUserGroup := userGroup.Group("")
	protectedUserGroup.Use(authMiddleware)

	// Profile management
	protectedUserGroup.GET("/me", userHandler.GetMyProfile)
	protectedUserGroup.PUT("/me", userHandler.UpdateProfile)
	protectedUserGroup.POST("/me/username", userHandler.ChangeUsername)
	protectedUserGroup.POST("/:id/view", userHandler.ViewProfile)

	// User settings
	protectedUserGroup.GET("/settings", userHandler.GetSettings)
	protectedUserGroup.PUT("/settings", userHandler.UpdateSettings)
	protectedUserGroup.PUT("/settings/notifications", userHandler.UpdateNotificationSettings)
	protectedUserGroup.PUT("/settings/privacy", userHandler.UpdatePrivacySettings)

	// Follow operations
	protectedUserGroup.POST("/:id/follow", userHandler.FollowUser)
	protectedUserGroup.DELETE("/:id/follow", userHandler.UnfollowUser)
	protectedUserGroup.DELETE("/:id/follow-request", userHandler.CancelFollowRequest)
	protectedUserGroup.GET("/follow-requests", userHandler.GetFollowRequests)
	protectedUserGroup.POST("/follow-requests/:id/approve", userHandler.ApproveFollowRequest)
	protectedUserGroup.POST("/follow-requests/:id/reject", userHandler.RejectFollowRequest)
	protectedUserGroup.DELETE("/followers/:id", userHandler.RemoveFollower)

	// Followers and following
	protectedUserGroup.GET("/:id/followers", userHandler.GetFollowers)
	protectedUserGroup.GET("/:id/following", userHandler.GetFollowing)

	// Blocking and muting
	protectedUserGroup.POST("/:id/block", userHandler.BlockUser)
	protectedUserGroup.DELETE("/:id/block", userHandler.UnblockUser)
	protectedUserGroup.GET("/blocked", userHandler.GetBlockedUsers)
	protectedUserGroup.POST("/:id/mute", userHandler.MuteUser)
	protectedUserGroup.DELETE("/:id/mute", userHandler.UnmuteUser)
	protectedUserGroup.GET("/muted", userHandler.GetMutedUsers)

	// Verification
	protectedUserGroup.POST("/verify/email", userHandler.SendVerificationEmail)
	protectedUserGroup.POST("/verify/phone", userHandler.SendPhoneVerification)
	protectedUserGroup.POST("/verify/phone/confirm", userHandler.VerifyPhone)
	protectedUserGroup.POST("/verify/account", userHandler.RequestVerification)
	protectedUserGroup.GET("/verify/status", userHandler.GetVerificationStatus)

	// Suggestions
	protectedUserGroup.GET("/suggestions", userHandler.GetSuggestedUsers)
	protectedUserGroup.GET("/suggestions/interests", userHandler.GetSuggestedUsersByInterests)
	protectedUserGroup.GET("/suggestions/mutual", userHandler.GetSuggestedUsersByMutualFriends)
	protectedUserGroup.POST("/suggestions/:id/dismiss", userHandler.DismissSuggestion)

	// User analytics
	protectedUserGroup.GET("/analytics", userHandler.GetUserAnalytics)
	protectedUserGroup.GET("/analytics/engagement", userHandler.GetEngagementMetrics)
	protectedUserGroup.GET("/analytics/followers", userHandler.GetFollowerGrowth)
	protectedUserGroup.GET("/analytics/profile-views", userHandler.GetProfileViewers)

	// Data export
	protectedUserGroup.POST("/export", userHandler.RequestDataExport)
	protectedUserGroup.GET("/export", userHandler.GetDataExportStatus)
	protectedUserGroup.GET("/export/:id", userHandler.GetDataExportStatus)
	protectedUserGroup.GET("/export/:id/download", userHandler.DownloadDataExport)
	protectedUserGroup.DELETE("/export/:id", userHandler.DeleteDataExport)
}
