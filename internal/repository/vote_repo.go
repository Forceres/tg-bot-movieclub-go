package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type IVoteRepo interface {
	Create(vote *model.Vote) error
	FindVotesByVotingID(id int64) (*model.Vote, error)
}

type VoteRepo struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) *VoteRepo {
	return &VoteRepo{db: db}
}

func (r *VoteRepo) Create(vote *model.Vote) error {
	if err := r.db.Create(&vote).Error; err != nil {
		return err
	}
	return nil
}

func (r *VoteRepo) FindVotesByVotingID(id int64) (*model.Vote, error) {
	var vote model.Vote
	if err := r.db.First(&vote, "voting_id = ?", id).Error; err != nil {
		return nil, err
	}
	return &vote, nil
}
