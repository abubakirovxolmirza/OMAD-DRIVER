package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"taxi-service/internal/models"
	"taxi-service/internal/utils"
)

// AuthMiddleware validates JWT token (Gin version)
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// AuthMiddlewareFiber validates JWT token (Fiber version)
func AuthMiddlewareFiber(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.H{"error": "Authorization header required"})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.H{"error": "Invalid authorization header format"})
		}

		token := parts[1]
		claims, err := utils.ValidateToken(token, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.H{"error": "Invalid or expired token"})
		}

		// Set user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

// RoleMiddleware checks if user has required role (Gin version)
func RoleMiddleware(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleInterface, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		userRole := roleInterface.(models.UserRole)

		// Check if user role is in allowed roles
		allowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleMiddlewareFiber checks if user has required role (Fiber version)
func RoleMiddlewareFiber(allowedRoles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleInterface := c.Locals("user_role")
		if roleInterface == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.H{"error": "User role not found"})
		}

		userRole := roleInterface.(models.UserRole)

		// Check if user role is in allowed roles
		allowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.H{"error": "Insufficient permissions"})
		}

		return c.Next()
	}
}

// GetUserID retrieves user ID from context (Gin)
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(int64), true
}

// GetUserIDFiber retrieves user ID from context (Fiber)
func GetUserIDFiber(c *fiber.Ctx) (int64, bool) {
	userID := c.Locals("user_id")
	if userID == nil {
		return 0, false
	}
	return userID.(int64), true
}

// GetUserRole retrieves user role from context (Gin)
func GetUserRole(c *gin.Context) (models.UserRole, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	return role.(models.UserRole), true
}

// GetUserRoleFiber retrieves user role from context (Fiber)
func GetUserRoleFiber(c *fiber.Ctx) (models.UserRole, bool) {
	role := c.Locals("user_role")
	if role == nil {
		return "", false
	}
	return role.(models.UserRole), true
}
