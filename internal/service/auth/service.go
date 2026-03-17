package auth

import (
	"errors"
	"strings"

	"tramplin/internal/authjwt"
	"tramplin/internal/dto"
	"tramplin/internal/repository"
)

type Service struct {
	repo repository.PlatformRepository
	jwt  *authjwt.Manager
}

func New(repo repository.PlatformRepository, jwtManager *authjwt.Manager) *Service {
	return &Service{repo: repo, jwt: jwtManager}
}

func (s *Service) Register(input dto.RegisterInput) (map[string]any, error) {
	user, role, err := s.repo.RegisterUser(repository.RegisterUserParams{
		Email:       input.Email,
		Password:    input.Password,
		DisplayName: input.DisplayName,
		Role:        strings.ToLower(input.Role),
		CompanyName: input.CompanyName,
	})
	if err != nil {
		return nil, err
	}
	token, expiresAt, err := s.jwt.Generate(user.ID, []string{role})
	if err != nil {
		return nil, err
	}
	return map[string]any{"user": user, "role": role, "access_token": token, "token_type": "Bearer", "expires_at": expiresAt}, nil
}

func (s *Service) Login(input dto.LoginInput, curatorOnly bool) (map[string]any, error) {
	user, roles, err := s.repo.Login(input.Email, input.Password)
	if err != nil {
		return nil, err
	}
	if curatorOnly {
		allowed := false
		for _, role := range roles {
			if role == repository.RoleCurator || role == repository.RoleAdmin {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, errors.New("curator access required")
		}
	}
	token, expiresAt, err := s.jwt.Generate(user.ID, roles)
	if err != nil {
		return nil, err
	}
	return map[string]any{"user": user, "roles": roles, "access_token": token, "token_type": "Bearer", "expires_at": expiresAt}, nil
}
