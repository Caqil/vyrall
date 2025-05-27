package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/posts"
	"github.com/gin-gonic/gin"
)

// SetupPostRoutes configures the post routes
func SetupPostRoutes(router *gin.Engine, postHandler *posts.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Posts routes group
	postGroup := router.Group("/api/posts")

	// Public post endpoints
	postGroup.Use(optionalAuth)
	postGroup.GET("", postHandler.ListPosts)
	postGroup.GET("/:id", postHandler.GetPost)
	postGroup.GET("/trending", postHandler.GetTrendingPosts)
	postGroup.GET("/tag/:tag", postHandler.GetPostsByTag)
	postGroup.GET("/user/:userId", postHandler.GetUserPosts)

	// Protected post endpoints (require authentication)
	protectedPostGroup := postGroup.Group("")
	protectedPostGroup.Use(authMiddleware)

	// Post creation and management
	protectedPostGroup.POST("", postHandler.CreatePost)
	protectedPostGroup.PUT("/:id", postHandler.UpdatePost)
	protectedPostGroup.DELETE("/:id", postHandler.DeletePost)

	// Post interactions
	protectedPostGroup.POST("/:id/like", postHandler.LikePost)
	protectedPostGroup.DELETE("/:id/like", postHandler.UnlikePost)
	protectedPostGroup.GET("/:id/likes", postHandler.GetPostLikes)
	protectedPostGroup.POST("/:id/share", postHandler.SharePost)
	protectedPostGroup.GET("/:id/shares", postHandler.GetPostShares)

	// Post bookmarking
	protectedPostGroup.POST("/:id/bookmark", postHandler.BookmarkPost)
	protectedPostGroup.DELETE("/:id/bookmark", postHandler.UnbookmarkPost)
	protectedPostGroup.GET("/bookmarks", postHandler.GetBookmarkedPosts)

	// Post collections (for bookmarks)
	protectedPostGroup.GET("/collections", postHandler.GetBookmarkCollections)
	protectedPostGroup.POST("/collections", postHandler.CreateBookmarkCollection)
	protectedPostGroup.PUT("/collections/:id", postHandler.UpdateBookmarkCollection)
	protectedPostGroup.DELETE("/collections/:id", postHandler.DeleteBookmarkCollection)

	// Post reporting
	protectedPostGroup.POST("/:id/report", postHandler.ReportPost)

	// Post feeds
	protectedPostGroup.GET("/feed", postHandler.GetFeed)
	protectedPostGroup.GET("/feed/refresh", postHandler.RefreshFeed)
	protectedPostGroup.GET("/feed/explore", postHandler.GetExploreFeed)

	// Post analytics
	protectedPostGroup.GET("/:id/analytics", postHandler.GetPostAnalytics)

	// Polls
	protectedPostGroup.POST("/poll", postHandler.CreatePollPost)
	protectedPostGroup.POST("/:id/poll/vote", postHandler.VoteOnPoll)
	protectedPostGroup.GET("/:id/poll/results", postHandler.GetPollResults)

	// Scheduled posts
	protectedPostGroup.POST("/schedule", postHandler.SchedulePost)
	protectedPostGroup.GET("/scheduled", postHandler.GetScheduledPosts)
	protectedPostGroup.PUT("/scheduled/:id", postHandler.UpdateScheduledPost)
	protectedPostGroup.DELETE("/scheduled/:id", postHandler.DeleteScheduledPost)
	protectedPostGroup.POST("/scheduled/:id/publish", postHandler.PublishScheduledPost)

	// Draft posts
	protectedPostGroup.POST("/draft", postHandler.CreateDraft)
	protectedPostGroup.GET("/drafts", postHandler.GetDrafts)
	protectedPostGroup.PUT("/drafts/:id", postHandler.UpdateDraft)
	protectedPostGroup.DELETE("/drafts/:id", postHandler.DeleteDraft)
	protectedPostGroup.POST("/drafts/:id/publish", postHandler.PublishDraft)
}
