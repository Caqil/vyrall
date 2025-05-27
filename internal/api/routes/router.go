package routes

import (
	"github.com/Caqil/vyrall/internal/api/handlers"
	"github.com/Caqil/vyrall/internal/config"
	"github.com/Caqil/vyrall/internal/middleware"
	"github.com/Caqil/vyrall/internal/services"
	"github.com/Caqil/vyrall/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the API router
func SetupRouter(config *config.Config, services *services.Services, log *logger.Logger) *gin.Engine {
	// Create router
	router := gin.New()

	// Apply global middleware
	router.Use(middleware.Recovery(log))
	router.Use(middleware.Logging(log))
	router.Use(middleware.CORS())
	router.Use(middleware.Security())
	router.Use(middleware.ErrorHandler())

	// Conditionally apply middleware based on configuration
	if config.EnableCompression {
		router.Use(middleware.Compression())
	}

	if config.EnableMetrics {
		router.Use(middleware.Metrics(services.MetricsService))
	}

	// Create handlers
	handlers := handlers.NewHandlers(services)

	// Create authentication middleware
	authMiddleware := middleware.Auth(services.AuthService)
	optionalAuth := middleware.OptionalAuth(services.AuthService)

	// Setup API routes
	router.GET("/api/health", handlers.Health.Check)
	router.GET("/api/version", handlers.Health.Version)

	// Setup domain-specific routes
	SetupAdminRoutes(router, handlers.Admin, authMiddleware)
	SetupAuthRoutes(router, handlers.Auth, authMiddleware)
	SetupBusinessRoutes(router, handlers.Business, authMiddleware, optionalAuth)
	SetupCommentRoutes(router, handlers.Comments, authMiddleware, optionalAuth)
	SetupEventRoutes(router, handlers.Events, authMiddleware, optionalAuth)
	SetupGroupRoutes(router, handlers.Groups, authMiddleware, optionalAuth)
	SetupHashtagRoutes(router, handlers.Hashtags, authMiddleware, optionalAuth)
	SetupLiveRoutes(router, handlers.Live, authMiddleware, optionalAuth)
	SetupMediaRoutes(router, handlers.Media, authMiddleware, optionalAuth)
	SetupMessageRoutes(router, handlers.Messages, authMiddleware)
	SetupNotificationRoutes(router, handlers.Notifications, authMiddleware)
	SetupPostRoutes(router, handlers.Posts, authMiddleware, optionalAuth)
	SetupSearchRoutes(router, handlers.Search, authMiddleware, optionalAuth)
	SetupStoryRoutes(router, handlers.Stories, authMiddleware, optionalAuth)
	SetupUserRoutes(router, handlers.Users, authMiddleware, optionalAuth)

	// Setup static file serving if enabled
	if config.ServeStaticFiles {
		router.Static("/static", "./static")
	}

	// Setup Swagger documentation if enabled
	if config.EnableSwagger {
		router.GET("/swagger/*any", handlers.Swagger.Handler)
	}

	// Setup 404 handler for undefined routes
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"status": "error", "message": "Route not found"})
	})

	return router
}
