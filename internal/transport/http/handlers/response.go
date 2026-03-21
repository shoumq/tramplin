package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"tramplin/internal/authjwt"
	"tramplin/internal/service"
)

type Container struct {
	Account  *AccountHandler
	Auth     *AuthHandler
	Chat     *ChatHandler
	Public   *PublicHandler
	Student  *StudentHandler
	Employer *EmployerHandler
	Curator  *CuratorHandler
}

func New(services *service.Services, jwtManager *authjwt.Manager) *Container {
	return &Container{
		Account:  &AccountHandler{service: services.Account},
		Auth:     &AuthHandler{service: services.Auth},
		Chat:     NewChatHandler(services.Chat, jwtManager),
		Public:   &PublicHandler{service: services.Public, jwt: jwtManager},
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

func optionalUserID(c *fiber.Ctx, manager *authjwt.Manager) string {
	if manager == nil {
		return ""
	}
	header := strings.TrimSpace(c.Get("Authorization"))
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	claims, err := manager.Parse(strings.TrimSpace(parts[1]))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(claims.UserID)
}
