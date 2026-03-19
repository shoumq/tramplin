package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/dto"
	employerservice "tramplin/internal/service/employer"
)

type EmployerHandler struct{ service *employerservice.Service }

// GetCompanyProfile godoc
// @Summary Получить профиль компании работодателя
// @Tags employer
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/company [get]
func (h *EmployerHandler) GetCompanyProfile(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.GetCompany(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// UpdateCompanyProfile godoc
// @Summary Обновить профиль компании работодателя
// @Tags employer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.CompanyInput true "Данные компании"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/company [put]
func (h *EmployerHandler) UpdateCompanyProfile(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.CompanyInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateCompany(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// UploadCompanyAvatar godoc
// @Summary Загрузить аватар компании
// @Tags employer
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Файл аватара компании"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/company/avatar [put]
func (h *EmployerHandler) UploadCompanyAvatar(c *fiber.Ctx) error {
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

	company, err := h.service.UploadCompanyAvatar(c.UserContext(), userID, fileHeader.Filename, fileHeader.Header.Get("Content-Type"), fileHeader.Size, file)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, company)
}

// CreateCompanyLink godoc
// @Summary Добавить ссылку компании
// @Tags employer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.CompanyLinkInput true "Данные ссылки компании"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/company-links [post]
func (h *EmployerHandler) CreateCompanyLink(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.CompanyLinkInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateCompanyLink(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

// SubmitCompanyVerification godoc
// @Summary Отправить компанию на верификацию
// @Tags employer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.VerificationInput true "Данные верификации"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/company-verifications [post]
func (h *EmployerHandler) SubmitCompanyVerification(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.VerificationInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.SubmitVerification(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

// ListOpportunities godoc
// @Summary Список возможностей работодателя
// @Tags employer
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/opportunities [get]
func (h *EmployerHandler) ListOpportunities(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListOpportunities(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// CreateOpportunity godoc
// @Summary Создать возможность
// @Tags employer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.OpportunityInput true "Данные возможности"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/opportunities [post]
func (h *EmployerHandler) CreateOpportunity(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.OpportunityInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateOpportunity(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

// GetOpportunity godoc
// @Summary Получить возможность работодателя
// @Tags employer
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID возможности"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/opportunities/{id} [get]
func (h *EmployerHandler) GetOpportunity(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.GetOpportunity(userID, c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// UpdateOpportunity godoc
// @Summary Обновить возможность работодателя
// @Tags employer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID возможности"
// @Param payload body dto.OpportunityInput true "Данные возможности"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/opportunities/{id} [patch]
func (h *EmployerHandler) UpdateOpportunity(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.OpportunityInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateOpportunity(userID, c.Params("id"), input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListOpportunityApplications godoc
// @Summary Список откликов на возможность
// @Tags employer
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID возможности"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/opportunities/{id}/applications [get]
func (h *EmployerHandler) ListOpportunityApplications(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListApplications(userID, c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// UpdateApplicationStatus godoc
// @Summary Изменить статус отклика
// @Tags employer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID отклика"
// @Param payload body dto.StatusPayload true "Данные статуса"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/employer/applications/{id}/status [patch]
func (h *EmployerHandler) UpdateApplicationStatus(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	status := c.Query("status")
	if status == "" {
		var payload map[string]string
		_ = parseBody(c, &payload)
		status = payload["status"]
	}
	data, err := h.service.UpdateApplicationStatus(userID, c.Params("id"), status)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
