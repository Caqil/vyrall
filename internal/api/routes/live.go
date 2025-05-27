package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/live"
	"github.com/gin-gonic/gin"
)

// SetupLiveRoutes configures the live streaming routes
func SetupLiveRoutes(router *gin.Engine, liveHandler *live.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Live routes group
	liveGroup := router.Group("/api/live")

	// Public live endpoints
	liveGroup.Use(optionalAuth)
	liveGroup.GET("", liveHandler.ListLiveStreams)
	liveGroup.GET("/:id", liveHandler.GetLiveStream)
	liveGroup.GET("/categories", liveHandler.GetLiveCategories)
	liveGroup.GET("/featured", liveHandler.GetFeaturedStreams)
	liveGroup.GET("/trending", liveHandler.GetTrendingStreams)

	// Protected live endpoints (require authentication)
	protectedLiveGroup := liveGroup.Group("")
	protectedLiveGroup.Use(authMiddleware)

	// Stream management
	protectedLiveGroup.POST("", liveHandler.CreateLiveStream)
	protectedLiveGroup.PUT("/:id", liveHandler.UpdateLiveStream)
	protectedLiveGroup.DELETE("/:id", liveHandler.DeleteLiveStream)
	protectedLiveGroup.POST("/:id/start", liveHandler.StartLiveStream)
	protectedLiveGroup.POST("/:id/stop", liveHandler.StopLiveStream)
	protectedLiveGroup.POST("/:id/pause", liveHandler.PauseLiveStream)
	protectedLiveGroup.POST("/:id/resume", liveHandler.ResumeLiveStream)

	// Stream viewing
	protectedLiveGroup.POST("/:id/join", liveHandler.JoinLiveStream)
	protectedLiveGroup.POST("/:id/leave", liveHandler.LeaveLiveStream)
	protectedLiveGroup.GET("/:id/viewers", liveHandler.GetStreamViewers)

	// Stream interactions
	protectedLiveGroup.POST("/:id/like", liveHandler.LikeLiveStream)
	protectedLiveGroup.DELETE("/:id/like", liveHandler.UnlikeLiveStream)
	protectedLiveGroup.POST("/:id/comments", liveHandler.AddStreamComment)
	protectedLiveGroup.GET("/:id/comments", liveHandler.GetStreamComments)
	protectedLiveGroup.DELETE("/:id/comments/:commentId", liveHandler.DeleteStreamComment)

	// Stream moderation
	protectedLiveGroup.POST("/:id/moderators", liveHandler.AddModerator)
	protectedLiveGroup.DELETE("/:id/moderators/:userId", liveHandler.RemoveModerator)
	protectedLiveGroup.POST("/:id/ban", liveHandler.BanUser)
	protectedLiveGroup.DELETE("/:id/ban/:userId", liveHandler.UnbanUser)
	protectedLiveGroup.POST("/:id/mute/:userId", liveHandler.MuteUser)
	protectedLiveGroup.DELETE("/:id/mute/:userId", liveHandler.UnmuteUser)

	// Stream settings
	protectedLiveGroup.GET("/:id/settings", liveHandler.GetStreamSettings)
	protectedLiveGroup.PUT("/:id/settings", liveHandler.UpdateStreamSettings)

	// Stream analytics
	protectedLiveGroup.GET("/:id/analytics", liveHandler.GetStreamAnalytics)

	// User's streams
	protectedLiveGroup.GET("/my/streams", liveHandler.GetMyStreams)
	protectedLiveGroup.GET("/my/scheduled", liveHandler.GetMyScheduledStreams)
	protectedLiveGroup.GET("/following", liveHandler.GetFollowingStreams)

	// Stream recordings
	protectedLiveGroup.GET("/:id/recordings", liveHandler.GetStreamRecordings)
	protectedLiveGroup.DELETE("/:id/recordings/:recordingId", liveHandler.DeleteStreamRecording)
}
