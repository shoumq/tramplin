package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/service"
)

type AuthHandler struct{ service *service.AuthService }

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input service.RegisterInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.Register(input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input service.LoginInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.Login(input, false)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	return respond(c, fiber.StatusOK, data)
}

func (h *AuthHandler) CuratorLogin(c *fiber.Ctx) error {
	var input service.LoginInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.Login(input, true)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	return respond(c, fiber.StatusOK, data)
}
