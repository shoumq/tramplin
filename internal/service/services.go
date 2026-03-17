package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"tramplin/internal/domain"
	"tramplin/internal/repository"
)

type Services struct {
	Auth     *AuthService
	Public   *PublicService
	Student  *StudentService
	Employer *EmployerService
	Curator  *CuratorService
}

func New(repo repository.PlatformRepository) *Services {
	return &Services{
		Auth:     &AuthService{repo: repo},
		Public:   &PublicService{repo: repo},
		Student:  &StudentService{repo: repo},
		Employer: &EmployerService{repo: repo},
		Curator:  &CuratorService{repo: repo},
	}
}

type AuthService struct{ repo repository.PlatformRepository }
type PublicService struct{ repo repository.PlatformRepository }
type StudentService struct{ repo repository.PlatformRepository }
type EmployerService struct{ repo repository.PlatformRepository }
type CuratorService struct{ repo repository.PlatformRepository }

type RegisterInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	CompanyName string `json:"company_name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CuratorCreateInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	CuratorType string `json:"curator_type"`
}

type CompanyInput struct {
	LegalName   string `json:"legal_name"`
	BrandName   string `json:"brand_name"`
	Description string `json:"description"`
	Industry    string `json:"industry"`
	WebsiteURL  string `json:"website_url"`
	EmailDomain string `json:"email_domain"`
	INN         string `json:"inn"`
	OGRN        string `json:"ogrn"`
	CompanySize string `json:"company_size"`
	FoundedYear int    `json:"founded_year"`
	HQCityID    int64  `json:"hq_city_id"`
}

type CompanyLinkInput struct {
	LinkType string `json:"link_type"`
	URL      string `json:"url"`
}

type VerificationInput struct {
	VerificationMethod string `json:"verification_method"`
	CorporateEmail     string `json:"corporate_email"`
	INNSubmitted       string `json:"inn_submitted"`
	DocumentsComment   string `json:"documents_comment"`
}

type OpportunityInput struct {
	Title               string   `json:"title"`
	ShortDescription    string   `json:"short_description"`
	FullDescription     string   `json:"full_description"`
	OpportunityType     string   `json:"opportunity_type"`
	VacancyLevel        string   `json:"vacancy_level"`
	EmploymentType      string   `json:"employment_type"`
	WorkFormat          string   `json:"work_format"`
	LocationID          string   `json:"location_id"`
	SalaryMin           float64  `json:"salary_min"`
	SalaryMax           float64  `json:"salary_max"`
	SalaryCurrency      string   `json:"salary_currency"`
	IsSalaryVisible     bool     `json:"is_salary_visible"`
	ContactsInfo        string   `json:"contacts_info"`
	ExternalURL         string   `json:"external_url"`
	Status              string   `json:"status"`
	TagIDs              []string `json:"tag_ids"`
	ApplicationDeadline string   `json:"application_deadline"`
	EventStartAt        string   `json:"event_start_at"`
	EventEndAt          string   `json:"event_end_at"`
	ExpiresAt           string   `json:"expires_at"`
}

type StudentProfileInput struct {
	LastName            string `json:"last_name"`
	FirstName           string `json:"first_name"`
	MiddleName          string `json:"middle_name"`
	UniversityName      string `json:"university_name"`
	Faculty             string `json:"faculty"`
	Specialization      string `json:"specialization"`
	StudyYear           int    `json:"study_year"`
	GraduationYear      int    `json:"graduation_year"`
	About               string `json:"about"`
	ProfileVisibility   string `json:"profile_visibility"`
	ShowResume          bool   `json:"show_resume"`
	ShowApplications    bool   `json:"show_applications"`
	ShowCareerInterests bool   `json:"show_career_interests"`
	Telegram            string `json:"telegram"`
	GithubURL           string `json:"github_url"`
	LinkedinURL         string `json:"linkedin_url"`
	WebsiteURL          string `json:"website_url"`
	CityID              int64  `json:"city_id"`
}

type ResumeInput struct {
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	ExperienceText string `json:"experience_text"`
	EducationText  string `json:"education_text"`
}

type PortfolioProjectInput struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	ProjectURL    string `json:"project_url"`
	RepositoryURL string `json:"repository_url"`
	DemoURL       string `json:"demo_url"`
	StartedAt     string `json:"started_at"`
	FinishedAt    string `json:"finished_at"`
}

type ApplicationInput struct {
	ResumeID    string `json:"resume_id"`
	CoverLetter string `json:"cover_letter"`
}

type ContactRequestInput struct {
	ReceiverUserID string `json:"receiver_user_id"`
	Message        string `json:"message"`
	Status         string `json:"status"`
}

type RecommendationInput struct {
	ToUserID      string `json:"to_user_id"`
	OpportunityID string `json:"opportunity_id"`
	Message       string `json:"message"`
}

func (s *AuthService) Register(input RegisterInput) (map[string]any, error) {
	user, role, err := s.repo.RegisterUser(repository.RegisterUserParams{Email: input.Email, Password: input.Password, DisplayName: input.DisplayName, Role: strings.ToLower(input.Role), CompanyName: input.CompanyName})
	if err != nil {
		return nil, err
	}
	return map[string]any{"user": user, "role": role}, nil
}

func (s *AuthService) Login(input LoginInput, curatorOnly bool) (map[string]any, error) {
	user, roles, err := s.repo.Login(input.Email, input.Password)
	if err != nil {
		return nil, err
	}
	if curatorOnly {
		allowed := false
		for _, role := range roles {
			if role == "curator" || role == "admin" {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, errors.New("curator access required")
		}
	}
	return map[string]any{"user": user, "roles": roles}, nil
}

func buildFilter(params map[string]string) repository.OpportunityFilter {
	salary, _ := strconv.ParseFloat(params["salary_from"], 64)
	return repository.OpportunityFilter{Tag: params["tag"], WorkFormat: params["work_format"], Type: params["type"], Search: params["search"], SalaryFrom: salary}
}

func (s *PublicService) ListOpportunities(params map[string]string) ([]domain.PublicOpportunity, error) {
	return s.repo.ListOpportunities(buildFilter(params))
}
func (s *PublicService) ListOpportunityMarkers(params map[string]string) ([]domain.OpportunityMarker, error) {
	return s.repo.ListOpportunityMarkers(buildFilter(params))
}
func (s *PublicService) GetOpportunity(id string) (*domain.PublicOpportunity, error) {
	return s.repo.GetOpportunity(id)
}
func (s *PublicService) ListCompanies() ([]domain.Company, error)      { return s.repo.ListCompanies() }
func (s *PublicService) GetCompany(id string) (*domain.Company, error) { return s.repo.GetCompany(id) }
func (s *PublicService) ListTags() ([]domain.Tag, error)               { return s.repo.ListTags() }
func (s *PublicService) ListCities() ([]domain.City, error)            { return s.repo.ListCities() }
func (s *PublicService) ListLocations() ([]domain.Location, error)     { return s.repo.ListLocations() }
func (s *PublicService) CreateApplication(userID, opportunityID string, input ApplicationInput) (*domain.Application, error) {
	return s.repo.CreateApplication(domain.Application{OpportunityID: opportunityID, StudentUserID: userID, ResumeID: input.ResumeID, CoverLetter: input.CoverLetter})
}

func (s *StudentService) GetProfile(userID string) (*domain.StudentProfile, error) {
	return s.repo.GetStudentProfile(userID)
}
func (s *StudentService) UpdateProfile(userID string, input StudentProfileInput) (*domain.StudentProfile, error) {
	return s.repo.UpsertStudentProfile(domain.StudentProfile{UserID: userID, LastName: input.LastName, FirstName: input.FirstName, MiddleName: input.MiddleName, UniversityName: input.UniversityName, Faculty: input.Faculty, Specialization: input.Specialization, StudyYear: input.StudyYear, GraduationYear: input.GraduationYear, About: input.About, ProfileVisibility: input.ProfileVisibility, ShowResume: input.ShowResume, ShowApplications: input.ShowApplications, ShowCareerInterests: input.ShowCareerInterests, Telegram: input.Telegram, GithubURL: input.GithubURL, LinkedinURL: input.LinkedinURL, WebsiteURL: input.WebsiteURL, CityID: input.CityID}, userID)
}
func (s *StudentService) ListResumes(userID string) ([]domain.Resume, error) {
	return s.repo.ListResumes(userID)
}
func (s *StudentService) CreateResume(userID string, input ResumeInput) (*domain.Resume, error) {
	return s.repo.CreateResume(domain.Resume{StudentUserID: userID, Title: input.Title, Summary: input.Summary, ExperienceText: input.ExperienceText, EducationText: input.EducationText})
}
func (s *StudentService) SetPrimaryResume(userID, resumeID string) (*domain.Resume, error) {
	return s.repo.SetPrimaryResume(userID, resumeID)
}
func (s *StudentService) ListPortfolioProjects(userID string) ([]domain.PortfolioProject, error) {
	return s.repo.ListPortfolioProjects(userID)
}
func (s *StudentService) CreatePortfolioProject(userID string, input PortfolioProjectInput) (*domain.PortfolioProject, error) {
	return s.repo.CreatePortfolioProject(domain.PortfolioProject{StudentUserID: userID, Title: input.Title, Description: input.Description, ProjectURL: input.ProjectURL, RepositoryURL: input.RepositoryURL, DemoURL: input.DemoURL, StartedAt: input.StartedAt, FinishedAt: input.FinishedAt})
}
func (s *StudentService) ListApplications(userID string) ([]domain.Application, error) {
	return s.repo.ListStudentApplications(userID)
}
func (s *StudentService) ListFavoriteOpportunities(userID string) ([]domain.PublicOpportunity, error) {
	return s.repo.ListFavoriteOpportunities(userID)
}
func (s *StudentService) AddFavoriteOpportunity(userID, opportunityID string) error {
	return s.repo.AddFavoriteOpportunity(userID, opportunityID)
}
func (s *StudentService) RemoveFavoriteOpportunity(userID, opportunityID string) error {
	return s.repo.RemoveFavoriteOpportunity(userID, opportunityID)
}
func (s *StudentService) ListFavoriteCompanies(userID string) ([]domain.Company, error) {
	return s.repo.ListFavoriteCompanies(userID)
}
func (s *StudentService) AddFavoriteCompany(userID, companyID string) error {
	return s.repo.AddFavoriteCompany(userID, companyID)
}
func (s *StudentService) RemoveFavoriteCompany(userID, companyID string) error {
	return s.repo.RemoveFavoriteCompany(userID, companyID)
}
func (s *StudentService) ListContacts(userID string) ([]domain.User, error) {
	return s.repo.ListContacts(userID)
}
func (s *StudentService) ListContactRequests(userID string) ([]domain.ContactRequest, error) {
	return s.repo.ListContactRequests(userID)
}
func (s *StudentService) CreateContactRequest(userID string, input ContactRequestInput) (*domain.ContactRequest, error) {
	return s.repo.CreateContactRequest(userID, input.ReceiverUserID, input.Message)
}
func (s *StudentService) UpdateContactRequestStatus(userID, requestID string, input ContactRequestInput) (*domain.ContactRequest, error) {
	return s.repo.UpdateContactRequestStatus(requestID, userID, input.Status)
}
func (s *StudentService) CreateRecommendation(userID string, input RecommendationInput) (*domain.Recommendation, error) {
	return s.repo.CreateRecommendation(domain.Recommendation{FromUserID: userID, ToUserID: input.ToUserID, OpportunityID: input.OpportunityID, Message: input.Message})
}
func (s *StudentService) ListNotifications(userID string) ([]domain.Notification, error) {
	return s.repo.ListNotifications(userID)
}

func (s *EmployerService) GetCompany(userID string) (*domain.Company, error) {
	return s.repo.GetEmployerCompany(userID)
}
func (s *EmployerService) UpdateCompany(userID string, input CompanyInput) (*domain.Company, error) {
	return s.repo.UpdateEmployerCompany(userID, repository.CompanyUpdate{LegalName: input.LegalName, BrandName: input.BrandName, Description: input.Description, Industry: input.Industry, WebsiteURL: input.WebsiteURL, EmailDomain: input.EmailDomain, INN: input.INN, OGRN: input.OGRN, CompanySize: input.CompanySize, FoundedYear: input.FoundedYear, HQCityID: input.HQCityID})
}
func (s *EmployerService) CreateCompanyLink(userID string, input CompanyLinkInput) (*domain.CompanyLink, error) {
	return s.repo.CreateCompanyLink(userID, input.LinkType, input.URL)
}
func (s *EmployerService) SubmitVerification(userID string, input VerificationInput) (*domain.CompanyVerification, error) {
	return s.repo.SubmitCompanyVerification(userID, input.VerificationMethod, input.CorporateEmail, input.INNSubmitted, input.DocumentsComment)
}
func (s *EmployerService) ListOpportunities(userID string) ([]domain.Opportunity, error) {
	return s.repo.ListEmployerOpportunities(userID)
}
func (s *EmployerService) CreateOpportunity(userID string, input OpportunityInput) (*domain.Opportunity, error) {
	return s.repo.CreateOpportunity(buildOpportunity(userID, "", input))
}
func (s *EmployerService) GetOpportunity(userID, opportunityID string) (*domain.Opportunity, error) {
	return s.repo.GetEmployerOpportunity(userID, opportunityID)
}
func (s *EmployerService) UpdateOpportunity(userID, opportunityID string, input OpportunityInput) (*domain.Opportunity, error) {
	return s.repo.UpdateEmployerOpportunity(userID, buildOpportunity(userID, opportunityID, input))
}
func (s *EmployerService) ListApplications(userID, opportunityID string) ([]domain.Application, error) {
	return s.repo.ListOpportunityApplications(userID, opportunityID)
}
func (s *EmployerService) UpdateApplicationStatus(userID, applicationID, status string) (*domain.Application, error) {
	return s.repo.UpdateApplicationStatus(userID, applicationID, status)
}

func buildOpportunity(userID, opportunityID string, input OpportunityInput) domain.Opportunity {
	return domain.Opportunity{ID: opportunityID, CreatedByUserID: userID, Title: input.Title, ShortDescription: input.ShortDescription, FullDescription: input.FullDescription, OpportunityType: input.OpportunityType, VacancyLevel: input.VacancyLevel, EmploymentType: input.EmploymentType, WorkFormat: input.WorkFormat, LocationID: input.LocationID, SalaryMin: input.SalaryMin, SalaryMax: input.SalaryMax, SalaryCurrency: input.SalaryCurrency, IsSalaryVisible: input.IsSalaryVisible, ContactsInfo: input.ContactsInfo, ExternalURL: input.ExternalURL, Status: input.Status, TagIDs: input.TagIDs, ApplicationDeadline: parseTime(input.ApplicationDeadline), EventStartAt: parseTime(input.EventStartAt), EventEndAt: parseTime(input.EventEndAt), ExpiresAt: parseTime(input.ExpiresAt)}
}

func (s *CuratorService) CreateCurator(actorID string, input CuratorCreateInput) (*domain.User, error) {
	return s.repo.CreateCurator(input.Email, input.Password, input.DisplayName, input.CuratorType, actorID)
}
func (s *CuratorService) UpdateUserStatus(actorID, userID, status string) (*domain.User, error) {
	return s.repo.UpdateUserStatus(userID, status, actorID)
}
func (s *CuratorService) UpdateStudentProfile(actorID, userID string, input StudentProfileInput) (*domain.StudentProfile, error) {
	return s.repo.UpsertStudentProfile(domain.StudentProfile{UserID: userID, LastName: input.LastName, FirstName: input.FirstName, MiddleName: input.MiddleName, UniversityName: input.UniversityName, Faculty: input.Faculty, Specialization: input.Specialization, StudyYear: input.StudyYear, GraduationYear: input.GraduationYear, About: input.About, ProfileVisibility: input.ProfileVisibility, ShowResume: input.ShowResume, ShowApplications: input.ShowApplications, ShowCareerInterests: input.ShowCareerInterests, Telegram: input.Telegram, GithubURL: input.GithubURL, LinkedinURL: input.LinkedinURL, WebsiteURL: input.WebsiteURL, CityID: input.CityID}, actorID)
}
func (s *CuratorService) UpdateEmployerProfile(actorID, userID string, canCreate, canEdit, owner bool, position string) (*domain.EmployerProfile, error) {
	return s.repo.UpdateEmployerProfile(userID, domain.EmployerProfile{UserID: userID, PositionTitle: position, CanCreateOpportunities: canCreate, CanEditCompanyProfile: canEdit, IsCompanyOwner: owner}, actorID)
}
func (s *CuratorService) ListModerationQueue() ([]domain.ModerationQueueItem, error) {
	return s.repo.ListModerationQueue()
}
func (s *CuratorService) ReviewModerationQueue(actorID, itemID, status, comment string) (*domain.ModerationQueueItem, error) {
	return s.repo.ReviewModerationQueueItem(itemID, actorID, status, comment)
}
func (s *CuratorService) ListCompanyVerifications() ([]domain.CompanyVerification, error) {
	return s.repo.ListCompanyVerifications()
}
func (s *CuratorService) ReviewCompanyVerification(actorID, verificationID, status, comment string) (*domain.CompanyVerification, error) {
	return s.repo.ReviewCompanyVerification(verificationID, actorID, status, comment)
}
func (s *CuratorService) UpdateOpportunityStatus(actorID, opportunityID, status string) (*domain.Opportunity, error) {
	return s.repo.UpdateOpportunityStatus(actorID, opportunityID, status)
}
func (s *CuratorService) ListAuditLogs() ([]domain.AuditLog, error) { return s.repo.ListAuditLogs() }

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
