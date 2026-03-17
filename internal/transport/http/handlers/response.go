package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"tramplin/internal/service"
)

type Container struct {
	Account  *AccountHandler
	Auth     *AuthHandler
	Public   *PublicHandler
	Student  *StudentHandler
	Employer *EmployerHandler
	Curator  *CuratorHandler
}

func New(services *service.Services) *Container {
	return &Container{
		Account:  &AccountHandler{service: services.Account},
		Auth:     &AuthHandler{service: services.Auth},
		Public:   &PublicHandler{service: services.Public},
		Student:  &StudentHandler{service: services.Student},
		Employer: &EmployerHandler{service: services.Employer},
		Curator:  &CuratorHandler{service: services.Curator},
	}
}

type envelope struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
	Error  string `json:"error,omitempty"`
}

func respond(c *fiber.Ctx, statusCode int, data any) error {
	return c.Status(statusCode).JSON(envelope{Status: "ok", Data: data})
}

func fail(c *fiber.Ctx, statusCode int, err error) error {
	return c.Status(statusCode).JSON(envelope{Status: "error", Error: err.Error()})
}

func parseBody(c *fiber.Ctx, dst any) error {
	if len(c.Body()) == 0 {
		return nil
	}
	return c.BodyParser(dst)
}

func requiredUserID(c *fiber.Ctx) (string, error) {
	userID, _ := c.Locals("user_id").(string)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "bearer token is required")
	}
	return userID, nil
}
