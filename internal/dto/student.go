package dto

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
