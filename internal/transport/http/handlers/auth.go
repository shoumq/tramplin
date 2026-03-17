package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/dto"
	authservice "tramplin/internal/service/auth"
)

type AuthHandler struct{ service *authservice.Service }

// Register godoc
// @Summary Регистрация пользователя
// @Description Самостоятельная регистрация доступна только для ролей `student` и `employer`. Поле `company_name` необязательно и используется только для `employer`.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body dto.RegisterInput true "Данные регистрации"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input dto.RegisterInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.Register(input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

// Login godoc
// @Summary Вход пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body dto.LoginInput true "Данные для входа"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input dto.LoginInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.Login(input, false)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// CuratorLogin godoc
// @Summary Вход куратора
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body dto.LoginInput true "Данные для входа куратора"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/curator/login [post]
func (h *AuthHandler) CuratorLogin(c *fiber.Ctx) error {
	var input dto.LoginInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.Login(input, true)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	return respond(c, fiber.StatusOK, data)
}
