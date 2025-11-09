package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type ISessionRepo interface {
	FindOrCreateSession(createdBy *int64) (*model.Session, error)
	ConnectMoviesToSession(sessionID int64, movieIDs []int) error
}

type SessionRepo struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) ISessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) FindOrCreateSession(createdBy *int64) (*model.Session, error) {
	var session *model.Session
	err := r.db.Where("status = ?", "ongoing").Attrs(&model.Session{Status: "ongoing", CreatedBy: *createdBy}).FirstOrCreate(&session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *SessionRepo) ConnectMoviesToSession(sessionID int64, movieIDs []int) error {
	var movies []model.Movie
	if err := r.db.Where("id IN ?", movieIDs).Find(&movies).Error; err != nil {
		return err
	}
	var session model.Session
	if err := r.db.First(&session, sessionID).Error; err != nil {
		return err
	}
	if err := r.db.Model(&session).Association("Movies").Append(&movies); err != nil {
		return err
	}
	return nil
}
