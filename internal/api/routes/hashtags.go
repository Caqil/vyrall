package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/hashtags"
	"github.com/gin-gonic/gin"
)

// SetupHashtagRoutes configures the hashtag routes
func SetupHashtagRoutes(router *gin.Engine, hashtagHandler *hashtags.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Hashtags routes group
	hashtagGroup := router.Group("/api/hashtags")

	// Public hashtag endpoints
	hashtagGroup.Use(optionalAuth)
	hashtagGroup.GET("", hashtagHandler.ListHashtags)
	hashtagGroup.GET("/:tag", hashtagHandler.GetHashtag)
	hashtagGroup.GET("/:tag/posts", hashtagHandler.GetHashtagPosts)
	hashtagGroup.GET("/trending", hashtagHandler.GetTrendingHashtags)
	hashtagGroup.GET("/search", hashtagHandler.SearchHashtags)

	// Protected hashtag endpoints (require authentication)
	protectedHashtagGroup := hashtagGroup.Group("")
	protectedHashtagGroup.Use(authMiddleware)

	// Hashtag follow operations
	protectedHashtagGroup.POST("/:tag/follow", hashtagHandler.FollowHashtag)
	protectedHashtagGroup.DELETE("/:tag/follow", hashtagHandler.UnfollowHashtag)
	protectedHashtagGroup.GET("/following", hashtagHandler.GetFollowingHashtags)

	// Hashtag analytics
	protectedHashtagGroup.GET("/:tag/analytics", hashtagHandler.GetHashtagAnalytics)

	// Related hashtags
	protectedHashtagGroup.GET("/:tag/related", hashtagHandler.GetRelatedHashtags)

	// User's hashtag feed
	protectedHashtagGroup.GET("/feed", hashtagHandler.GetHashtagFeed)

	// Hashtag reports
	protectedHashtagGroup.POST("/:tag/report", hashtagHandler.ReportHashtag)
}
