package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/comments"
	"github.com/gin-gonic/gin"
)

// SetupCommentRoutes configures the comment routes
func SetupCommentRoutes(router *gin.Engine, commentHandler *comments.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Comments routes group
	commentGroup := router.Group("/api/comments")

	// Public comment endpoints
	commentGroup.Use(optionalAuth)
	commentGroup.GET("/:id", commentHandler.GetComment)
	commentGroup.GET("/post/:postId", commentHandler.GetPostComments)

	// Protected comment endpoints (require authentication)
	protectedCommentGroup := commentGroup.Group("")
	protectedCommentGroup.Use(authMiddleware)

	// Comment operations
	protectedCommentGroup.POST("", commentHandler.CreateComment)
	protectedCommentGroup.PUT("/:id", commentHandler.UpdateComment)
	protectedCommentGroup.DELETE("/:id", commentHandler.DeleteComment)

	// Comment interactions
	protectedCommentGroup.POST("/:id/like", commentHandler.LikeComment)
	protectedCommentGroup.DELETE("/:id/like", commentHandler.UnlikeComment)
	protectedCommentGroup.GET("/:id/likes", commentHandler.GetCommentLikes)

	// Replies
	protectedCommentGroup.POST("/:id/replies", commentHandler.CreateReply)
	protectedCommentGroup.GET("/:id/replies", commentHandler.GetReplies)

	// Reporting
	protectedCommentGroup.POST("/:id/report", commentHandler.ReportComment)

	// Pinning comments (for post owners)
	protectedCommentGroup.POST("/:id/pin", commentHandler.PinComment)
	protectedCommentGroup.POST("/:id/unpin", commentHandler.UnpinComment)

	// Moderation
	protectedCommentGroup.POST("/:id/hide", commentHandler.HideComment)
	protectedCommentGroup.POST("/:id/unhide", commentHandler.UnhideComment)

	// Batch operations
	protectedCommentGroup.POST("/batch/delete", commentHandler.BatchDeleteComments)
}
