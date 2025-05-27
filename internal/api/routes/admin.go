package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/admin"
	"github.com/Caqil/vyrall/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes configures the admin routes
func SetupAdminRoutes(router *gin.Engine, adminHandler *admin.Handler, authMiddleware gin.HandlerFunc) {
	// Admin routes group
	adminGroup := router.Group("/api/admin")

	// Apply authentication and admin-only middleware to all admin routes
	adminGroup.Use(authMiddleware)
	adminGroup.Use(middleware.AdminOnly())

	// Dashboard
	adminGroup.GET("/dashboard", adminHandler.GetDashboard)

	// Users management
	adminGroup.GET("/users", adminHandler.ListUsers)
	adminGroup.GET("/users/:id", adminHandler.GetUser)
	adminGroup.PUT("/users/:id", adminHandler.UpdateUser)
	adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)
	adminGroup.POST("/users/:id/ban", adminHandler.BanUser)
	adminGroup.POST("/users/:id/unban", adminHandler.UnbanUser)
	adminGroup.POST("/users/:id/verify", adminHandler.VerifyUser)
	adminGroup.POST("/users/:id/unverify", adminHandler.UnverifyUser)

	// Content moderation
	adminGroup.GET("/reports", adminHandler.ListReports)
	adminGroup.GET("/reports/:id", adminHandler.GetReport)
	adminGroup.POST("/reports/:id/resolve", adminHandler.ResolveReport)
	adminGroup.POST("/reports/:id/dismiss", adminHandler.DismissReport)

	// Posts management
	adminGroup.GET("/posts", adminHandler.ListPosts)
	adminGroup.DELETE("/posts/:id", adminHandler.DeletePost)
	adminGroup.POST("/posts/:id/feature", adminHandler.FeaturePost)
	adminGroup.POST("/posts/:id/unfeature", adminHandler.UnfeaturePost)

	// Comments management
	adminGroup.GET("/comments", adminHandler.ListComments)
	adminGroup.DELETE("/comments/:id", adminHandler.DeleteComment)

	// System settings
	adminGroup.GET("/settings", adminHandler.GetSettings)
	adminGroup.PUT("/settings", adminHandler.UpdateSettings)

	// Analytics
	adminGroup.GET("/analytics/overview", adminHandler.GetAnalyticsOverview)
	adminGroup.GET("/analytics/users", adminHandler.GetUserAnalytics)
	adminGroup.GET("/analytics/content", adminHandler.GetContentAnalytics)
	adminGroup.GET("/analytics/engagement", adminHandler.GetEngagementAnalytics)

	// Logs
	adminGroup.GET("/logs", adminHandler.GetLogs)
	adminGroup.GET("/logs/errors", adminHandler.GetErrorLogs)
	adminGroup.GET("/logs/access", adminHandler.GetAccessLogs)
}
