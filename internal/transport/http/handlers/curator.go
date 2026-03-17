package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/service"
)

type CuratorHandler struct{ service *service.CuratorService }

func (h *CuratorHandler) CreateUser(c *fiber.Ctx) error {
	actorID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.CuratorCreateInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateCurator(actorID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}
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
func (h *CuratorHandler) UpdateStudentProfile(c *fiber.Ctx) error {
	actorID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.StudentProfileInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateStudentProfile(actorID, c.Params("id"), input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
func (h *CuratorHandler) UpdateEmployerProfile(c *fiber.Ctx) error {
	actorID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var payload struct {
		PositionTitle          string `json:"position_title"`
		IsCompanyOwner         bool   `json:"is_company_owner"`
		CanCreateOpportunities bool   `json:"can_create_opportunities"`
		CanEditCompanyProfile  bool   `json:"can_edit_company_profile"`
	}
	if err := parseBody(c, &payload); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateEmployerProfile(actorID, c.Params("id"), payload.CanCreateOpportunities, payload.CanEditCompanyProfile, payload.IsCompanyOwner, payload.PositionTitle)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
func (h *CuratorHandler) ListModerationQueue(c *fiber.Ctx) error {
	data, err := h.service.ListModerationQueue()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
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
func (h *CuratorHandler) ListCompanyVerifications(c *fiber.Ctx) error {
	data, err := h.service.ListCompanyVerifications()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
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
func (h *CuratorHandler) ListAuditLogs(c *fiber.Ctx) error {
	data, err := h.service.ListAuditLogs()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
