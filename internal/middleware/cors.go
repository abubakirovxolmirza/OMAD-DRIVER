package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSMiddleware configures CORS access based on allowed origins list.
func CORSMiddleware(allowedOrigins string) fiber.Handler {
	originList := parseAllowedOrigins(allowedOrigins)
	allowOrigins := strings.Join(originList, ",")
	allowAll := false
	for _, origin := range originList {
		if origin == "*" {
			allowAll = true
			allowOrigins = "*"
			break
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowHeaders:     "Content-Type, Content-Length, Accept-Encoding, Authorization, Cache-Control, X-Requested-With",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		ExposeHeaders:    "Content-Disposition",
		AllowCredentials: !allowAll,
		AllowOriginsFunc: func(origin string) bool {
			if allowAll || origin == "" {
				return true
			}
			for _, allowed := range originList {
				if allowed == "*" {
					return true
				}
				if strings.Contains(allowed, "*") {
					if matchWildcard(strings.ToLower(allowed), strings.ToLower(origin)) {
						return true
					}
					continue
				}
				if strings.EqualFold(allowed, origin) {
					return true
				}
			}
			return false
		},
	})
}

func parseAllowedOrigins(origins string) []string {
	if origins == "" {
		return []string{"*"}
	}
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return []string{"*"}
	}
	return result
}

func matchWildcard(pattern, origin string) bool {
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == origin
	}
	if !strings.HasPrefix(origin, parts[0]) {
		return false
	}
	origin = origin[len(parts[0]):]

	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(origin, parts[i])
		if idx == -1 {
			return false
		}
		origin = origin[idx+len(parts[i]):]
	}

	last := parts[len(parts)-1]
	if last == "" {
		return true
	}
	return strings.HasSuffix(origin, last)
}
