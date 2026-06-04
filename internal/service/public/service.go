package public

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"tramplin/internal/dto"
	"tramplin/internal/models"
	"tramplin/internal/repository"
	aiservice "tramplin/internal/service/ai"
)

var ErrOpportunityIsNotVacancy = errors.New("opportunity is not a vacancy")
var ErrStudentResumesHidden = errors.New("student resumes are hidden")

type Service struct {
	repo repository.PlatformRepository
	ai   *aiservice.Client
}

func New(repo repository.PlatformRepository, aiClient *aiservice.Client) *Service {
	return &Service{repo: repo, ai: aiClient}
}

func (s *Service) ListOpportunities(params map[string]string) ([]models.PublicOpportunity, error) {
	return s.repo.ListOpportunities(buildFilter(params))
}

func (s *Service) ListRecommendationOpportunities(params map[string]string) ([]models.PublicOpportunity, error) {
	paramsCopy := make(map[string]string, len(params)+1)
	for k, v := range params {
		paramsCopy[k] = v
	}
	paramsCopy["type"] = "vacancy"
	return s.repo.ListOpportunities(buildFilter(paramsCopy))
}

func (s *Service) ListOpportunityMarkers(params map[string]string) ([]models.OpportunityMarker, error) {
	return s.repo.ListOpportunityMarkers(buildFilter(params))
}

func (s *Service) GetOpportunity(id string) (*models.PublicOpportunity, error) {
	return s.repo.GetOpportunity(id)
}

func (s *Service) GetNetworking(userID string) (*models.NetworkingOverview, error) {
	contacts, err := s.repo.ListContacts(userID)
	if err != nil {
		return nil, err
	}
	requests, err := s.repo.ListContactRequests(userID)
	if err != nil {
		return nil, err
	}

	overview := &models.NetworkingOverview{
		UserID:           userID,
		Contacts:         contacts,
		IncomingRequests: make([]models.ContactRequest, 0),
		OutgoingRequests: make([]models.ContactRequest, 0),
		SuggestedPeople:  make([]models.NetworkingSuggestion, 0),
	}
	for _, request := range requests {
		switch {
		case request.ReceiverUserID == userID:
			overview.IncomingRequests = append(overview.IncomingRequests, request)
		case request.SenderUserID == userID:
			overview.OutgoingRequests = append(overview.OutgoingRequests, request)
		}
	}

	suggestionMap := make(map[string]models.NetworkingSuggestion)
	for _, suggestion := range s.buildMutualContactSuggestions(userID) {
		suggestionMap[suggestion.User.ID] = suggestion
	}
	for _, suggestion := range s.buildProfileSuggestions(userID, suggestionMap) {
		if _, exists := suggestionMap[suggestion.User.ID]; exists {
			continue
		}
		suggestionMap[suggestion.User.ID] = suggestion
	}

	for _, suggestion := range suggestionMap {
		overview.SuggestedPeople = append(overview.SuggestedPeople, suggestion)
	}
	sort.SliceStable(overview.SuggestedPeople, func(i, j int) bool {
		if overview.SuggestedPeople[i].MutualContactsCount == overview.SuggestedPeople[j].MutualContactsCount {
			return strings.ToLower(overview.SuggestedPeople[i].User.DisplayName) < strings.ToLower(overview.SuggestedPeople[j].User.DisplayName)
		}
		return overview.SuggestedPeople[i].MutualContactsCount > overview.SuggestedPeople[j].MutualContactsCount
	})
	if len(overview.SuggestedPeople) > 8 {
		overview.SuggestedPeople = overview.SuggestedPeople[:8]
	}
	return overview, nil
}

func (s *Service) buildMutualContactSuggestions(userID string) []models.NetworkingSuggestion {
	suggestions, err := s.repo.ListNetworkingSuggestions(userID, 8)
	if err != nil {
		return nil
	}
	return suggestions
}

func (s *Service) buildProfileSuggestions(userID string, existing map[string]models.NetworkingSuggestion) []models.NetworkingSuggestion {
	roles, err := s.repo.GetUserRoles(userID)
	if err != nil {
		return nil
	}
	if !containsRole(roles, repository.RoleStudent) {
		return nil
	}

	profile, err := s.repo.GetStudentProfile(userID)
	if err != nil {
		return nil
	}

	limit := 20
	people, err := s.repo.ListPublicStudentProfiles(repository.StudentFilter{
		ViewerUserID:   userID,
		UniversityName: profile.UniversityName,
		Faculty:        profile.Faculty,
		Specialization: profile.Specialization,
	})
	if err != nil {
		return nil
	}

	result := make([]models.NetworkingSuggestion, 0, limit)
	for _, person := range people {
		if person.UserID == userID || person.ContactRelation != "none" {
			continue
		}
		if _, exists := existing[person.UserID]; exists {
			continue
		}
		user, err := s.repo.GetUser(person.UserID)
		if err != nil {
			continue
		}
		result = append(result, models.NetworkingSuggestion{
			User:                *user,
			MutualContactsCount: 0,
			Reason:              networkingProfileReason(profile, person),
			Source:              "student_profile",
		})
		if len(result) >= limit {
			break
		}
	}
	return result
}

