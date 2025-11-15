package service

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
)

type UserService struct {
	repo     repository.IUserRepo
	roleRepo repository.IRoleRepo
}

type IUserService interface {
	Create(user *model.User, role string) error
	FindByID(userID int64) (*model.User, error)
}

func NewUserService(repo repository.IUserRepo, roleRepo repository.IRoleRepo) *UserService {
	return &UserService{repo: repo, roleRepo: roleRepo}
}

func (s *UserService) Create(user *model.User, role string) error {
	roleModel, err := s.roleRepo.FindByName(role)
	if err != nil {
		return err
	}
	user.RoleID = roleModel.ID
	return s.repo.Save(user)
}

func (s *UserService) FindByID(userID int64) (*model.User, error) {
	return s.repo.FindByID(userID)
}
