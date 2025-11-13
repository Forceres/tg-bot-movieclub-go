package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type UpdateStatusParams struct {
	PollID string
	Status string
	Tx     *gorm.DB
}

type CreatePollParams struct {
	Poll *model.Poll
	Tx   *gorm.DB
}

type IPollRepo interface {
	Create(params *CreatePollParams) (*model.Poll, error)
	CreatePollOption(option *model.PollOption) error
	FindByPollID(pollID string) (*model.Poll, error)
	FindOpenedByMovieID(movieID int) (*model.Poll, error)
	FindPollOptionsByPollID(pollID int64) ([]model.PollOption, error)
	UpdateStatus(params *UpdateStatusParams) error
}

type PollRepo struct {
	db *gorm.DB
}

func NewPollRepository(db *gorm.DB) IPollRepo {
	return &PollRepo{db: db}
}

func (r *PollRepo) Create(params *CreatePollParams) (*model.Poll, error) {
	var tx *gorm.DB = r.db
	if params.Tx != nil {
		tx = params.Tx
	}
	err := tx.Create(params.Poll).Error
	return params.Poll, err
}

func (r *PollRepo) CreatePollOption(option *model.PollOption) error {
	return r.db.Create(option).Error
}

func (r *PollRepo) FindByPollID(pollID string) (*model.Poll, error) {
	var poll *model.Poll
	err := r.db.Model(&model.Poll{}).Preload("Voting").Preload("Movie").Where("poll_id = ? AND status = ?", pollID, model.POLL_OPENED_STATUS).First(&poll).Error
	return poll, err
}

func (r *PollRepo) FindOpenedByMovieID(movieID int) (*model.Poll, error) {
	var poll model.Poll
	err := r.db.Preload("Voting").Where("movie_id = ? AND status = ?", movieID, model.POLL_OPENED_STATUS).First(&poll).Error
	if err != nil {
		return nil, err
	}
	return &poll, nil
}

func (r *PollRepo) FindPollOptionsByPollID(pollID int64) ([]model.PollOption, error) {
	var options []model.PollOption
	err := r.db.Preload("Movie").Where("poll_id = ?", pollID).Order("option_index").Find(&options).Error
	return options, err
}

func (r *PollRepo) UpdateStatus(params *UpdateStatusParams) error {
	var tx *gorm.DB = r.db
	if params.Tx != nil {
		tx = params.Tx
	}
	return tx.Model(&model.Poll{}).Where(&model.Poll{PollID: params.PollID}).Update("status", params.Status).Error
}
