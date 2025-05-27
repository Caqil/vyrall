package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/events"
	"github.com/gin-gonic/gin"
)

// SetupEventRoutes configures the event routes
func SetupEventRoutes(router *gin.Engine, eventHandler *events.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Events routes group
	eventGroup := router.Group("/api/events")

	// Public event endpoints
	eventGroup.Use(optionalAuth)
	eventGroup.GET("", eventHandler.ListEvents)
	eventGroup.GET("/:id", eventHandler.GetEvent)
	eventGroup.GET("/featured", eventHandler.GetFeaturedEvents)
	eventGroup.GET("/nearby", eventHandler.GetNearbyEvents)
	eventGroup.GET("/categories", eventHandler.GetEventCategories)
	eventGroup.GET("/search", eventHandler.SearchEvents)

	// Protected event endpoints (require authentication)
	protectedEventGroup := eventGroup.Group("")
	protectedEventGroup.Use(authMiddleware)

	// Event operations
	protectedEventGroup.POST("", eventHandler.CreateEvent)
	protectedEventGroup.PUT("/:id", eventHandler.UpdateEvent)
	protectedEventGroup.DELETE("/:id", eventHandler.DeleteEvent)

	// Event management
	protectedEventGroup.POST("/:id/publish", eventHandler.PublishEvent)
	protectedEventGroup.POST("/:id/cancel", eventHandler.CancelEvent)
	protectedEventGroup.POST("/:id/reschedule", eventHandler.RescheduleEvent)

	// Attendee management
	protectedEventGroup.GET("/:id/attendees", eventHandler.GetEventAttendees)
	protectedEventGroup.POST("/:id/attendees", eventHandler.AddEventAttendees)
	protectedEventGroup.DELETE("/:id/attendees/:userId", eventHandler.RemoveEventAttendee)

	// RSVP
	protectedEventGroup.POST("/:id/rsvp", eventHandler.RSVP)
	protectedEventGroup.GET("/:id/rsvp", eventHandler.GetRSVPStatus)
	protectedEventGroup.DELETE("/:id/rsvp", eventHandler.CancelRSVP)

	// User's events
	protectedEventGroup.GET("/my/hosting", eventHandler.GetHostingEvents)
	protectedEventGroup.GET("/my/attending", eventHandler.GetAttendingEvents)
	protectedEventGroup.GET("/my/interested", eventHandler.GetInterestedEvents)

	// Analytics
	protectedEventGroup.GET("/:id/analytics", eventHandler.GetEventAnalytics)

	// Check-in
	protectedEventGroup.POST("/:id/checkin", eventHandler.CheckInAttendee)
	protectedEventGroup.GET("/:id/checkins", eventHandler.GetEventCheckIns)

	// Tickets
	protectedEventGroup.GET("/:id/tickets", eventHandler.GetEventTickets)
	protectedEventGroup.POST("/:id/tickets", eventHandler.CreateEventTicket)
	protectedEventGroup.PUT("/:id/tickets/:ticketId", eventHandler.UpdateEventTicket)
	protectedEventGroup.DELETE("/:id/tickets/:ticketId", eventHandler.DeleteEventTicket)

	// Event invitations
	protectedEventGroup.POST("/:id/invitations", eventHandler.SendEventInvitations)
	protectedEventGroup.GET("/invitations", eventHandler.GetEventInvitations)
	protectedEventGroup.POST("/invitations/:id/respond", eventHandler.RespondToEventInvitation)
}
