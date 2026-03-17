package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"tramplin/internal/authjwt"
)

func RequireJWT(manager *authjwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := strings.TrimSpace(c.Get("Authorization"))
		if header == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization header is required")
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization header must use Bearer token")
		}

		claims, err := manager.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid bearer token")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("roles", claims.Roles)
		return c.Next()
	}
}
