package http

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/transport/http/handlers"
)

func RegisterRoutes(app *fiber.App, h *handlers.Container) {
	app.Get("/swagger", handlers.SwaggerUI)
	app.Get("/swagger/", handlers.SwaggerUI)
	app.Get("/swagger/doc.json", handlers.SwaggerJSON)

	api := app.Group("/api")

	api.Get("/health", handlers.Health)

	auth := api.Group("/auth")
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.Login)
	auth.Post("/curator/login", h.Auth.CuratorLogin)

	api.Get("/opportunities", h.Public.ListOpportunities)
	api.Get("/opportunities/map", h.Public.ListOpportunityMarkers)
	api.Get("/opportunities/:id", h.Public.GetOpportunity)
	api.Post("/opportunities/:id/applications", h.Public.CreateApplication)
	api.Get("/companies", h.Public.ListCompanies)
	api.Get("/companies/:id", h.Public.GetCompany)
	api.Get("/tags", h.Public.ListTags)
	api.Get("/cities", h.Public.ListCities)
	api.Get("/locations", h.Public.ListLocations)

	student := api.Group("/me")
	student.Get("/student-profile", h.Student.GetProfile)
	student.Put("/student-profile", h.Student.UpdateProfile)
	student.Get("/resumes", h.Student.ListResumes)
	student.Post("/resumes", h.Student.CreateResume)
	student.Patch("/resumes/:id/primary", h.Student.SetPrimaryResume)
	student.Get("/portfolio-projects", h.Student.ListPortfolioProjects)
	student.Post("/portfolio-projects", h.Student.CreatePortfolioProject)
	student.Get("/applications", h.Student.ListApplications)
	student.Get("/favorite-opportunities", h.Student.ListFavoriteOpportunities)
	student.Post("/favorite-opportunities/:opportunityId", h.Student.AddFavoriteOpportunity)
	student.Delete("/favorite-opportunities/:opportunityId", h.Student.RemoveFavoriteOpportunity)
	student.Get("/favorite-companies", h.Student.ListFavoriteCompanies)
	student.Post("/favorite-companies/:companyId", h.Student.AddFavoriteCompany)
	student.Delete("/favorite-companies/:companyId", h.Student.RemoveFavoriteCompany)
	student.Get("/contacts", h.Student.ListContacts)
	student.Get("/contact-requests", h.Student.ListContactRequests)
	student.Post("/contact-requests", h.Student.CreateContactRequest)
	student.Patch("/contact-requests/:id", h.Student.UpdateContactRequestStatus)
	student.Post("/recommendations", h.Student.CreateRecommendation)
	student.Get("/notifications", h.Student.ListNotifications)

	employer := api.Group("/employer")
	employer.Get("/company", h.Employer.GetCompanyProfile)
	employer.Put("/company", h.Employer.UpdateCompanyProfile)
	employer.Post("/company-links", h.Employer.CreateCompanyLink)
	employer.Post("/company-verifications", h.Employer.SubmitCompanyVerification)
	employer.Get("/opportunities", h.Employer.ListOpportunities)
	employer.Post("/opportunities", h.Employer.CreateOpportunity)
	employer.Get("/opportunities/:id", h.Employer.GetOpportunity)
	employer.Patch("/opportunities/:id", h.Employer.UpdateOpportunity)
	employer.Get("/opportunities/:id/applications", h.Employer.ListOpportunityApplications)
	employer.Patch("/applications/:id/status", h.Employer.UpdateApplicationStatus)

	curator := api.Group("/curator")
	curator.Post("/users", h.Curator.CreateUser)
	curator.Patch("/users/:id/status", h.Curator.UpdateUserStatus)
	curator.Patch("/student-profiles/:id", h.Curator.UpdateStudentProfile)
	curator.Patch("/employer-profiles/:id", h.Curator.UpdateEmployerProfile)
	curator.Get("/moderation-queue", h.Curator.ListModerationQueue)
	curator.Patch("/moderation-queue/:id", h.Curator.ReviewModerationQueueItem)
	curator.Get("/company-verifications", h.Curator.ListCompanyVerifications)
	curator.Patch("/company-verifications/:id", h.Curator.ReviewCompanyVerification)
	curator.Patch("/opportunities/:id/status", h.Curator.UpdateOpportunityStatus)
	curator.Get("/audit-logs", h.Curator.ListAuditLogs)
}
