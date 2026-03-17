package curator

import (
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
	return s.repo.ReviewCompanyVerification(verificationID, actorID, status, comment)
}

func (s *Service) UpdateOpportunityStatus(actorID, opportunityID, status string) (*models.Opportunity, error) {
	return s.repo.UpdateOpportunityStatus(actorID, opportunityID, status)
}

func (s *Service) ListAuditLogs() ([]models.AuditLog, error) {
	return s.repo.ListAuditLogs()
}
