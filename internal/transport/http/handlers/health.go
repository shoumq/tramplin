package handlers

import "github.com/gofiber/fiber/v2"

// Health godoc
// @Summary Проверка состояния сервиса
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/health [get]
func Health(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "tramplin",
	})
}
