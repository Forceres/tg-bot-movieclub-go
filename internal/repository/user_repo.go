package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

type IUserRepo interface {
	CreateIfNotExist(user *model.User) error
	FindByID(userID int64) (*model.User, error)
}

func NewUserRepository(db *gorm.DB) IUserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateIfNotExist(user *model.User) error {
	return r.db.Where(&model.User{ID: user.ID}).Attrs(user).FirstOrCreate(&user).Error
}

func (r *UserRepo) FindByID(userID int64) (*model.User, error) {
	var user model.User
	if err := r.db.Where(&model.User{ID: userID}).Preload("Role").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