func networkingProfileReason(viewer *models.StudentProfile, target models.PublicStudentProfile) string {
	if viewer == nil {
		return "Публичный профиль на платформе"
	}
	sameUniversity := strings.EqualFold(strings.TrimSpace(viewer.UniversityName), strings.TrimSpace(target.UniversityName))
	sameFaculty := strings.EqualFold(strings.TrimSpace(viewer.Faculty), strings.TrimSpace(target.Faculty)) && strings.TrimSpace(viewer.Faculty) != ""
	sameSpecialization := strings.EqualFold(strings.TrimSpace(viewer.Specialization), strings.TrimSpace(target.Specialization)) && strings.TrimSpace(viewer.Specialization) != ""

	switch {
	case sameUniversity && sameFaculty && sameSpecialization:
		return "Общий университет, факультет и специализация"
	case sameUniversity && sameFaculty:
		return "Общий университет и факультет"
	case sameUniversity:
		return "Общий университет"
	default:
		return "Публичный профиль на платформе"
	}
}

func containsRole(roles []string, target string) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}

func (s *Service) AnalyzeOpportunity(ctx context.Context, opportunityID string) (*dto.OpportunityAIAnalytics, error) {
	opportunity, err := s.repo.GetOpportunity(opportunityID)
	if err != nil {
		return nil, err
	}
	if opportunity.OpportunityType != "vacancy" {
		return nil, ErrOpportunityIsNotVacancy
	}
	analysis, err := s.ai.AnalyzeVacancyForApplicant(ctx, opportunity.Opportunity)
	if err != nil {
		return nil, err
	}
	return &dto.OpportunityAIAnalytics{
		OpportunityID: opportunity.ID,
		Model:         s.ai.Model(),
		Analysis:      analysis,
		GeneratedAt:   time.Now(),
	}, nil
}

func (s *Service) ListCompanies() ([]models.Company, error) {
	return s.repo.ListCompanies()
}

func (s *Service) GetCompany(id string) (*models.Company, error) {
	return s.repo.GetCompany(id)
}

func (s *Service) GetStudentProfile(id, viewerUserID string) (*models.PublicStudentProfile, error) {
	return s.repo.GetPublicStudentProfile(id, viewerUserID)
}

func (s *Service) ListStudentResumes(studentID, viewerUserID string) ([]models.Resume, error) {
	if strings.TrimSpace(studentID) != strings.TrimSpace(viewerUserID) {
		profile, err := s.repo.GetPublicStudentProfile(studentID, viewerUserID)
		if err != nil {
			return nil, err
		}
		if !profile.ShowResume {
			return nil, ErrStudentResumesHidden
		}
	}
	return s.repo.ListResumes(studentID)
}

func (s *Service) GetStudentResume(studentID, resumeID, viewerUserID string) (*models.StudentResumeDetail, error) {
	if strings.TrimSpace(studentID) != strings.TrimSpace(viewerUserID) {
		profile, err := s.repo.GetPublicStudentProfile(studentID, viewerUserID)
		if err != nil {
			return nil, err
		}
		if !profile.ShowResume {
			return nil, ErrStudentResumesHidden
		}
	}
	resume, err := s.repo.GetResume(studentID, resumeID)
	if err != nil {
		return nil, err
	}
	experiences, err := s.repo.ListResumeWorkExperiences(studentID, resumeID)
	if err != nil {
		return nil, err
	}
	return &models.StudentResumeDetail{
		Resume:          *resume,
		WorkExperiences: experiences,
	}, nil
}

func (s *Service) ListStudents(params map[string]string, viewerUserID string) ([]models.PublicStudentProfile, error) {
	studyYear, _ := strconv.Atoi(strings.TrimSpace(params["study_year"]))
	return s.repo.ListPublicStudentProfiles(repository.StudentFilter{
		ViewerUserID:   viewerUserID,
		Search:         strings.TrimSpace(params["search"]),
		UniversityName: strings.TrimSpace(params["university_name"]),
		Faculty:        strings.TrimSpace(params["faculty"]),
		Specialization: strings.TrimSpace(params["specialization"]),
		StudyYear:      studyYear,
	})
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
