package http

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"tramplin/internal/authjwt"
	"tramplin/internal/transport/http/handlers"
	"tramplin/internal/transport/http/middleware"
)

func RegisterRoutes(app *fiber.App, h *handlers.Container, jwtManager *authjwt.Manager) {
	app.Get("/docs", handlers.SwaggerRedirect)
	app.Get("/docs/", handlers.SwaggerRedirect)
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	api := app.Group("/api")

	api.Get("/health", handlers.Health)

	auth := api.Group("/auth")
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.Login)
	auth.Post("/curator/login", h.Auth.CuratorLogin)
	api.Get("/ws/chat", h.Chat.WebSocket)

	api.Get("/opportunities", h.Public.ListOpportunities)
	api.Get("/opportunities/map", h.Public.ListOpportunityMarkers)
	api.Get("/opportunities/:id", h.Public.GetOpportunity)
	api.Get("/companies", h.Public.ListCompanies)
	api.Get("/companies/:id", h.Public.GetCompany)
	api.Get("/students/:id", h.Public.GetStudentProfile)
	api.Get("/companies/:id/presence", h.Public.GetCompanyPresence)
	api.Get("/users/:id/presence", h.Public.GetUserPresence)
	api.Get("/tags", h.Public.ListTags)
	api.Get("/cities", h.Public.ListCities)
	api.Get("/locations", h.Public.ListLocations)
	api.Post("/opportunities/:id/applications", middleware.RequireJWT(jwtManager), h.Public.CreateApplication)

	api.Get("/me", middleware.RequireJWT(jwtManager), h.Account.GetMe)
	student := api.Group("/me", middleware.RequireJWT(jwtManager))
	student.Put("/avatar", h.Account.UploadAvatar)
	student.Post("/presence", h.Account.TouchPresence)
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
	student.Get("/chats", h.Chat.ListConversations)
	student.Post("/chats", h.Chat.CreateConversation)
	student.Get("/chats/:id/messages", h.Chat.ListMessages)
	student.Post("/chats/:id/messages", h.Chat.CreateMessage)
	student.Post("/chats/:id/read", h.Chat.MarkRead)

	employer := api.Group("/employer", middleware.RequireJWT(jwtManager))
	employer.Get("/company", h.Employer.GetCompanyProfile)
	employer.Put("/company", h.Employer.UpdateCompanyProfile)
	employer.Put("/company/avatar", h.Employer.UploadCompanyAvatar)
	employer.Post("/company-links", h.Employer.CreateCompanyLink)
	employer.Post("/company-verifications", h.Employer.SubmitCompanyVerification)
	employer.Get("/opportunities", h.Employer.ListOpportunities)
	employer.Post("/opportunities", h.Employer.CreateOpportunity)
	employer.Get("/opportunities/:id", h.Employer.GetOpportunity)
	employer.Patch("/opportunities/:id", h.Employer.UpdateOpportunity)
	employer.Get("/opportunities/:id/applications", h.Employer.ListOpportunityApplications)
	employer.Patch("/applications/:id/status", h.Employer.UpdateApplicationStatus)

	curator := api.Group("/curator", middleware.RequireJWT(jwtManager))
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
