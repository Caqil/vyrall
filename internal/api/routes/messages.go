package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/messages"
	"github.com/gin-gonic/gin"
)

// SetupMessageRoutes configures the messaging routes
func SetupMessageRoutes(router *gin.Engine, messageHandler *messages.Handler, authMiddleware gin.HandlerFunc) {
	// Messages routes group
	messageGroup := router.Group("/api/messages")

	// All messaging endpoints require authentication
	messageGroup.Use(authMiddleware)

	// Conversations
	messageGroup.GET("/conversations", messageHandler.GetConversations)
	messageGroup.POST("/conversations", messageHandler.CreateConversation)
	messageGroup.GET("/conversations/:id", messageHandler.GetConversation)
	messageGroup.PUT("/conversations/:id", messageHandler.UpdateConversation)
	messageGroup.DELETE("/conversations/:id", messageHandler.DeleteConversation)
	messageGroup.POST("/conversations/:id/archive", messageHandler.ArchiveConversation)
	messageGroup.POST("/conversations/:id/unarchive", messageHandler.UnarchiveConversation)
	messageGroup.POST("/conversations/:id/mute", messageHandler.MuteConversation)
	messageGroup.POST("/conversations/:id/unmute", messageHandler.UnmuteConversation)

	// Group chat management
	messageGroup.POST("/conversations/:id/members", messageHandler.AddMembers)
	messageGroup.DELETE("/conversations/:id/members/:userId", messageHandler.RemoveMember)
	messageGroup.POST("/conversations/:id/members/:userId/admin", messageHandler.MakeMemberAdmin)
	messageGroup.DELETE("/conversations/:id/members/:userId/admin", messageHandler.RemoveMemberAdmin)
	messageGroup.POST("/conversations/:id/leave", messageHandler.LeaveConversation)

	// Messages
	messageGroup.GET("/conversations/:id/messages", messageHandler.GetMessages)
	messageGroup.POST("/conversations/:id/messages", messageHandler.SendMessage)
	messageGroup.PUT("/messages/:id", messageHandler.UpdateMessage)
	messageGroup.DELETE("/messages/:id", messageHandler.DeleteMessage)

	// Message interactions
	messageGroup.POST("/messages/:id/read", messageHandler.MarkAsRead)
	messageGroup.POST("/conversations/:id/read", messageHandler.MarkConversationAsRead)
	messageGroup.POST("/messages/:id/react", messageHandler.ReactToMessage)
	messageGroup.DELETE("/messages/:id/react/:reaction", messageHandler.RemoveReaction)
	messageGroup.POST("/messages/:id/forward", messageHandler.ForwardMessage)
	messageGroup.POST("/messages/:id/reply", messageHandler.ReplyToMessage)

	// Message attachments
	messageGroup.POST("/messages/:id/attachments", messageHandler.AddAttachment)
	messageGroup.DELETE("/messages/:id/attachments/:attachmentId", messageHandler.RemoveAttachment)

	// Message search
	messageGroup.GET("/search", messageHandler.SearchMessages)

	// Message encryption
	messageGroup.POST("/conversations/:id/encrypt", messageHandler.EnableEncryption)
	messageGroup.POST("/conversations/:id/decrypt", messageHandler.DisableEncryption)
	messageGroup.GET("/conversations/:id/keys", messageHandler.GetEncryptionKeys)
	messageGroup.POST("/conversations/:id/keys", messageHandler.UpdateEncryptionKeys)

	// User presence
	messageGroup.GET("/presence", messageHandler.GetUserPresence)
	messageGroup.POST("/presence", messageHandler.UpdateUserPresence)

	// Disappearing messages
	messageGroup.POST("/conversations/:id/disappearing", messageHandler.EnableDisappearingMessages)
	messageGroup.DELETE("/conversations/:id/disappearing", messageHandler.DisableDisappearingMessages)
}
