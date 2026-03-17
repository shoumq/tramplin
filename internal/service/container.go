package service

import (
	"tramplin/internal/authjwt"
	accountservice "tramplin/internal/service/account"
	authservice "tramplin/internal/service/auth"
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
	Public   *publicservice.Service
	Student  *studentservice.Service
	Employer *employerservice.Service
	Curator  *curatorservice.Service
}

func New(repo repository.PlatformRepository, storage storage.Storage, jwtManager *authjwt.Manager) *Services {
	return &Services{
		Account:  accountservice.New(repo, storage),
		Auth:     authservice.New(repo, jwtManager),
		Public:   publicservice.New(repo),
		Student:  studentservice.New(repo),
		Employer: employerservice.New(repo),
		Curator:  curatorservice.New(repo),
	}
}
