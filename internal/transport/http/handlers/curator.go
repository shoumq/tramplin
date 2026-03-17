package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/dto"
	curatorservice "tramplin/internal/service/curator"
)

type CuratorHandler struct{ service *curatorservice.Service }

// CreateUser godoc
// @Summary Создать куратора
// @Tags curator
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.CuratorCreateInput true "Данные куратора"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/curator/users [post]
func (h *CuratorHandler) CreateUser(c *fiber.Ctx) error {
	actorID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.CuratorCreateInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateCurator(actorID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

// UpdateUserStatus godoc
// @Summary Изменить статус пользователя
// @Tags curator
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID пользователя"
// @Param payload body dto.StatusPayload true "Данные статуса"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/curator/users/{id}/status [patch]
func (h *CuratorHandler) UpdateUserStatus(c *fiber.Ctx) error {
	actorID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	status := c.Query("status")
	if status == "" {
		var payload map[string]string
		_ = parseBody(c, &payload)
		status = payload["status"]
	}
	data, err := h.service.UpdateUserStatus(actorID, c.Params("id"), status)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// UpdateStudentProfile godoc
// @Summary Обновить профиль студента от имени куратора
// @Tags curator
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID пользователя-студента"
// @Param payload body dto.StudentProfileInput true "Данные профиля студента"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/curator/student-profiles/{id} [patch]
func (h *CuratorHandler) UpdateStudentProfile(c *fiber.Ctx) error {
	actorID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input dto.StudentProfileInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateStudentProfile(actorID, c.Params("id"), input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// UpdateEmployerProfile godoc
// @Summary Обновить профиль работодателя от имени куратора
// @Tags curator
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID пользователя-работодателя"
// @Param payload body dto.EmployerProfileInput true "Данные профиля работодателя"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/curator/employer-profiles/{id} [patch]
func (h *CuratorHandler) UpdateEmployerProfile(c *fiber.Ctx) error {
	actorID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var payload dto.EmployerProfileInput
	if err := parseBody(c, &payload); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateEmployerProfile(actorID, c.Params("id"), payload)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListModerationQueue godoc
// @Summary Список очереди модерации
// @Tags curator
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/curator/moderation-queue [get]
func (h *CuratorHandler) ListModerationQueue(c *fiber.Ctx) error {
	data, err := h.service.ListModerationQueue()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ReviewModerationQueueItem godoc
// @Summary Принять решение по элементу очереди модерации
// @Tags curator
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID элемента очереди модерации"
// @Param payload body dto.StatusPayload true "Данные решения"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/curator/moderation-queue/{id} [patch]
func (h *CuratorHandler) ReviewModerationQueueItem(c *fiber.Ctx) error {
	actorID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var payload map[string]string
	if err := parseBody(c, &payload); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.ReviewModerationQueue(actorID, c.Params("id"), payload["status"], payload["comment"])
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListCompanyVerifications godoc
// @Summary Список верификаций компаний
// @Tags curator
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/curator/company-verifications [get]
func (h *CuratorHandler) ListCompanyVerifications(c *fiber.Ctx) error {
	data, err := h.service.ListCompanyVerifications()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ReviewCompanyVerification godoc
// @Summary Принять решение по верификации компании
// @Tags curator
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID верификации"
// @Param payload body dto.StatusPayload true "Данные решения"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/curator/company-verifications/{id} [patch]
func (h *CuratorHandler) ReviewCompanyVerification(c *fiber.Ctx) error {
	actorID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var payload map[string]string
	if err := parseBody(c, &payload); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.ReviewCompanyVerification(actorID, c.Params("id"), payload["status"], payload["comment"])
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// UpdateOpportunityStatus godoc
// @Summary Изменить статус возможности
// @Tags curator
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID возможности"
// @Param payload body dto.StatusPayload true "Данные статуса"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/curator/opportunities/{id}/status [patch]
func (h *CuratorHandler) UpdateOpportunityStatus(c *fiber.Ctx) error {
	actorID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	status := c.Query("status")
	if status == "" {
		var payload map[string]string
		_ = parseBody(c, &payload)
		status = payload["status"]
	}
	data, err := h.service.UpdateOpportunityStatus(actorID, c.Params("id"), status)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

// ListAuditLogs godoc
// @Summary Список записей аудита
// @Tags curator
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/curator/audit-logs [get]
func (h *CuratorHandler) ListAuditLogs(c *fiber.Ctx) error {
	data, err := h.service.ListAuditLogs()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
