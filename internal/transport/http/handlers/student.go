package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/dto"
	studentservice "tramplin/internal/service/student"
)

type StudentHandler struct{ service *studentservice.Service }

// GetProfile godoc
// @Summary Получить профиль студента
// @Tags student
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/me/student-profile [get]
func (h *StudentHandler) GetProfile(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.GetProfile(userID)
	if err != nil {
		return fail(c, fiber.StatusNotFound, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// UpdateProfile godoc
// @Summary Обновить профиль студента
// @Tags student
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.StudentProfileInput true "Данные профиля студента"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/student-profile [put]
func (h *StudentHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.StudentProfileInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateProfile(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListResumes godoc
// @Summary Список резюме
// @Tags student
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/resumes [get]
func (h *StudentHandler) ListResumes(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListResumes(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// CreateResume godoc
// @Summary Создать резюме
// @Tags student
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.ResumeInput true "Данные резюме"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/resumes [post]
func (h *StudentHandler) CreateResume(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.ResumeInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateResume(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

// SetPrimaryResume godoc
// @Summary Сделать резюме основным
// @Tags student
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID резюме"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/resumes/{id}/primary [patch]
func (h *StudentHandler) SetPrimaryResume(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.SetPrimaryResume(userID, c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListPortfolioProjects godoc
// @Summary Список проектов портфолио
// @Tags student
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/portfolio-projects [get]
func (h *StudentHandler) ListPortfolioProjects(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListPortfolioProjects(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// CreatePortfolioProject godoc
// @Summary Создать проект портфолио
// @Tags student
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.PortfolioProjectInput true "Данные проекта портфолио"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/portfolio-projects [post]
func (h *StudentHandler) CreatePortfolioProject(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.PortfolioProjectInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreatePortfolioProject(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

// ListApplications godoc
// @Summary Список моих откликов
// @Tags student
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/applications [get]
func (h *StudentHandler) ListApplications(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListApplications(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListFavoriteOpportunities godoc
// @Summary Список избранных возможностей
// @Tags student
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/favorite-opportunities [get]
func (h *StudentHandler) ListFavoriteOpportunities(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListFavoriteOpportunities(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// AddFavoriteOpportunity godoc
// @Summary Добавить возможность в избранное
// @Tags student
// @Produce json
// @Security BearerAuth
// @Param opportunityId path string true "ID возможности"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/favorite-opportunities/{opportunityId} [post]
func (h *StudentHandler) AddFavoriteOpportunity(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	if err := h.service.AddFavoriteOpportunity(userID, c.Params("opportunityId")); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, fiber.Map{"saved": true})
}

// RemoveFavoriteOpportunity godoc
// @Summary Удалить возможность из избранного
// @Tags student
// @Produce json
// @Security BearerAuth
// @Param opportunityId path string true "ID возможности"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/favorite-opportunities/{opportunityId} [delete]
func (h *StudentHandler) RemoveFavoriteOpportunity(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	if err := h.service.RemoveFavoriteOpportunity(userID, c.Params("opportunityId")); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, fiber.Map{"removed": true})
}

// ListFavoriteCompanies godoc
// @Summary Список избранных компаний
// @Tags student
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/favorite-companies [get]
func (h *StudentHandler) ListFavoriteCompanies(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListFavoriteCompanies(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// AddFavoriteCompany godoc
// @Summary Добавить компанию в избранное
// @Tags student
// @Produce json
// @Security BearerAuth
// @Param companyId path string true "ID компании"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/favorite-companies/{companyId} [post]
func (h *StudentHandler) AddFavoriteCompany(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	if err := h.service.AddFavoriteCompany(userID, c.Params("companyId")); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, fiber.Map{"saved": true})
}

// RemoveFavoriteCompany godoc
// @Summary Удалить компанию из избранного
// @Tags student
// @Produce json
// @Security BearerAuth
// @Param companyId path string true "ID компании"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/favorite-companies/{companyId} [delete]
func (h *StudentHandler) RemoveFavoriteCompany(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	if err := h.service.RemoveFavoriteCompany(userID, c.Params("companyId")); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, fiber.Map{"removed": true})
}

// ListContacts godoc
// @Summary Список контактов
// @Tags student
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/contacts [get]
func (h *StudentHandler) ListContacts(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListContacts(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListContactRequests godoc
// @Summary Список запросов в контакты
// @Tags student
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/contact-requests [get]
func (h *StudentHandler) ListContactRequests(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListContactRequests(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// CreateContactRequest godoc
// @Summary Создать запрос в контакты
// @Tags student
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.ContactRequestInput true "Данные запроса в контакты"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/contact-requests [post]
func (h *StudentHandler) CreateContactRequest(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.ContactRequestInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateContactRequest(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

// UpdateContactRequestStatus godoc
// @Summary Изменить статус запроса в контакты
// @Tags student
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID запроса в контакты"
// @Param payload body dto.ContactRequestInput true "Данные запроса в контакты"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/contact-requests/{id} [patch]
func (h *StudentHandler) UpdateContactRequestStatus(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.ContactRequestInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateContactRequestStatus(userID, c.Params("id"), input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// CreateRecommendation godoc
// @Summary Создать рекомендацию
// @Tags student
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.RecommendationInput true "Данные рекомендации"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/recommendations [post]
func (h *StudentHandler) CreateRecommendation(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.RecommendationInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateRecommendation(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

// ListNotifications godoc
// @Summary Список уведомлений
// @Tags student
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/me/notifications [get]
func (h *StudentHandler) ListNotifications(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	data, err := h.service.ListNotifications(userID)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
