package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/notifications"
	"github.com/gin-gonic/gin"
)

// SetupNotificationRoutes configures the notification routes
func SetupNotificationRoutes(router *gin.Engine, notificationHandler *notifications.Handler, authMiddleware gin.HandlerFunc) {
	// Notifications routes group
	notificationGroup := router.Group("/api/notifications")

	// All notification endpoints require authentication
	notificationGroup.Use(authMiddleware)

	// Notifications retrieval
	notificationGroup.GET("", notificationHandler.GetNotifications)
	notificationGroup.GET("/:id", notificationHandler.GetNotification)
	notificationGroup.GET("/unread", notificationHandler.GetUnreadNotifications)
	notificationGroup.GET("/count", notificationHandler.GetUnreadCount)

	// Notification management
	notificationGroup.POST("/:id/read", notificationHandler.MarkAsRead)
	notificationGroup.POST("/read-all", notificationHandler.MarkAllAsRead)
	notificationGroup.DELETE("/:id", notificationHandler.DeleteNotification)
	notificationGroup.DELETE("", notificationHandler.DeleteAllNotifications)

	// Notification settings
	notificationGroup.GET("/settings", notificationHandler.GetNotificationSettings)
	notificationGroup.PUT("/settings", notificationHandler.UpdateNotificationSettings)
	notificationGroup.PUT("/settings/:type", notificationHandler.UpdateTypeSettings)

	// Device management for push notifications
	notificationGroup.POST("/devices", notificationHandler.RegisterDevice)
	notificationGroup.GET("/devices", notificationHandler.GetRegisteredDevices)
	notificationGroup.DELETE("/devices/:id", notificationHandler.UnregisterDevice)

	// Notification categories and preferences
	notificationGroup.GET("/categories", notificationHandler.GetNotificationCategories)
	notificationGroup.PUT("/categories/:category", notificationHandler.UpdateCategoryPreference)

	// Do Not Disturb settings
	notificationGroup.GET("/do-not-disturb", notificationHandler.GetDoNotDisturbSettings)
	notificationGroup.PUT("/do-not-disturb", notificationHandler.UpdateDoNotDisturbSettings)
	notificationGroup.POST("/do-not-disturb/enable", notificationHandler.EnableDoNotDisturb)
	notificationGroup.POST("/do-not-disturb/disable", notificationHandler.DisableDoNotDisturb)

	// Notification actions
	notificationGroup.POST("/:id/action", notificationHandler.PerformNotificationAction)

	// Notification subscriptions
	notificationGroup.GET("/subscriptions", notificationHandler.GetSubscriptions)
	notificationGroup.POST("/subscriptions", notificationHandler.CreateSubscription)
	notificationGroup.DELETE("/subscriptions/:id", notificationHandler.DeleteSubscription)
}
