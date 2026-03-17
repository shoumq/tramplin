package models

import "time"

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	DisplayName   string    `json:"display_name"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	AvatarObject  string    `json:"avatar_object,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	Status        string    `json:"status"`
	LastLoginAt   time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type StudentProfile struct {
	UserID              string    `json:"user_id"`
	AvatarURL           string    `json:"avatar_url,omitempty"`
	LastName            string    `json:"last_name"`
	FirstName           string    `json:"first_name"`
	MiddleName          string    `json:"middle_name,omitempty"`
	UniversityName      string    `json:"university_name"`
	Faculty             string    `json:"faculty,omitempty"`
	Specialization      string    `json:"specialization,omitempty"`
	StudyYear           int       `json:"study_year,omitempty"`
	GraduationYear      int       `json:"graduation_year,omitempty"`
	About               string    `json:"about,omitempty"`
	ProfileVisibility   string    `json:"profile_visibility"`
	ShowResume          bool      `json:"show_resume"`
	ShowApplications    bool      `json:"show_applications"`
	ShowCareerInterests bool      `json:"show_career_interests"`
	Telegram            string    `json:"telegram,omitempty"`
	GithubURL           string    `json:"github_url,omitempty"`
	LinkedinURL         string    `json:"linkedin_url,omitempty"`
	WebsiteURL          string    `json:"website_url,omitempty"`
	CityID              int64     `json:"city_id,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type EmployerProfile struct {
	UserID                 string    `json:"user_id"`
	AvatarURL              string    `json:"avatar_url,omitempty"`
	CompanyID              string    `json:"company_id"`
	PositionTitle          string    `json:"position_title,omitempty"`
	IsCompanyOwner         bool      `json:"is_company_owner"`
	CanCreateOpportunities bool      `json:"can_create_opportunities"`
	CanEditCompanyProfile  bool      `json:"can_edit_company_profile"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type CuratorProfile struct {
	UserID          string    `json:"user_id"`
	CuratorType     string    `json:"curator_type"`
	CreatedByUserID string    `json:"created_by_user_id,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Company struct {
	ID          string    `json:"id"`
	LegalName   string    `json:"legal_name"`
	BrandName   string    `json:"brand_name,omitempty"`
	Description string    `json:"description,omitempty"`
	Industry    string    `json:"industry,omitempty"`
	WebsiteURL  string    `json:"website_url,omitempty"`
	EmailDomain string    `json:"email_domain,omitempty"`
	INN         string    `json:"inn,omitempty"`
	OGRN        string    `json:"ogrn,omitempty"`
	CompanySize string    `json:"company_size,omitempty"`
	FoundedYear int       `json:"founded_year,omitempty"`
	HQCityID    int64     `json:"hq_city_id,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CompanyLink struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	LinkType  string    `json:"link_type"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type CompanyVerification struct {
	ID                 string    `json:"id"`
	CompanyID          string    `json:"company_id"`
	VerificationMethod string    `json:"verification_method"`
	SubmittedByUserID  string    `json:"submitted_by_user_id"`
	CorporateEmail     string    `json:"corporate_email,omitempty"`
	INNSubmitted       string    `json:"inn_submitted,omitempty"`
	DocumentsComment   string    `json:"documents_comment,omitempty"`
	Status             string    `json:"status"`
	ReviewedByUserID   string    `json:"reviewed_by_user_id,omitempty"`
	ReviewComment      string    `json:"review_comment,omitempty"`
	SubmittedAt        time.Time `json:"submitted_at"`
	ReviewedAt         time.Time `json:"reviewed_at,omitempty"`
}

type City struct {
	ID        int64   `json:"id"`
	Country   string  `json:"country"`
	Region    string  `json:"region,omitempty"`
	CityName  string  `json:"city_name"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

type Location struct {
	ID           string    `json:"id"`
	CityID       int64     `json:"city_id,omitempty"`
	AddressLine  string    `json:"address_line,omitempty"`
	PostalCode   string    `json:"postal_code,omitempty"`
	Latitude     float64   `json:"latitude,omitempty"`
	Longitude    float64   `json:"longitude,omitempty"`
	LocationType string    `json:"location_type"`
	DisplayText  string    `json:"display_text"`
	CreatedAt    time.Time `json:"created_at"`
}

type Tag struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	TagType         string    `json:"tag_type"`
	CreatedByUserID string    `json:"created_by_user_id,omitempty"`
	IsSystem        bool      `json:"is_system"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

type Opportunity struct {
	ID                  string    `json:"id"`
	CompanyID           string    `json:"company_id"`
	CreatedByUserID     string    `json:"created_by_user_id"`
	Title               string    `json:"title"`
	ShortDescription    string    `json:"short_description"`
	FullDescription     string    `json:"full_description"`
	OpportunityType     string    `json:"opportunity_type"`
	VacancyLevel        string    `json:"vacancy_level,omitempty"`
	EmploymentType      string    `json:"employment_type,omitempty"`
	WorkFormat          string    `json:"work_format"`
	LocationID          string    `json:"location_id,omitempty"`
	Latitude            float64   `json:"latitude,omitempty"`
	Longitude           float64   `json:"longitude,omitempty"`
	SalaryMin           float64   `json:"salary_min,omitempty"`
	SalaryMax           float64   `json:"salary_max,omitempty"`
	SalaryCurrency      string    `json:"salary_currency,omitempty"`
	IsSalaryVisible     bool      `json:"is_salary_visible"`
	ApplicationDeadline time.Time `json:"application_deadline,omitempty"`
	EventStartAt        time.Time `json:"event_start_at,omitempty"`
	EventEndAt          time.Time `json:"event_end_at,omitempty"`
	PublishedAt         time.Time `json:"published_at,omitempty"`
	ExpiresAt           time.Time `json:"expires_at,omitempty"`
	Status              string    `json:"status"`
	ContactsInfo        string    `json:"contacts_info,omitempty"`
	ExternalURL         string    `json:"external_url,omitempty"`
	ViewsCount          int       `json:"views_count"`
	FavoritesCount      int       `json:"favorites_count"`
	ApplicationsCount   int       `json:"applications_count"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	TagIDs              []string  `json:"tag_ids,omitempty"`
}

type Resume struct {
	ID             string    `json:"id"`
	StudentUserID  string    `json:"student_user_id"`
	Title          string    `json:"title"`
	Summary        string    `json:"summary,omitempty"`
	ExperienceText string    `json:"experience_text,omitempty"`
	EducationText  string    `json:"education_text,omitempty"`
	IsPrimary      bool      `json:"is_primary"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PortfolioProject struct {
	ID            string    `json:"id"`
	StudentUserID string    `json:"student_user_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	ProjectURL    string    `json:"project_url,omitempty"`
	RepositoryURL string    `json:"repository_url,omitempty"`
	DemoURL       string    `json:"demo_url,omitempty"`
	StartedAt     string    `json:"started_at,omitempty"`
	FinishedAt    string    `json:"finished_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Application struct {
	ID                    string    `json:"id"`
	OpportunityID         string    `json:"opportunity_id"`
	StudentUserID         string    `json:"student_user_id"`
	StudentAvatarURL      string    `json:"student_avatar_url,omitempty"`
	ResumeID              string    `json:"resume_id,omitempty"`
	CoverLetter           string    `json:"cover_letter,omitempty"`
	Status                string    `json:"status"`
	StatusChangedByUserID string    `json:"status_changed_by_user_id,omitempty"`
	StatusChangedAt       time.Time `json:"status_changed_at,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type ContactRequest struct {
	ID                string    `json:"id"`
	SenderUserID      string    `json:"sender_user_id"`
	ReceiverUserID    string    `json:"receiver_user_id"`
	SenderAvatarURL   string    `json:"sender_avatar_url,omitempty"`
	ReceiverAvatarURL string    `json:"receiver_avatar_url,omitempty"`
	Message           string    `json:"message,omitempty"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Recommendation struct {
	ID            string    `json:"id"`
	FromUserID    string    `json:"from_user_id"`
	ToUserID      string    `json:"to_user_id"`
	OpportunityID string    `json:"opportunity_id"`
	Message       string    `json:"message,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type ModerationQueueItem struct {
	ID                string    `json:"id"`
	EntityType        string    `json:"entity_type"`
	EntityID          string    `json:"entity_id"`
	SubmittedByUserID string    `json:"submitted_by_user_id"`
	AssignedToUserID  string    `json:"assigned_to_user_id,omitempty"`
	Status            string    `json:"status"`
	ModeratorComment  string    `json:"moderator_comment,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AuditLog struct {
	ID          string    `json:"id"`
	ActorUserID string    `json:"actor_user_id,omitempty"`
	EntityType  string    `json:"entity_type"`
	EntityID    string    `json:"entity_id"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
	Details     string    `json:"details,omitempty"`
}

type Notification struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	Type              string    `json:"type"`
	Title             string    `json:"title"`
	Body              string    `json:"body"`
	IsRead            bool      `json:"is_read"`
	RelatedEntityType string    `json:"related_entity_type,omitempty"`
	RelatedEntityID   string    `json:"related_entity_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type PublicOpportunity struct {
	Opportunity
	CompanyName string   `json:"company_name"`
	Location    string   `json:"location"`
	Tags        []string `json:"tags,omitempty"`
}

type OpportunityMarker struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	CompanyName     string  `json:"company_name"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	WorkFormat      string  `json:"work_format"`
	OpportunityType string  `json:"opportunity_type"`
	SalaryLabel     string  `json:"salary_label,omitempty"`
}
