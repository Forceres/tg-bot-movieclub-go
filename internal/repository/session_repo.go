package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConnectMoviesToSessionParams struct {
	SessionID int64
	MovieIDs  []int
	Tx        *gorm.DB
}

type FindOrCreateSessionParams struct {
	CreatedBy  int64
	FinishedAt *int64
	Tx         *gorm.DB
}

type FinishSessionParams struct {
	SessionID int64
	Tx        *gorm.DB
}

type ISessionRepo interface {
	FindOrCreateSession(params *FindOrCreateSessionParams) (*model.Session, error)
	ConnectMoviesToSession(params *ConnectMoviesToSessionParams) error
	FinishSession(params *FinishSessionParams) (*model.Session, error)
	CancelSession() (*model.Session, error)
	Transaction(fc func(tx *gorm.DB) error) error
}

type SessionRepo struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) ISessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Transaction(fc func(tx *gorm.DB) error) error {
	return r.db.Transaction(fc)
}

func (r *SessionRepo) CancelSession() (*model.Session, error) {
	var session model.Session
	err := r.db.Model(&session).Where(&model.Session{Status: "ongoing"}).Clauses(clause.Returning{}).Update("status", "canceled").Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepo) FinishSession(params *FinishSessionParams) (*model.Session, error) {
	var tx *gorm.DB = r.db
	if params.Tx != nil {
		tx = params.Tx
	}
	sessionID := params.SessionID
	session := &model.Session{ID: sessionID}
	err := tx.Model(session).Clauses(clause.Returning{}).Update("status", "finished").Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *SessionRepo) FindOrCreateSession(params *FindOrCreateSessionParams) (*model.Session, error) {
	var session *model.Session
	var tx *gorm.DB = r.db
	if params.Tx != nil {
		tx = params.Tx
	}
	err := tx.Where("status = ?", "ongoing").Attrs(&model.Session{Status: "ongoing", CreatedBy: params.CreatedBy, FinishedAt: *params.FinishedAt}).FirstOrCreate(&session).Error
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *SessionRepo) ConnectMoviesToSession(params *ConnectMoviesToSessionParams) error {
	var tx *gorm.DB = r.db
	if params.Tx != nil {
		tx = params.Tx
	}

	var movies []model.Movie
	if err := tx.Where("id IN ?", params.MovieIDs).Find(&movies).Error; err != nil {
		return err
	}
	var session model.Session
	if err := tx.First(&session, params.SessionID).Error; err != nil {
		return err
	}
	if err := tx.Model(&session).Association("Movies").Append(&movies); err != nil {
		return err
	}
	return nil
}
