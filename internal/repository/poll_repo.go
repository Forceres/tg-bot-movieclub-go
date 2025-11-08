package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type IPollRepository interface {
	Create(poll *model.Poll) (*model.Poll, error)
	CreatePollOption(option *model.PollOption) error
	FindByPollID(pollID string) (*model.Poll, error)
	FindPollOptionsByPollID(pollID int64) ([]model.PollOption, error)
	FindActivePolls() ([]model.Poll, error)
	UpdateStatus(pollID string, status string) error
}

type PollRepository struct {
	db *gorm.DB
}

func NewPollRepository(db *gorm.DB) IPollRepository {
	return &PollRepository{db: db}
}

func (r *PollRepository) Create(poll *model.Poll) (*model.Poll, error) {
	err := r.db.Create(poll).Error
	return poll, err
}

func (r *PollRepository) CreatePollOption(option *model.PollOption) error {
	return r.db.Create(option).Error
}

func (r *PollRepository) FindByPollID(pollID string) (*model.Poll, error) {
	var poll model.Poll
	err := r.db.Preload("Voting").Preload("Movie").Where("poll_id = ? AND status = ?", pollID, "active").First(&poll).Error
	return &poll, err
}

func (r *PollRepository) FindPollOptionsByPollID(pollID int64) ([]model.PollOption, error) {
	var options []model.PollOption
	err := r.db.Preload("Movie").Where("poll_id = ?", pollID).Order("option_index").Find(&options).Error
	return options, err
}

func (r *PollRepository) FindActivePolls() ([]model.Poll, error) {
	var polls []model.Poll
	err := r.db.Preload("Voting").Preload("Movie").Where("status = ?", "active").Find(&polls).Error
	return polls, err
}

func (r *PollRepository) UpdateStatus(pollID string, status string) error {
	return r.db.Model(&model.Poll{}).Where("poll_id = ?", pollID).Update("status", status).Error
}
