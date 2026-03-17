package dto

import "tramplin/internal/models"

type MeStudentProfile struct {
	UserID              string `json:"user_id"`
	DisplayName         string `json:"display_name"`
	UniversityName      string `json:"university_name"`
	Faculty             string `json:"faculty,omitempty"`
	Specialization      string `json:"specialization,omitempty"`
	StudyYear           int    `json:"study_year,omitempty"`
	GraduationYear      int    `json:"graduation_year,omitempty"`
	About               string `json:"about,omitempty"`
	ProfileVisibility   string `json:"profile_visibility"`
	ShowResume          bool   `json:"show_resume"`
	ShowApplications    bool   `json:"show_applications"`
	ShowCareerInterests bool   `json:"show_career_interests"`
	Telegram            string `json:"telegram,omitempty"`
	GithubURL           string `json:"github_url,omitempty"`
	LinkedinURL         string `json:"linkedin_url,omitempty"`
	WebsiteURL          string `json:"website_url,omitempty"`
	CityID              int64  `json:"city_id,omitempty"`
}

type MeEmployerProfile struct {
	UserID                 string `json:"user_id"`
	CompanyID              string `json:"company_id"`
	PositionTitle          string `json:"position_title,omitempty"`
	IsCompanyOwner         bool   `json:"is_company_owner"`
	CanCreateOpportunities bool   `json:"can_create_opportunities"`
	CanEditCompanyProfile  bool   `json:"can_edit_company_profile"`
}

type MeResponse struct {
	User            *models.User       `json:"user"`
	Roles           []string           `json:"roles"`
	StudentProfile  *MeStudentProfile  `json:"student_profile,omitempty"`
	EmployerProfile *MeEmployerProfile `json:"employer_profile,omitempty"`
}
