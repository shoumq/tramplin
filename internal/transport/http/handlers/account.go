package handlers

import (
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"

	accountservice "tramplin/internal/service/account"
)

type AccountHandler struct{ service *accountservice.Service }

// GetMe godoc
// @Summary Получить текущего пользователя
// @Tags account
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/me [get]
func (h *AccountHandler) GetMe(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.GetMe(userID)
	if err != nil {
		return fail(c, fiber.StatusNotFound, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// UploadAvatar godoc
// @Summary Загрузить аватар пользователя
// @Tags account
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Файл аватара"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/avatar [put]
func (h *AccountHandler) UploadAvatar(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, fiber.NewError(fiber.StatusBadRequest, "file is required"))
	}
	if err := validateAvatarFile(fileHeader); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	defer file.Close()

	user, err := h.service.UploadAvatar(c.UserContext(), userID, fileHeader.Filename, fileHeader.Header.Get("Content-Type"), fileHeader.Size, file)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, user)
}

func validateAvatarFile(file *multipart.FileHeader) error {
	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	switch contentType {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
	default:
		return fiber.NewError(fiber.StatusBadRequest, "only jpeg, png, webp or gif avatars are supported")
	}
	if file.Size <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "avatar file is empty")
	}
	if file.Size > 5*1024*1024 {
		return fiber.NewError(fiber.StatusBadRequest, "avatar file is too large")
	}
	return nil
}
