package service

import (
	"tramplin/internal/authjwt"
	accountservice "tramplin/internal/service/account"
	aiservice "tramplin/internal/service/ai"
	authservice "tramplin/internal/service/auth"
	chatservice "tramplin/internal/service/chat"
	curatorservice "tramplin/internal/service/curator"
	employerservice "tramplin/internal/service/employer"
	publicservice "tramplin/internal/service/public"
	studentservice "tramplin/internal/service/student"

	"tramplin/internal/repository"
	"tramplin/internal/storage"
)

type Services struct {
	Account  *accountservice.Service
	Auth     *authservice.Service
	Chat     *chatservice.Service
	Public   *publicservice.Service
	Student  *studentservice.Service
	Employer *employerservice.Service
	Curator  *curatorservice.Service
}

func New(repo repository.PlatformRepository, storage storage.Storage, jwtManager *authjwt.Manager, aiClient *aiservice.Client) *Services {
	return &Services{
		Account:  accountservice.New(repo, storage),
		Auth:     authservice.New(repo, jwtManager),
		Chat:     chatservice.New(repo),
		Public:   publicservice.New(repo, aiClient),
		Student:  studentservice.New(repo),
		Employer: employerservice.New(repo, storage, aiClient),
		Curator:  curatorservice.New(repo),
	}
}
