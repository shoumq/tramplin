package employer

import (
	"time"

	"tramplin/internal/dto"
	"tramplin/internal/models"
	"tramplin/internal/repository"
)

type Service struct{ repo repository.PlatformRepository }

func New(repo repository.PlatformRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCompany(userID string) (*models.Company, error) {
	return s.repo.GetEmployerCompany(userID)
}

func (s *Service) UpdateCompany(userID string, input dto.CompanyInput) (*models.Company, error) {
	return s.repo.UpdateEmployerCompany(userID, repository.CompanyUpdate{
		LegalName:   input.LegalName,
		BrandName:   input.BrandName,
		Description: input.Description,
		Industry:    input.Industry,
		WebsiteURL:  input.WebsiteURL,
		EmailDomain: input.EmailDomain,
		INN:         input.INN,
		OGRN:        input.OGRN,
		CompanySize: input.CompanySize,
		FoundedYear: input.FoundedYear,
		HQCityID:    input.HQCityID,
	})
}

func (s *Service) CreateCompanyLink(userID string, input dto.CompanyLinkInput) (*models.CompanyLink, error) {
	return s.repo.CreateCompanyLink(userID, input.LinkType, input.URL)
}

func (s *Service) SubmitVerification(userID string, input dto.VerificationInput) (*models.CompanyVerification, error) {
	return s.repo.SubmitCompanyVerification(userID, input.VerificationMethod, input.CorporateEmail, input.INNSubmitted, input.DocumentsComment)
}

func (s *Service) ListOpportunities(userID string) ([]models.Opportunity, error) {
	return s.repo.ListEmployerOpportunities(userID)
}

func (s *Service) CreateOpportunity(userID string, input dto.OpportunityInput) (*models.Opportunity, error) {
	return s.repo.CreateOpportunity(buildOpportunity(userID, "", input))
}

func (s *Service) GetOpportunity(userID, opportunityID string) (*models.Opportunity, error) {
	return s.repo.GetEmployerOpportunity(userID, opportunityID)
}

func (s *Service) UpdateOpportunity(userID, opportunityID string, input dto.OpportunityInput) (*models.Opportunity, error) {
	return s.repo.UpdateEmployerOpportunity(userID, buildOpportunity(userID, opportunityID, input))
}

func (s *Service) ListApplications(userID, opportunityID string) ([]models.Application, error) {
	return s.repo.ListOpportunityApplications(userID, opportunityID)
}

func (s *Service) UpdateApplicationStatus(userID, applicationID, status string) (*models.Application, error) {
	return s.repo.UpdateApplicationStatus(userID, applicationID, status)
}

func buildOpportunity(userID, opportunityID string, input dto.OpportunityInput) models.Opportunity {
	return models.Opportunity{
		ID:                  opportunityID,
		CreatedByUserID:     userID,
		Title:               input.Title,
		ShortDescription:    input.ShortDescription,
		FullDescription:     input.FullDescription,
		OpportunityType:     input.OpportunityType,
		VacancyLevel:        input.VacancyLevel,
		EmploymentType:      input.EmploymentType,
		WorkFormat:          input.WorkFormat,
		LocationID:          input.LocationID,
		SalaryMin:           input.SalaryMin,
		SalaryMax:           input.SalaryMax,
		SalaryCurrency:      input.SalaryCurrency,
		IsSalaryVisible:     input.IsSalaryVisible,
		ContactsInfo:        input.ContactsInfo,
		ExternalURL:         input.ExternalURL,
		Status:              input.Status,
		TagIDs:              input.TagIDs,
		ApplicationDeadline: parseTime(input.ApplicationDeadline),
		EventStartAt:        parseTime(input.EventStartAt),
		EventEndAt:          parseTime(input.EventEndAt),
		ExpiresAt:           parseTime(input.ExpiresAt),
	}
}

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
