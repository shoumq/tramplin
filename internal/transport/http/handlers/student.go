package handlers

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/service"
)

type StudentHandler struct{ service *service.StudentService }

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
func (h *StudentHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.StudentProfileInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateProfile(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
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
func (h *StudentHandler) CreateResume(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.ResumeInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateResume(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}
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
func (h *StudentHandler) CreatePortfolioProject(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.PortfolioProjectInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreatePortfolioProject(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}
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
func (h *StudentHandler) CreateContactRequest(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.ContactRequestInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateContactRequest(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}
func (h *StudentHandler) UpdateContactRequestStatus(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.ContactRequestInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.UpdateContactRequestStatus(userID, c.Params("id"), input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusOK, data)
}
func (h *StudentHandler) CreateRecommendation(c *fiber.Ctx) error {
	userID, err := requiredUserID(c)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, err)
	}
	var input service.RecommendationInput
	if err := parseBody(c, &input); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	data, err := h.service.CreateRecommendation(userID, input)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}
	return respond(c, fiber.StatusCreated, data)
}
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
