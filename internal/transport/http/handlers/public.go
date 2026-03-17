package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/service"
)

type PublicHandler struct{ service *service.PublicService }

func (h *PublicHandler) ListOpportunities(c *fiber.Ctx) error {
	data, err := h.service.ListOpportunities(c.Queries())
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

func (h *PublicHandler) ListOpportunityMarkers(c *fiber.Ctx) error {
	data, err := h.service.ListOpportunityMarkers(c.Queries())
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}

func (h *PublicHandler) GetOpportunity(c *fiber.Ctx) error {
	data, err := h.service.GetOpportunity(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusNotFound, err)
	}
	return respond(c, fiber.StatusOK, data)
}

func (h *PublicHandler) CreateApplication(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.ApplicationInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateApplication(userID, c.Params("id"), input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}

func (h *PublicHandler) ListCompanies(c *fiber.Ctx) error {
	data, err := h.service.ListCompanies()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
func (h *PublicHandler) GetCompany(c *fiber.Ctx) error {
	data, err := h.service.GetCompany(c.Params("id"))
	if err != nil {
		return fail(c, fiber.StatusNotFound, err)
	}
	return respond(c, fiber.StatusOK, data)
}
func (h *PublicHandler) ListTags(c *fiber.Ctx) error {
	data, err := h.service.ListTags()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
func (h *PublicHandler) ListCities(c *fiber.Ctx) error {
	data, err := h.service.ListCities()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
func (h *PublicHandler) ListLocations(c *fiber.Ctx) error {
	data, err := h.service.ListLocations()
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
