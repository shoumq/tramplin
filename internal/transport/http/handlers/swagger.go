package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/docs"
)

func SwaggerUI(c *fiber.Ctx) error {
	c.Type("html", "utf-8")
	return c.SendString(docs.SwaggerUIHTML())
}

func SwaggerJSON(c *fiber.Ctx) error {
	payload, err := docs.SpecJSON()
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, err)
	}
	c.Type("json", "utf-8")
	return c.Send(payload)
}
