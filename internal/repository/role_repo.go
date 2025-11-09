package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type RoleRepo struct {
	db *gorm.DB
}

type IRoleRepo interface {
	FindByName(name string) (*model.Role, error)
}

func NewRoleRepository(db *gorm.DB) IRoleRepo {
	return &RoleRepo{db: db}
}

func (r *RoleRepo) FindByName(name string) (*model.Role, error) {
	var role *model.Role
	if err := r.db.Where(&model.Role{Name: name}).Find(&role).Error; err != nil {
		return nil, err
	}
	return role, nil
}
