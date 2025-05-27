package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/media"
	"github.com/gin-gonic/gin"
)

// SetupMediaRoutes configures the media routes
func SetupMediaRoutes(router *gin.Engine, mediaHandler *media.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Media routes group
	mediaGroup := router.Group("/api/media")

	// Public media endpoints
	mediaGroup.Use(optionalAuth)
	mediaGroup.GET("/:id", mediaHandler.GetMedia)

	// Protected media endpoints (require authentication)
	protectedMediaGroup := mediaGroup.Group("")
	protectedMediaGroup.Use(authMiddleware)

	// Media upload
	protectedMediaGroup.POST("/upload", mediaHandler.UploadMedia)
	protectedMediaGroup.POST("/upload/multiple", mediaHandler.UploadMultipleMedia)
	protectedMediaGroup.POST("/upload/chunked/start", mediaHandler.StartChunkedUpload)
	protectedMediaGroup.POST("/upload/chunked/:uploadId", mediaHandler.UploadChunk)
	protectedMediaGroup.POST("/upload/chunked/:uploadId/complete", mediaHandler.CompleteChunkedUpload)
	protectedMediaGroup.DELETE("/upload/chunked/:uploadId", mediaHandler.CancelChunkedUpload)

	// Media management
	protectedMediaGroup.PUT("/:id", mediaHandler.UpdateMedia)
	protectedMediaGroup.DELETE("/:id", mediaHandler.DeleteMedia)
	protectedMediaGroup.POST("/batch/delete", mediaHandler.BatchDeleteMedia)

	// Media processing
	protectedMediaGroup.POST("/:id/process", mediaHandler.ProcessMedia)
	protectedMediaGroup.GET("/:id/status", mediaHandler.GetProcessingStatus)

	// Media organization
	protectedMediaGroup.GET("/library", mediaHandler.GetMediaLibrary)
	protectedMediaGroup.POST("/collections", mediaHandler.CreateCollection)
	protectedMediaGroup.GET("/collections", mediaHandler.GetCollections)
	protectedMediaGroup.GET("/collections/:id", mediaHandler.GetCollection)
	protectedMediaGroup.PUT("/collections/:id", mediaHandler.UpdateCollection)
	protectedMediaGroup.DELETE("/collections/:id", mediaHandler.DeleteCollection)
	protectedMediaGroup.POST("/collections/:id/media", mediaHandler.AddMediaToCollection)
	protectedMediaGroup.DELETE("/collections/:id/media/:mediaId", mediaHandler.RemoveMediaFromCollection)

	// Media sharing
	protectedMediaGroup.POST("/:id/share", mediaHandler.ShareMedia)

	// Media analytics
	protectedMediaGroup.GET("/:id/analytics", mediaHandler.GetMediaAnalytics)

	// Media transcoding
	protectedMediaGroup.POST("/:id/transcode", mediaHandler.TranscodeMedia)
	protectedMediaGroup.GET("/:id/formats", mediaHandler.GetMediaFormats)

	// Download
	protectedMediaGroup.GET("/:id/download", mediaHandler.DownloadMedia)

	// Media search
	protectedMediaGroup.GET("/search", mediaHandler.SearchMedia)
}
