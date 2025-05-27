package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers/search"
	"github.com/gin-gonic/gin"
)

// SetupSearchRoutes configures the search routes
func SetupSearchRoutes(router *gin.Engine, searchHandler *search.Handler, authMiddleware, optionalAuth gin.HandlerFunc) {
	// Search routes group
	searchGroup := router.Group("/api/search")

	// Public search endpoints
	searchGroup.Use(optionalAuth)
	searchGroup.GET("", searchHandler.GlobalSearch)
	searchGroup.GET("/users", searchHandler.SearchUsers)
	searchGroup.GET("/posts", searchHandler.SearchPosts)
	searchGroup.GET("/hashtags", searchHandler.SearchHashtags)
	searchGroup.GET("/groups", searchHandler.SearchGroups)
	searchGroup.GET("/events", searchHandler.SearchEvents)
	searchGroup.GET("/locations", searchHandler.SearchLocations)
	searchGroup.GET("/trending", searchHandler.GetTrendingSearches)

	// Protected search endpoints (require authentication)
	protectedSearchGroup := searchGroup.Group("")
	protectedSearchGroup.Use(authMiddleware)

	// Search history and suggestions
	protectedSearchGroup.GET("/history", searchHandler.GetSearchHistory)
	protectedSearchGroup.DELETE("/history", searchHandler.ClearSearchHistory)
	protectedSearchGroup.DELETE("/history/:id", searchHandler.DeleteSearchHistoryItem)
	protectedSearchGroup.GET("/suggestions", searchHandler.GetSearchSuggestions)

	// Advanced search
	protectedSearchGroup.POST("/advanced", searchHandler.AdvancedSearch)
	protectedSearchGroup.GET("/filters", searchHandler.GetAvailableFilters)

	// Personalized search
	protectedSearchGroup.GET("/people", searchHandler.SearchPeople)
	protectedSearchGroup.GET("/people/nearby", searchHandler.SearchNearbyPeople)
	protectedSearchGroup.GET("/people/interests", searchHandler.SearchPeopleByInterests)

	// Content search
	protectedSearchGroup.GET("/content", searchHandler.SearchContent)
	protectedSearchGroup.GET("/messages", searchHandler.SearchMessages)

	// Save searches
	protectedSearchGroup.POST("/save", searchHandler.SaveSearch)
	protectedSearchGroup.GET("/saved", searchHandler.GetSavedSearches)
	protectedSearchGroup.DELETE("/saved/:id", searchHandler.DeleteSavedSearch)
}
