package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/business"
	"github.com/gin-gonic/gin"
)

// SetupBusinessRoutes configures the business routes
func SetupBusinessRoutes(router *gin.Engine, businessHandler *business.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Business routes group
	businessGroup := router.Group("/api/business")

	// Public business endpoints
	businessGroup.Use(optionalAuth)
	businessGroup.GET("/profiles/:id", businessHandler.GetBusinessProfile)
	businessGroup.GET("/categories", businessHandler.GetBusinessCategories)
	businessGroup.GET("/popular", businessHandler.GetPopularBusinesses)

	// Protected business endpoints (require authentication)
	protectedBusinessGroup := businessGroup.Group("")
	protectedBusinessGroup.Use(authMiddleware)

	// Business profile management
	protectedBusinessGroup.POST("/profiles", businessHandler.CreateBusinessProfile)
	protectedBusinessGroup.PUT("/profiles/:id", businessHandler.UpdateBusinessProfile)
	protectedBusinessGroup.DELETE("/profiles/:id", businessHandler.DeleteBusinessProfile)

	// Business page management
	protectedBusinessGroup.GET("/pages", businessHandler.ListBusinessPages)
	protectedBusinessGroup.POST("/pages", businessHandler.CreateBusinessPage)
	protectedBusinessGroup.GET("/pages/:id", businessHandler.GetBusinessPage)
	protectedBusinessGroup.PUT("/pages/:id", businessHandler.UpdateBusinessPage)
	protectedBusinessGroup.DELETE("/pages/:id", businessHandler.DeleteBusinessPage)

	// Business analytics
	protectedBusinessGroup.GET("/analytics/:id", businessHandler.GetBusinessAnalytics)
	protectedBusinessGroup.GET("/analytics/:id/audience", businessHandler.GetBusinessAudience)
	protectedBusinessGroup.GET("/analytics/:id/engagement", businessHandler.GetBusinessEngagement)

	// Ad campaigns
	protectedBusinessGroup.GET("/campaigns", businessHandler.ListAdCampaigns)
	protectedBusinessGroup.POST("/campaigns", businessHandler.CreateAdCampaign)
	protectedBusinessGroup.GET("/campaigns/:id", businessHandler.GetAdCampaign)
	protectedBusinessGroup.PUT("/campaigns/:id", businessHandler.UpdateAdCampaign)
	protectedBusinessGroup.DELETE("/campaigns/:id", businessHandler.DeleteAdCampaign)
	protectedBusinessGroup.POST("/campaigns/:id/pause", businessHandler.PauseAdCampaign)
	protectedBusinessGroup.POST("/campaigns/:id/resume", businessHandler.ResumeAdCampaign)

	// Ads
	protectedBusinessGroup.GET("/ads", businessHandler.ListAds)
	protectedBusinessGroup.POST("/ads", businessHandler.CreateAd)
	protectedBusinessGroup.GET("/ads/:id", businessHandler.GetAd)
	protectedBusinessGroup.PUT("/ads/:id", businessHandler.UpdateAd)
	protectedBusinessGroup.DELETE("/ads/:id", businessHandler.DeleteAd)

	// Payments
	protectedBusinessGroup.GET("/payments", businessHandler.ListPayments)
	protectedBusinessGroup.GET("/payments/:id", businessHandler.GetPayment)
	protectedBusinessGroup.POST("/payments/methods", businessHandler.AddPaymentMethod)
	protectedBusinessGroup.GET("/payments/methods", businessHandler.ListPaymentMethods)
	protectedBusinessGroup.DELETE("/payments/methods/:id", businessHandler.DeletePaymentMethod)
}
