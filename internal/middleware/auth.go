package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"taxi-service/internal/models"
	"taxi-service/internal/utils"
)

// AuthMiddleware validates JWT token and attaches identity to the request context.
func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header required"})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization header format"})
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization header format"})
		}

		claims, err := utils.ValidateToken(token, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

// RoleMiddleware checks if user has required role.
func RoleMiddleware(allowedRoles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleInterface := c.Locals("user_role")
		if roleInterface == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User role not found"})
		}

		var userRole models.UserRole
		switch value := roleInterface.(type) {
		case models.UserRole:
			userRole = value
		case string:
			userRole = models.UserRole(value)
		default:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User role malformed"})
		}

		allowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
		}

		return c.Next()
	}
}

// GetUserID retrieves user ID from context.
func GetUserID(c *fiber.Ctx) (int64, bool) {
	userID := c.Locals("user_id")
	if userID == nil {
		return 0, false
	}
	switch value := userID.(type) {
	case int64:
		return value, true
	case int:
		return int64(value), true
	case float64:
		return int64(value), true
	default:
		return 0, false
	}
}

// GetUserRole retrieves user role from context.
func GetUserRole(c *fiber.Ctx) (models.UserRole, bool) {
	role := c.Locals("user_role")
	if role == nil {
		return "", false
	}
	switch value := role.(type) {
	case models.UserRole:
		return value, true
	case string:
		return models.UserRole(value), true
	default:
		return "", false
	}
}
