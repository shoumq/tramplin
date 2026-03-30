package dto

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

type OpportunityLocationInput struct {
	AddressLine string   `json:"address_line"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	DisplayText string   `json:"display_text"`
}

type OpportunityInput struct {
	Title               string                    `json:"title"`
	ShortDescription    string                    `json:"short_description"`
	FullDescription     string                    `json:"full_description"`
	OpportunityType     string                    `json:"opportunity_type"`
	VacancyLevel        string                    `json:"vacancy_level"`
	EmploymentType      string                    `json:"employment_type"`
	WorkFormat          string                    `json:"work_format"`
	LocationID          string                    `json:"location_id"`
	LocationInput       *OpportunityLocationInput `json:"location_input"`
	SalaryMin           float64                   `json:"salary_min"`
	SalaryMax           float64                   `json:"salary_max"`
	SalaryCurrency      string                    `json:"salary_currency"`
	IsSalaryVisible     bool                      `json:"is_salary_visible"`
	ContactsInfo        string                    `json:"contacts_info"`
	ExternalURL         string                    `json:"external_url"`
	Status              string                    `json:"status"`
	TagIDs              []string                  `json:"tag_ids"`
	ApplicationDeadline string                    `json:"application_deadline"`
	EventStartAt        string                    `json:"event_start_at"`
	EventEndAt          string                    `json:"event_end_at"`
	ExpiresAt           string                    `json:"expires_at"`
}

type EmployerProfileInput struct {
	PositionTitle          string `json:"position_title" example:"HR Lead"`
	IsCompanyOwner         bool   `json:"is_company_owner" example:"true"`
	CanCreateOpportunities bool   `json:"can_create_opportunities" example:"true"`
	CanEditCompanyProfile  bool   `json:"can_edit_company_profile" example:"true"`
}
