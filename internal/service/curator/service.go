package curator

import (
	"fmt"
	"strings"

	"tramplin/internal/dto"
	"tramplin/internal/models"
	"tramplin/internal/repository"
)

type Service struct{ repo repository.PlatformRepository }

func New(repo repository.PlatformRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCurator(actorID string, input dto.CuratorCreateInput) (*models.User, error) {
	return s.repo.CreateCurator(input.Email, input.Password, input.DisplayName, input.CuratorType, actorID)
}

func (s *Service) UpdateUserStatus(actorID, userID, status string) (*models.User, error) {
	return s.repo.UpdateUserStatus(userID, status, actorID)
}

func (s *Service) UpdateStudentProfile(actorID, userID string, input dto.StudentProfileInput) (*models.StudentProfile, error) {
	return s.repo.UpsertStudentProfile(models.StudentProfile{
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
	}, actorID)
}

func (s *Service) UpdateEmployerProfile(actorID, userID string, input dto.EmployerProfileInput) (*models.EmployerProfile, error) {
	return s.repo.UpdateEmployerProfile(userID, models.EmployerProfile{
		UserID:                 userID,
		PositionTitle:          input.PositionTitle,
		CanCreateOpportunities: input.CanCreateOpportunities,
		CanEditCompanyProfile:  input.CanEditCompanyProfile,
		IsCompanyOwner:         input.IsCompanyOwner,
	}, actorID)
}

func (s *Service) ListModerationQueue() ([]models.ModerationQueueItem, error) {
	return s.repo.ListModerationQueue()
}

func (s *Service) ReviewModerationQueue(actorID, itemID, status, comment string) (*models.ModerationQueueItem, error) {
	return s.repo.ReviewModerationQueueItem(itemID, actorID, status, comment)
}

func (s *Service) ListCompanyVerifications() ([]models.CompanyVerification, error) {
	return s.repo.ListCompanyVerifications()
}

func (s *Service) ReviewCompanyVerification(actorID, verificationID, status, comment string) (*models.CompanyVerification, error) {
	normalized, err := normalizeCompanyVerificationReviewStatus(status)
	if err != nil {
		return nil, err
	}
	return s.repo.ReviewCompanyVerification(verificationID, actorID, normalized, comment)
}

func (s *Service) UpdateOpportunityStatus(actorID, opportunityID, status string) (*models.Opportunity, error) {
	return s.repo.UpdateOpportunityStatus(actorID, opportunityID, status)
}

func (s *Service) ListAuditLogs() ([]models.AuditLog, error) {
	return s.repo.ListAuditLogs()
}

func normalizeCompanyVerificationReviewStatus(status string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "approved", "approve", "accept", "accepted":
		return "approved", nil
	case "rejected", "reject", "declined", "decline":
		return "rejected", nil
	case "needs_revision", "revision", "needs_changes", "changes_requested":
		return "needs_revision", nil
	default:
		return "", fmt.Errorf("status must be one of: approved, rejected, needs_revision")
	}
}
