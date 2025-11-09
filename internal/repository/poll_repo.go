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

type IPollRepo interface {
	Create(poll *model.Poll) (*model.Poll, error)
	CreatePollOption(option *model.PollOption) error
	FindByPollID(pollID string) (*model.Poll, error)
	FindPollOptionsByPollID(pollID int64) ([]model.PollOption, error)
	FindActivePolls() ([]model.Poll, error)
	UpdateStatus(params *UpdateStatusParams) error
}

type PollRepo struct {
	db *gorm.DB
}

func NewPollRepository(db *gorm.DB) IPollRepo {
	return &PollRepo{db: db}
}

func (r *PollRepo) Create(poll *model.Poll) (*model.Poll, error) {
	err := r.db.Create(poll).Error
	return poll, err
}

func (r *PollRepo) CreatePollOption(option *model.PollOption) error {
	return r.db.Create(option).Error
}

func (r *PollRepo) FindByPollID(pollID string) (*model.Poll, error) {
	var poll model.Poll
	err := r.db.Preload("Voting").Preload("Movie").Where("poll_id = ? AND status = ?", pollID, "active").First(&poll).Error
	return &poll, err
}

func (r *PollRepo) FindPollOptionsByPollID(pollID int64) ([]model.PollOption, error) {
	var options []model.PollOption
	err := r.db.Preload("Movie").Where("poll_id = ?", pollID).Order("option_index").Find(&options).Error
	return options, err
}

func (r *PollRepo) FindActivePolls() ([]model.Poll, error) {
	var polls []model.Poll
	err := r.db.Preload("Voting").Preload("Movie").Where("status = ?", "active").Find(&polls).Error
	return polls, err
}

func (r *PollRepo) UpdateStatus(params *UpdateStatusParams) error {
	if params.Tx == nil {
		params.Tx = r.db
	}
	return params.Tx.Model(&model.Poll{}).Where("poll_id = ?", params.PollID).Update("status", params.Status).Error
}
