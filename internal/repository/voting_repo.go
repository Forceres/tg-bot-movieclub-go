package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type IVotingRepo interface {
	CreateVoting(voting *model.Voting) (*model.Voting, error)
	FindVotingByID(id int64) (*model.Voting, error)
	FindVotingsByStatus(status string) ([]*model.Voting, error)
	UpdateVotingStatus(voting *model.Voting) (*model.Voting, error)
}

type VotingRepo struct {
	db *gorm.DB
}

func NewVotingRepository(db *gorm.DB) *VotingRepo {
	return &VotingRepo{db: db}
}

func (r *VotingRepo) CreateVoting(voting *model.Voting) (*model.Voting, error) {
	if err := r.db.Create(&voting).Error; err != nil {
		return nil, err
	}
	return voting, nil
}

func (r *VotingRepo) UpdateVotingStatus(voting *model.Voting) (*model.Voting, error) {
	if err := r.db.Model(&voting).Update("status", voting.Status).Error; err != nil {
		return nil, err
	}
	return voting, nil
}

func (r *VotingRepo) FindVotingByID(id int64) (*model.Voting, error) {
	var voting model.Voting
	if err := r.db.First(&voting, id).Error; err != nil {
		return nil, err
	}
	return &voting, nil
}

func (r *VotingRepo) FindVotingsByStatus(status string) ([]*model.Voting, error) {
	var votings []*model.Voting
	if err := r.db.Where(&model.Voting{Status: status}).Find(&votings).Error; err != nil {
		return nil, err
	}
	return votings, nil
}
