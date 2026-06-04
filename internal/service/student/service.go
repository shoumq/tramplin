package student

import (
	"fmt"
	"strings"
	"time"
	"tramplin/internal/dto"
	"tramplin/internal/models"
	"tramplin/internal/repository"

	"github.com/google/uuid"
)

type Service struct{ repo repository.PlatformRepository }

func New(repo repository.PlatformRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetProfile(userID string) (*models.StudentProfile, error) {
	return s.repo.GetStudentProfile(userID)
}

func (s *Service) UpdateProfile(userID string, input dto.StudentProfileInput) (*models.StudentProfile, error) {
	return s.repo.UpsertStudentProfile(studentProfileFromDTO(userID, input), userID)
}

func (s *Service) ListResumes(userID string) ([]models.Resume, error) {
	return s.repo.ListResumes(userID)
}

func (s *Service) CreateResume(userID string, input dto.ResumeInput) (*models.Resume, error) {
	return s.repo.CreateResume(models.Resume{
		StudentUserID:  userID,
		Title:          input.Title,
		Summary:        input.Summary,
		ExperienceText: input.ExperienceText,
		EducationText:  input.EducationText,
	})
}

func (s *Service) SetPrimaryResume(userID, resumeID string) (*models.Resume, error) {
	return s.repo.SetPrimaryResume(userID, resumeID)
}

func (s *Service) ListResumeWorkExperiences(userID, resumeID string) ([]models.ResumeWorkExperience, error) {
	return s.repo.ListResumeWorkExperiences(userID, resumeID)
}

func (s *Service) CreateResumeWorkExperience(userID, resumeID string, input dto.ResumeWorkExperienceInput) (*models.ResumeWorkExperience, error) {
	experience, err := buildResumeWorkExperience(input)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateResumeWorkExperience(userID, resumeID, experience)
}

func (s *Service) UpdateResumeWorkExperience(userID, resumeID, experienceID string, input dto.ResumeWorkExperienceInput) (*models.ResumeWorkExperience, error) {
	experience, err := buildResumeWorkExperience(input)
	if err != nil {
		return nil, err
	}
	experience.ID = experienceID
	return s.repo.UpdateResumeWorkExperience(userID, resumeID, experience)
}

func (s *Service) DeleteResumeWorkExperience(userID, resumeID, experienceID string) error {
	return s.repo.DeleteResumeWorkExperience(userID, resumeID, experienceID)
}

func (s *Service) ListPortfolioProjects(userID string) ([]models.PortfolioProject, error) {
	return s.repo.ListPortfolioProjects(userID)
}

func (s *Service) CreatePortfolioProject(userID string, input dto.PortfolioProjectInput) (*models.PortfolioProject, error) {
	return s.repo.CreatePortfolioProject(models.PortfolioProject{
		StudentUserID: userID,
		Title:         input.Title,
		Description:   input.Description,
		ProjectURL:    input.ProjectURL,
		RepositoryURL: input.RepositoryURL,
		DemoURL:       input.DemoURL,
		StartedAt:     input.StartedAt,
		FinishedAt:    input.FinishedAt,
	})
}

func (s *Service) ListApplications(userID string) ([]models.Application, error) {
	return s.repo.ListStudentApplications(userID)
}

func (s *Service) ListFavoriteOpportunities(userID string) ([]models.PublicOpportunity, error) {
	return s.repo.ListFavoriteOpportunities(userID)
}

func (s *Service) AddFavoriteOpportunity(userID, opportunityID string) error {
	return s.repo.AddFavoriteOpportunity(userID, opportunityID)
}

func (s *Service) RemoveFavoriteOpportunity(userID, opportunityID string) error {
	return s.repo.RemoveFavoriteOpportunity(userID, opportunityID)
}

func (s *Service) ListFavoriteCompanies(userID string) ([]models.Company, error) {
	return s.repo.ListFavoriteCompanies(userID)
}

func (s *Service) AddFavoriteCompany(userID, companyID string) error {
	return s.repo.AddFavoriteCompany(userID, companyID)
}

func (s *Service) RemoveFavoriteCompany(userID, companyID string) error {
	return s.repo.RemoveFavoriteCompany(userID, companyID)
}

func (s *Service) ListContacts(userID string) ([]models.User, error) {
	return s.repo.ListContacts(userID)
}

func (s *Service) ListContactRequests(userID string) ([]models.ContactRequest, error) {
	return s.repo.ListContactRequests(userID)
}

func (s *Service) CreateContactRequest(userID string, input dto.ContactRequestInput) (*models.ContactRequest, error) {
	receiverUserID := strings.TrimSpace(input.ReceiverUserID)
	if receiverUserID == "" {
		return nil, fmt.Errorf("receiver_user_id is required")
	}
	if _, err := uuid.Parse(receiverUserID); err != nil {
		return nil, fmt.Errorf("receiver_user_id must be a valid UUID")
	}
	return s.repo.CreateContactRequest(userID, receiverUserID, strings.TrimSpace(input.Message))
}

func (s *Service) UpdateContactRequestStatus(userID, requestID string, input dto.ContactRequestInput) (*models.ContactRequest, error) {
	return s.repo.UpdateContactRequestStatus(requestID, userID, input.Status)
}

func (s *Service) CreateRecommendation(userID string, input dto.RecommendationInput) (*models.Recommendation, error) {
	return s.repo.CreateRecommendation(models.Recommendation{
		FromUserID:    userID,
		ToUserID:      input.ToUserID,
		OpportunityID: input.OpportunityID,
		Message:       input.Message,
	})
}

func (s *Service) ListNotifications(userID string) ([]models.Notification, error) {
	return s.repo.ListNotifications(userID)
}

func (s *Service) MarkAllNotificationsRead(userID string) error {
	return s.repo.MarkAllNotificationsRead(userID)
}

func (s *Service) MarkNotificationRead(userID, notificationID string) error {
	return s.repo.MarkNotificationRead(userID, notificationID)
}

func studentProfileFromDTO(userID string, input dto.StudentProfileInput) models.StudentProfile {
	return models.StudentProfile{
		UserID:              userID,
		LastName:            input.LastName,
		FirstName:           input.FirstName,
		MiddleName:          input.MiddleName,
		UniversityName:      input.UniversityName,
		Faculty:             input.Faculty,
		Specialization:      input.Specialization,
		StudyYear:           input.StudyYear,
		GraduationYear:      input.GraduationYear,
		About:               input.About,
		ProfileVisibility:   input.ProfileVisibility,
		ShowResume:          input.ShowResume,
		ShowApplications:    input.ShowApplications,
		ShowCareerInterests: input.ShowCareerInterests,
		Telegram:            input.Telegram,
		GithubURL:           input.GithubURL,
		LinkedinURL:         input.LinkedinURL,
		WebsiteURL:          input.WebsiteURL,
		CityID:              input.CityID,
	}
}

func buildResumeWorkExperience(input dto.ResumeWorkExperienceInput) (models.ResumeWorkExperience, error) {
	positionTitle := strings.TrimSpace(input.PositionTitle)
	if positionTitle == "" {
		return models.ResumeWorkExperience{}, fmt.Errorf("position_title is required")
	}
	description := strings.TrimSpace(input.Description)
	if description == "" {
		return models.ResumeWorkExperience{}, fmt.Errorf("description is required")
	}
	companyID := strings.TrimSpace(input.CompanyID)
	companyName := strings.TrimSpace(input.CompanyName)
	if companyID == "" && companyName == "" {
		return models.ResumeWorkExperience{}, fmt.Errorf("company_id or company_name is required")
	}

	startedAt, err := time.Parse("2006-01-02", strings.TrimSpace(input.StartedAt))
	if err != nil {
		return models.ResumeWorkExperience{}, fmt.Errorf("started_at must be in YYYY-MM-DD format")
	}

	finishedAtValue := strings.TrimSpace(input.FinishedAt)
	finishedAt := ""
	if finishedAtValue != "" {
		parsedFinishedAt, err := time.Parse("2006-01-02", finishedAtValue)
		if err != nil {
			return models.ResumeWorkExperience{}, fmt.Errorf("finished_at must be in YYYY-MM-DD format")
		}
		if parsedFinishedAt.Before(startedAt) {
			return models.ResumeWorkExperience{}, fmt.Errorf("finished_at cannot be earlier than started_at")
		}
		finishedAt = parsedFinishedAt.Format("2006-01-02")
	}

	return models.ResumeWorkExperience{
		CompanyID:     companyID,
		CompanyName:   companyName,
		PositionTitle: positionTitle,
		StartedAt:     startedAt.Format("2006-01-02"),
		FinishedAt:    finishedAt,
		Description:   description,
	}, nil
}
