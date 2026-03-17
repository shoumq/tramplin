package repository

import "tramplin/internal/domain"

type RegisterUserParams struct {
	Email       string
	Password    string
	DisplayName string
	Role        string
	CompanyName string
}

type CompanyUpdate struct {
	LegalName   string
	BrandName   string
	Description string
	Industry    string
	WebsiteURL  string
	EmailDomain string
	INN         string
	OGRN        string
	CompanySize string
	FoundedYear int
	HQCityID    int64
}

type OpportunityFilter struct {
	Tag        string
	WorkFormat string
	Type       string
	Search     string
	SalaryFrom float64
}

type PlatformRepository interface {
	RegisterUser(params RegisterUserParams) (*domain.User, string, error)
	Login(email, password string) (*domain.User, []string, error)
	GetUser(userID string) (*domain.User, error)
	GetUserRoles(userID string) ([]string, error)
	CreateCurator(email, password, displayName, curatorType, createdBy string) (*domain.User, error)
	UpdateUserStatus(userID, status, actorID string) (*domain.User, error)
	GetStudentProfile(userID string) (*domain.StudentProfile, error)
	UpsertStudentProfile(profile domain.StudentProfile, actorID string) (*domain.StudentProfile, error)
	ListResumes(studentUserID string) ([]domain.Resume, error)
	CreateResume(resume domain.Resume) (*domain.Resume, error)
	SetPrimaryResume(studentUserID, resumeID string) (*domain.Resume, error)
	ListPortfolioProjects(studentUserID string) ([]domain.PortfolioProject, error)
	CreatePortfolioProject(project domain.PortfolioProject) (*domain.PortfolioProject, error)
	ListStudentApplications(studentUserID string) ([]domain.Application, error)
	ListFavoriteOpportunities(userID string) ([]domain.PublicOpportunity, error)
	AddFavoriteOpportunity(userID, opportunityID string) error
	RemoveFavoriteOpportunity(userID, opportunityID string) error
	ListFavoriteCompanies(userID string) ([]domain.Company, error)
	AddFavoriteCompany(userID, companyID string) error
	RemoveFavoriteCompany(userID, companyID string) error
	ListContacts(userID string) ([]domain.User, error)
	ListContactRequests(userID string) ([]domain.ContactRequest, error)
	CreateContactRequest(senderUserID, receiverUserID, message string) (*domain.ContactRequest, error)
	UpdateContactRequestStatus(requestID, userID, status string) (*domain.ContactRequest, error)
	CreateRecommendation(rec domain.Recommendation) (*domain.Recommendation, error)
	ListNotifications(userID string) ([]domain.Notification, error)
	ListOpportunities(filter OpportunityFilter) ([]domain.PublicOpportunity, error)
	ListOpportunityMarkers(filter OpportunityFilter) ([]domain.OpportunityMarker, error)
	GetOpportunity(id string) (*domain.PublicOpportunity, error)
	CreateApplication(application domain.Application) (*domain.Application, error)
	ListCompanies() ([]domain.Company, error)
	GetCompany(id string) (*domain.Company, error)
	ListTags() ([]domain.Tag, error)
	ListCities() ([]domain.City, error)
	ListLocations() ([]domain.Location, error)
	GetEmployerProfile(userID string) (*domain.EmployerProfile, error)
	GetEmployerCompany(userID string) (*domain.Company, error)
	UpdateEmployerCompany(userID string, update CompanyUpdate) (*domain.Company, error)
	CreateCompanyLink(userID, linkType, url string) (*domain.CompanyLink, error)
	SubmitCompanyVerification(userID, method, corporateEmail, inn, comment string) (*domain.CompanyVerification, error)
	ListEmployerOpportunities(userID string) ([]domain.Opportunity, error)
	CreateOpportunity(opportunity domain.Opportunity) (*domain.Opportunity, error)
	GetEmployerOpportunity(userID, opportunityID string) (*domain.Opportunity, error)
	UpdateEmployerOpportunity(userID string, opportunity domain.Opportunity) (*domain.Opportunity, error)
	ListOpportunityApplications(userID, opportunityID string) ([]domain.Application, error)
	UpdateApplicationStatus(userID, applicationID, status string) (*domain.Application, error)
	UpdateEmployerProfile(userID string, profile domain.EmployerProfile, actorID string) (*domain.EmployerProfile, error)
	ListModerationQueue() ([]domain.ModerationQueueItem, error)
	ReviewModerationQueueItem(itemID, curatorID, status, comment string) (*domain.ModerationQueueItem, error)
	ListCompanyVerifications() ([]domain.CompanyVerification, error)
	ReviewCompanyVerification(verificationID, curatorID, status, comment string) (*domain.CompanyVerification, error)
	UpdateOpportunityStatus(curatorID, opportunityID, status string) (*domain.Opportunity, error)
	ListAuditLogs() ([]domain.AuditLog, error)
}
