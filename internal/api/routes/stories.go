package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/stories"
	"github.com/gin-gonic/gin"
)

// SetupStoryRoutes configures the story routes
func SetupStoryRoutes(router *gin.Engine, storyHandler *stories.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Stories routes group
	storyGroup := router.Group("/api/stories")

	// Public story endpoints
	storyGroup.Use(optionalAuth)
	storyGroup.GET("/:id", storyHandler.GetStory)
	storyGroup.GET("/user/:userId", storyHandler.GetUserStories)

	// Protected story endpoints (require authentication)
	protectedStoryGroup := storyGroup.Group("")
	protectedStoryGroup.Use(authMiddleware)

	// Story creation and management
	protectedStoryGroup.POST("", storyHandler.CreateStory)
	protectedStoryGroup.DELETE("/:id", storyHandler.DeleteStory)

	// Story viewing
	protectedStoryGroup.GET("/feed", storyHandler.GetStoryFeed)
	protectedStoryGroup.POST("/:id/view", storyHandler.MarkAsViewed)
	protectedStoryGroup.GET("/viewed", storyHandler.GetViewedStories)
	protectedStoryGroup.GET("/unviewed", storyHandler.GetUnviewedStories)

	// Story interactions
	protectedStoryGroup.POST("/:id/react", storyHandler.AddReaction)
	protectedStoryGroup.DELETE("/:id/react", storyHandler.RemoveReaction)
	protectedStoryGroup.GET("/:id/reactions", storyHandler.GetReactions)
	protectedStoryGroup.POST("/:id/reply", storyHandler.ReplyToStory)
	protectedStoryGroup.GET("/:id/replies", storyHandler.GetStoryReplies)

	// Story viewers
	protectedStoryGroup.GET("/:id/viewers", storyHandler.GetStoryViewers)
	protectedStoryGroup.GET("/:id/viewers/stats", storyHandler.GetStoryViewerStats)

	// Story highlights
	protectedStoryGroup.POST("/highlights", storyHandler.CreateHighlight)
	protectedStoryGroup.GET("/highlights", storyHandler.GetMyHighlights)
	protectedStoryGroup.GET("/highlights/:id", storyHandler.GetHighlight)
	protectedStoryGroup.PUT("/highlights/:id", storyHandler.UpdateHighlight)
	protectedStoryGroup.DELETE("/highlights/:id", storyHandler.DeleteHighlight)
	protectedStoryGroup.GET("/user/:userId/highlights", storyHandler.GetUserHighlights)

	// Story archive
	protectedStoryGroup.POST("/:id/archive", storyHandler.ArchiveStory)
	protectedStoryGroup.POST("/:id/unarchive", storyHandler.UnarchiveStory)
	protectedStoryGroup.GET("/archived", storyHandler.GetArchivedStories)

	// Interactive stories
	protectedStoryGroup.POST("/poll", storyHandler.CreatePollStory)
	protectedStoryGroup.POST("/:id/poll/vote", storyHandler.VoteOnPoll)
	protectedStoryGroup.GET("/:id/poll/results", storyHandler.GetPollResults)
	protectedStoryGroup.POST("/question", storyHandler.CreateQuestionStory)
	protectedStoryGroup.POST("/:id/question/answer", storyHandler.AnswerQuestion)
	protectedStoryGroup.GET("/:id/question/answers", storyHandler.GetQuestionAnswers)

	// Story analytics
	protectedStoryGroup.GET("/:id/analytics", storyHandler.GetStoryAnalytics)
	protectedStoryGroup.GET("/analytics", storyHandler.GetStoriesAnalytics)
}
