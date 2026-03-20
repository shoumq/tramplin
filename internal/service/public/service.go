package public

import (
	"strconv"

	"tramplin/internal/dto"
	"tramplin/internal/models"
	"tramplin/internal/repository"
)

type Service struct{ repo repository.PlatformRepository }

func New(repo repository.PlatformRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListOpportunities(params map[string]string) ([]models.PublicOpportunity, error) {
	return s.repo.ListOpportunities(buildFilter(params))
}

func (s *Service) ListOpportunityMarkers(params map[string]string) ([]models.OpportunityMarker, error) {
	return s.repo.ListOpportunityMarkers(buildFilter(params))
}

func (s *Service) GetOpportunity(id string) (*models.PublicOpportunity, error) {
	return s.repo.GetOpportunity(id)
}

func (s *Service) ListCompanies() ([]models.Company, error) {
	return s.repo.ListCompanies()
}

func (s *Service) GetCompany(id string) (*models.Company, error) {
	return s.repo.GetCompany(id)
}

func (s *Service) GetStudentProfile(id string) (*models.PublicStudentProfile, error) {
	return s.repo.GetPublicStudentProfile(id)
}

func (s *Service) ListTags() ([]models.Tag, error) {
	return s.repo.ListTags()
}

func (s *Service) ListCities() ([]models.City, error) {
	return s.repo.ListCities()
}

func (s *Service) ListLocations() ([]models.Location, error) {
	return s.repo.ListLocations()
}

func (s *Service) CreateApplication(userID, opportunityID string, input dto.ApplicationInput) (*models.Application, error) {
	return s.repo.CreateApplication(models.Application{
		OpportunityID: opportunityID,
		StudentUserID: userID,
		ResumeID:      input.ResumeID,
		CoverLetter:   input.CoverLetter,
	})
}

func (s *Service) GetUserPresence(userID string) (*models.Presence, error) {
	return s.repo.GetUserPresence(userID)
}

func (s *Service) GetCompanyPresence(companyID string) (*models.Presence, error) {
	return s.repo.GetCompanyPresence(companyID)
}

func buildFilter(params map[string]string) repository.OpportunityFilter {
	salary, _ := strconv.ParseFloat(params["salary_from"], 64)
	return repository.OpportunityFilter{
		Tag:        params["tag"],
		WorkFormat: params["work_format"],
		Type:       params["type"],
		Search:     params["search"],
		SalaryFrom: salary,
	}
}
