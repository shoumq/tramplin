package repository

import "tramplin/internal/models"

const (
	RoleStudent  = "student"
	RoleEmployer = "employer"
	RoleCurator  = "curator"
	RoleAdmin    = "admin"
)

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

type StudentFilter struct {
	ViewerUserID   string
	Search         string
	UniversityName string
	Faculty        string
	Specialization string
	StudyYear      int
}

type PlatformRepository interface {
	RegisterUser(params RegisterUserParams) (*models.User, string, error)
	Login(email, password string) (*models.User, []string, error)
	GetUser(userID string) (*models.User, error)
	GetUserRoles(userID string) ([]string, error)
	CreateCurator(email, password, displayName, curatorType, createdBy string) (*models.User, error)
	UpdateUserStatus(userID, status, actorID string) (*models.User, error)
	UpdateUserAvatar(userID, avatarObject, avatarURL string) (*models.User, error)
	GetStudentProfile(userID string) (*models.StudentProfile, error)
	GetPublicStudentProfile(userID, viewerUserID string) (*models.PublicStudentProfile, error)
	ListPublicStudentProfiles(filter StudentFilter) ([]models.PublicStudentProfile, error)
	UpsertStudentProfile(profile models.StudentProfile, actorID string) (*models.StudentProfile, error)
	ListResumes(studentUserID string) ([]models.Resume, error)
	CreateResume(resume models.Resume) (*models.Resume, error)
	SetPrimaryResume(studentUserID, resumeID string) (*models.Resume, error)
	ListPortfolioProjects(studentUserID string) ([]models.PortfolioProject, error)
	CreatePortfolioProject(project models.PortfolioProject) (*models.PortfolioProject, error)
	ListStudentApplications(studentUserID string) ([]models.Application, error)
	ListFavoriteOpportunities(userID string) ([]models.PublicOpportunity, error)
	AddFavoriteOpportunity(userID, opportunityID string) error
	RemoveFavoriteOpportunity(userID, opportunityID string) error
	ListFavoriteCompanies(userID string) ([]models.Company, error)
	AddFavoriteCompany(userID, companyID string) error
	RemoveFavoriteCompany(userID, companyID string) error
	ListContacts(userID string) ([]models.User, error)
	ListContactRequests(userID string) ([]models.ContactRequest, error)
	CreateContactRequest(senderUserID, receiverUserID, message string) (*models.ContactRequest, error)
	UpdateContactRequestStatus(requestID, userID, status string) (*models.ContactRequest, error)
	CreateRecommendation(rec models.Recommendation) (*models.Recommendation, error)
	ListNotifications(userID string) ([]models.Notification, error)
	CreateChatConversation(userID, participantUserID, opportunityID string) (*models.ChatConversation, error)
	GetChatConversation(userID, conversationID string) (*models.ChatConversation, error)
	ListChatConversations(userID string) ([]models.ChatConversation, error)
	ListChatMessages(userID, conversationID string) ([]models.ChatMessage, error)
	CreateChatMessage(userID, conversationID, body string) (*models.ChatMessage, error)
	MarkChatMessagesRead(userID, conversationID string) (int64, error)
	TouchUserPresence(userID string, isOnline bool) error
	GetUserPresence(userID string) (*models.Presence, error)
	GetCompanyPresence(companyID string) (*models.Presence, error)
	ListOpportunities(filter OpportunityFilter) ([]models.PublicOpportunity, error)
	ListOpportunityMarkers(filter OpportunityFilter) ([]models.OpportunityMarker, error)
	GetOpportunity(id string) (*models.PublicOpportunity, error)
	CreateApplication(application models.Application) (*models.Application, error)
	ListCompanies() ([]models.Company, error)
	GetCompany(id string) (*models.Company, error)
	ListTags() ([]models.Tag, error)
	ListCities() ([]models.City, error)
	ListLocations() ([]models.Location, error)
	GetEmployerProfile(userID string) (*models.EmployerProfile, error)
	GetEmployerCompany(userID string) (*models.Company, error)
	UpdateEmployerCompany(userID string, update CompanyUpdate) (*models.Company, error)
	UpdateCompanyAvatar(userID, avatarObject, avatarURL string) (*models.Company, error)
	CreateCompanyLink(userID, linkType, url string) (*models.CompanyLink, error)
	SubmitCompanyVerification(userID, method, corporateEmail, inn, comment string) (*models.CompanyVerification, error)
	ListEmployerOpportunities(userID string) ([]models.Opportunity, error)
	CreateOpportunity(opportunity models.Opportunity) (*models.Opportunity, error)
	GetEmployerOpportunity(userID, opportunityID string) (*models.Opportunity, error)
	UpdateEmployerOpportunity(userID string, opportunity models.Opportunity) (*models.Opportunity, error)
	ListOpportunityApplications(userID, opportunityID string) ([]models.Application, error)
	UpdateApplicationStatus(userID, applicationID, status string) (*models.Application, error)
	UpdateEmployerProfile(userID string, profile models.EmployerProfile, actorID string) (*models.EmployerProfile, error)
	ListModerationQueue() ([]models.ModerationQueueItem, error)
	ReviewModerationQueueItem(itemID, curatorID, status, comment string) (*models.ModerationQueueItem, error)
	ListCompanyVerifications() ([]models.CompanyVerification, error)
	ReviewCompanyVerification(verificationID, curatorID, status, comment string) (*models.CompanyVerification, error)
	UpdateOpportunityStatus(curatorID, opportunityID, status string) (*models.Opportunity, error)
	ListAuditLogs() ([]models.AuditLog, error)
}
