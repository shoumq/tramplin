package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/authjwt"
	"tramplin/internal/dto"
	publicservice "tramplin/internal/service/public"
)

type PublicHandler struct {
	service *publicservice.Service
	jwt     *authjwt.Manager
}

// ListOpportunities godoc
// @Summary Список публичных возможностей
// @Tags public
// @Produce json
// @Param tag query string false "Фильтр по тегу"
// @Param work_format query string false "Фильтр по формату работы"
// @Param type query string false "Фильтр по типу возможности"
// @Param search query string false "Поисковый запрос"
// @Param salary_from query number false "Минимальная зарплата"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/opportunities [get]
func (h *PublicHandler) ListOpportunities(c *fiber.Ctx) error {
	data, err := h.service.ListOpportunities(c.Queries())
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListOpportunityMarkers godoc
// @Summary Маркеры возможностей для карты
// @Tags public
// @Produce json
// @Param tag query string false "Фильтр по тегу"
// @Param work_format query string false "Фильтр по формату работы"
// @Param type query string false "Фильтр по типу возможности"
// @Param search query string false "Поисковый запрос"
// @Param salary_from query number false "Минимальная зарплата"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/opportunities/map [get]
func (h *PublicHandler) ListOpportunityMarkers(c *fiber.Ctx) error {
	data, err := h.service.ListOpportunityMarkers(c.Queries())
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// GetOpportunity godoc
// @Summary Получить возможность по ID
// @Tags public
// @Produce json
// @Param id path string true "ID возможности"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/opportunities/{id} [get]
func (h *PublicHandler) GetOpportunity(c *fiber.Ctx) error {
	data, err := h.service.GetOpportunity(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusNotFound, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// CreateApplication godoc
// @Summary Откликнуться на возможность
// @Tags public
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID возможности"
// @Param payload body dto.ApplicationInput true "Данные отклика"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/opportunities/{id}/applications [post]
func (h *PublicHandler) CreateApplication(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.ApplicationInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateApplication(userID, c.Params("id"), input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

// ListCompanies godoc
// @Summary Список компаний
// @Tags public
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/companies [get]
func (h *PublicHandler) ListCompanies(c *fiber.Ctx) error {
	data, err := h.service.ListCompanies()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListStudents godoc
// @Summary Список публичных профилей студентов
// @Tags public
// @Produce json
// @Param search query string false "Поиск по имени, университету, факультету, специализации"
// @Param university_name query string false "Фильтр по университету"
// @Param faculty query string false "Фильтр по факультету"
// @Param specialization query string false "Фильтр по специализации"
// @Param study_year query int false "Фильтр по курсу"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/students [get]
func (h *PublicHandler) ListStudents(c *fiber.Ctx) error {
	data, err := h.service.ListStudents(c.Queries(), optionalUserID(c, h.jwt))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// GetCompany godoc
// @Summary Получить компанию по ID
// @Tags public
// @Produce json
// @Param id path string true "ID компании"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/companies/{id} [get]
func (h *PublicHandler) GetCompany(c *fiber.Ctx) error {
	data, err := h.service.GetCompany(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusNotFound, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// GetStudentProfile godoc
// @Summary Получить публичный профиль студента по ID
// @Tags public
// @Produce json
// @Param id path string true "ID пользователя студента"
// @Success 200 {object} SuccessResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/students/{id} [get]
func (h *PublicHandler) GetStudentProfile(c *fiber.Ctx) error {
	data, err := h.service.GetStudentProfile(c.Params("id"), optionalUserID(c, h.jwt))
	if err != nil {
		status := fiber.StatusNotFound
		if err.Error() == "student profile is available only to authorized users" || err.Error() == "student profile is available only to contacts" {
			status = fiber.StatusForbidden
		}
		return fail(c, status, err)
	}
	return respond(c, fiber.StatusOK, data)
}

func (h *PublicHandler) GetUserPresence(c *fiber.Ctx) error {
	data, err := h.service.GetUserPresence(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusNotFound, err)
	}
	return respond(c, fiber.StatusOK, data)
}

func (h *PublicHandler) GetCompanyPresence(c *fiber.Ctx) error {
	data, err := h.service.GetCompanyPresence(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusNotFound, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListTags godoc
// @Summary Список тегов
// @Tags public
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/tags [get]
func (h *PublicHandler) ListTags(c *fiber.Ctx) error {
	data, err := h.service.ListTags()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListCities godoc
// @Summary Список городов
// @Tags public
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/cities [get]
func (h *PublicHandler) ListCities(c *fiber.Ctx) error {
	data, err := h.service.ListCities()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListLocations godoc
// @Summary Список локаций
// @Tags public
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/locations [get]
func (h *PublicHandler) ListLocations(c *fiber.Ctx) error {
	data, err := h.service.ListLocations()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
