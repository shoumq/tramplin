package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/service"
)

type EmployerHandler struct{ service *service.EmployerService }

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
func (h *EmployerHandler) UpdateCompanyProfile(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.CompanyInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateCompany(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
func (h *EmployerHandler) CreateCompanyLink(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.CompanyLinkInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateCompanyLink(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}
func (h *EmployerHandler) SubmitCompanyVerification(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.VerificationInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.SubmitVerification(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}
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
func (h *EmployerHandler) CreateOpportunity(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.OpportunityInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateOpportunity(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}
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
func (h *EmployerHandler) UpdateOpportunity(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.OpportunityInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateOpportunity(userID, c.Params("id"), input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
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
