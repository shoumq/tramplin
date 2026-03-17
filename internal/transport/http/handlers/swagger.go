package handlers

import "github.com/gofiber/fiber/v2"

func SwaggerRedirect(c *fiber.Ctx) error {
	return c.Redirect("/swagger/index.html", fiber.StatusTemporaryRedirect)
}
