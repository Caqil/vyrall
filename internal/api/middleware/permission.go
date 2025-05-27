package middleware

import (
	"github.com/Caqil/vyrall/internal/services/permission"
	"github.com/Caqil/vyrall/internal/utils/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Permission middleware for checking user permissions
func Permission(permissionService *permission.Service, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by Auth middleware)
		userIDValue, exists := c.Get("userID")
		if !exists {
			response.UnauthorizedError(c, "User not authenticated")
			c.Abort()
			return
		}

		userID := userIDValue.(primitive.ObjectID)

		// Check if user has the required permission
		hasPermission, err := permissionService.HasPermission(c.Request.Context(), userID, requiredPermission)
		if err != nil {
			response.Error(c, 500, "Failed to check permission", err)
			c.Abort()
			return
		}

		if !hasPermission {
			response.ForbiddenError(c, "Permission denied")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ResourceOwner middleware to check if the user owns a resource
func ResourceOwner(permissionService *permission.Service, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by Auth middleware)
		userIDValue, exists := c.Get("userID")
		if !exists {
			response.UnauthorizedError(c, "User not authenticated")
			c.Abort()
			return
		}

		userID := userIDValue.(primitive.ObjectID)

		// Get resource ID from URL parameter
		resourceIDStr := c.Param("id")
		if resourceIDStr == "" {
			response.ValidationError(c, "Resource ID is required", nil)
			c.Abort()
			return
		}

		resourceID, err := primitive.ObjectIDFromHex(resourceIDStr)
		if err != nil {
			response.ValidationError(c, "Invalid resource ID", nil)
			c.Abort()
			return
		}

		// Check if user owns the resource
		isOwner, err := permissionService.IsResourceOwner(c.Request.Context(), userID, resourceType, resourceID)
		if err != nil {
			response.Error(c, 500, "Failed to check resource ownership", err)
			c.Abort()
			return
		}

		if !isOwner {
			// Check if user is an admin (has permission to access any resource)
			isAdmin, err := permissionService.HasPermission(c.Request.Context(), userID, "admin")
			if err != nil || !isAdmin {
				response.ForbiddenError(c, "You don't have permission to access this resource")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RoleCheck middleware to check if the user has a specific role
func RoleCheck(permissionService *permission.Service, requiredRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by Auth middleware)
		userIDValue, exists := c.Get("userID")
		if !exists {
			response.UnauthorizedError(c, "User not authenticated")
			c.Abort()
			return
		}

		userID := userIDValue.(primitive.ObjectID)

		// Check if user has any of the required roles
		hasRole, err := permissionService.HasAnyRole(c.Request.Context(), userID, requiredRoles)
		if err != nil {
			response.Error(c, 500, "Failed to check user role", err)
			c.Abort()
			return
		}

		if !hasRole {
			response.ForbiddenError(c, "Insufficient privileges")
			c.Abort()
			return
		}

		c.Next()
	}
}
