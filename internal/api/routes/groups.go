package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/groups"
	"github.com/gin-gonic/gin"
)

// SetupGroupRoutes configures the group routes
func SetupGroupRoutes(router *gin.Engine, groupHandler *groups.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Groups routes group
	groupGroup := router.Group("/api/groups")

	// Public group endpoints
	groupGroup.Use(optionalAuth)
	groupGroup.GET("", groupHandler.ListGroups)
	groupGroup.GET("/:id", groupHandler.GetGroup)
	groupGroup.GET("/:id/posts", groupHandler.GetGroupPosts)
	groupGroup.GET("/categories", groupHandler.GetGroupCategories)
	groupGroup.GET("/search", groupHandler.SearchGroups)
	groupGroup.GET("/popular", groupHandler.GetPopularGroups)

	// Protected group endpoints (require authentication)
	protectedGroupGroup := groupGroup.Group("")
	protectedGroupGroup.Use(authMiddleware)

	// Group operations
	protectedGroupGroup.POST("", groupHandler.CreateGroup)
	protectedGroupGroup.PUT("/:id", groupHandler.UpdateGroup)
	protectedGroupGroup.DELETE("/:id", groupHandler.DeleteGroup)

	// Group membership
	protectedGroupGroup.POST("/:id/join", groupHandler.JoinGroup)
	protectedGroupGroup.POST("/:id/leave", groupHandler.LeaveGroup)
	protectedGroupGroup.GET("/:id/members", groupHandler.GetGroupMembers)
	protectedGroupGroup.POST("/:id/invite", groupHandler.InviteToGroup)

	// Group membership requests
	protectedGroupGroup.GET("/:id/requests", groupHandler.GetJoinRequests)
	protectedGroupGroup.POST("/:id/requests/:userId/approve", groupHandler.ApproveJoinRequest)
	protectedGroupGroup.POST("/:id/requests/:userId/reject", groupHandler.RejectJoinRequest)

	// Group roles
	protectedGroupGroup.POST("/:id/members/:userId/promote", groupHandler.PromoteMember)
	protectedGroupGroup.POST("/:id/members/:userId/demote", groupHandler.DemoteMember)
	protectedGroupGroup.POST("/:id/members/:userId/remove", groupHandler.RemoveMember)

	// Group settings
	protectedGroupGroup.GET("/:id/settings", groupHandler.GetGroupSettings)
	protectedGroupGroup.PUT("/:id/settings", groupHandler.UpdateGroupSettings)

	// Group rules
	protectedGroupGroup.GET("/:id/rules", groupHandler.GetGroupRules)
	protectedGroupGroup.POST("/:id/rules", groupHandler.AddGroupRule)
	protectedGroupGroup.PUT("/:id/rules/:ruleId", groupHandler.UpdateGroupRule)
	protectedGroupGroup.DELETE("/:id/rules/:ruleId", groupHandler.DeleteGroupRule)

	// Group analytics
	protectedGroupGroup.GET("/:id/analytics", groupHandler.GetGroupAnalytics)

	// User's groups
	protectedGroupGroup.GET("/my/groups", groupHandler.GetMyGroups)
	protectedGroupGroup.GET("/my/managed", groupHandler.GetManagedGroups)
	protectedGroupGroup.GET("/my/invitations", groupHandler.GetGroupInvitations)

	// Group content
	protectedGroupGroup.POST("/:id/posts", groupHandler.CreateGroupPost)
	protectedGroupGroup.GET("/:id/media", groupHandler.GetGroupMedia)
	protectedGroupGroup.GET("/:id/events", groupHandler.GetGroupEvents)
	protectedGroupGroup.POST("/:id/events", groupHandler.CreateGroupEvent)
}
