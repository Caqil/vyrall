package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/auth"
	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes configures the authentication routes
func SetupAuthRoutes(router *gin.Engine, authHandler *auth.Handler, authMiddleware gin.HandlerFunc) {
	// Auth routes group
	authGroup := router.Group("/api/auth")

	// Public authentication endpoints
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh-token", authHandler.RefreshToken)
	authGroup.POST("/forgot-password", authHandler.ForgotPassword)
	authGroup.POST("/reset-password", authHandler.ResetPassword)
	authGroup.POST("/verify-email", authHandler.VerifyEmail)
	authGroup.POST("/resend-verification", authHandler.ResendVerification)
	authGroup.POST("/validate-token", authHandler.ValidateToken)

	// OAuth routes
	authGroup.GET("/oauth/google", authHandler.GoogleOAuthRedirect)
	authGroup.GET("/oauth/google/callback", authHandler.GoogleOAuthCallback)
	authGroup.GET("/oauth/facebook", authHandler.FacebookOAuthRedirect)
	authGroup.GET("/oauth/facebook/callback", authHandler.FacebookOAuthCallback)
	authGroup.GET("/oauth/twitter", authHandler.TwitterOAuthRedirect)
	authGroup.GET("/oauth/twitter/callback", authHandler.TwitterOAuthCallback)
	authGroup.GET("/oauth/apple", authHandler.AppleOAuthRedirect)
	authGroup.GET("/oauth/apple/callback", authHandler.AppleOAuthCallback)

	// Protected authentication endpoints (require authentication)
	protectedAuthGroup := authGroup.Group("")
	protectedAuthGroup.Use(authMiddleware)

	protectedAuthGroup.POST("/logout", authHandler.Logout)
	protectedAuthGroup.POST("/change-password", authHandler.ChangePassword)
	protectedAuthGroup.GET("/sessions", authHandler.ListSessions)
	protectedAuthGroup.DELETE("/sessions/:id", authHandler.DeleteSession)
	protectedAuthGroup.POST("/two-factor/setup", authHandler.SetupTwoFactor)
	protectedAuthGroup.POST("/two-factor/enable", authHandler.EnableTwoFactor)
	protectedAuthGroup.POST("/two-factor/disable", authHandler.DisableTwoFactor)
	protectedAuthGroup.POST("/two-factor/verify", authHandler.VerifyTwoFactor)
	protectedAuthGroup.GET("/two-factor/backup-codes", authHandler.GetTwoFactorBackupCodes)
	protectedAuthGroup.POST("/two-factor/regenerate-backup-codes", authHandler.RegenerateTwoFactorBackupCodes)
}
