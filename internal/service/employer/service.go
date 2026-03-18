package employer

import (
	"fmt"
	"strings"
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
	method, err := normalizeVerificationMethod(input.VerificationMethod)
	if err != nil {
		return nil, err
	}
	if method == "corporate_email" && strings.TrimSpace(input.CorporateEmail) == "" {
		return nil, fmt.Errorf("corporate_email is required for verification_method=corporate_email")
	}
	if method == "inn_check" && strings.TrimSpace(input.INNSubmitted) == "" {
		return nil, fmt.Errorf("inn_submitted is required for verification_method=inn_check")
	}
	return s.repo.SubmitCompanyVerification(userID, method, input.CorporateEmail, input.INNSubmitted, input.DocumentsComment)
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

func normalizeVerificationMethod(method string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(method))
	switch normalized {
	case "corporate_email", "email", "corporate", "corporate_mail":
		return "corporate_email", nil
	case "inn_check", "inn":
		return "inn_check", nil
	case "manual_documents", "documents", "docs", "manual":
		return "manual_documents", nil
	case "social_links_review", "social_review", "social_links":
		return "social_links_review", nil
	case "combined", "all":
		return "combined", nil
	default:
		return "", fmt.Errorf("verification_method must be one of: corporate_email, inn_check, manual_documents, social_links_review, combined")
	}
}
