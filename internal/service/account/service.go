package account

import (
	"context"
	"fmt"
	"io"

	"tramplin/internal/dto"
	"tramplin/internal/models"
	"tramplin/internal/repository"
	"tramplin/internal/storage"
)

type Service struct {
	repo    repository.PlatformRepository
	storage storage.Storage
}

func New(repo repository.PlatformRepository, storage storage.Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

func (s *Service) UploadAvatar(ctx context.Context, userID, fileName, contentType string, size int64, body io.Reader) (*models.User, error) {
	result, err := s.storage.UploadAvatar(ctx, userID, fileName, contentType, size, body)
	if err != nil {
		return nil, fmt.Errorf("upload avatar: %w", err)
	}
	return s.repo.UpdateUserAvatar(userID, result.ObjectKey, result.URL)
}

func (s *Service) GetMe(userID string) (*dto.MeResponse, error) {
	user, err := s.repo.GetUser(userID)
	if err != nil {
		return nil, err
	}
	roles, err := s.repo.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	response := &dto.MeResponse{
		User:  user,
		Roles: roles,
	}

	for _, role := range roles {
		switch role {
		case repository.RoleStudent:
			profile, err := s.repo.GetStudentProfile(userID)
			if err == nil {
				response.StudentProfile = &dto.MeStudentProfile{
					UserID:              profile.UserID,
					DisplayName:         user.DisplayName,
					UniversityName:      profile.UniversityName,
					Faculty:             profile.Faculty,
					Specialization:      profile.Specialization,
					StudyYear:           profile.StudyYear,
					GraduationYear:      profile.GraduationYear,
					About:               profile.About,
					ProfileVisibility:   profile.ProfileVisibility,
					ShowResume:          profile.ShowResume,
					ShowApplications:    profile.ShowApplications,
					ShowCareerInterests: profile.ShowCareerInterests,
					Telegram:            profile.Telegram,
					GithubURL:           profile.GithubURL,
					LinkedinURL:         profile.LinkedinURL,
					WebsiteURL:          profile.WebsiteURL,
					CityID:              profile.CityID,
				}
			}
		case repository.RoleEmployer:
			profile, err := s.repo.GetEmployerProfile(userID)
			if err == nil {
				response.EmployerProfile = &dto.MeEmployerProfile{
					UserID:                 profile.UserID,
					CompanyID:              profile.CompanyID,
					PositionTitle:          profile.PositionTitle,
					IsCompanyOwner:         profile.IsCompanyOwner,
					CanCreateOpportunities: profile.CanCreateOpportunities,
					CanEditCompanyProfile:  profile.CanEditCompanyProfile,
				}
			}
		}
	}

	return response, nil
}

func (s *Service) TouchPresence(userID string, isOnline bool) error {
	return s.repo.TouchUserPresence(userID, isOnline)
}
