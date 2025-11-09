package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type IVoteRepo interface {
	Create(vote *model.Vote) error
	FindVotesByVotingID(id int64) (*model.Vote, error)
	CalculateRatingMean(votingID int64) (float64, error)
	CalculateMaxMovieCount(votingID int64) (int64, int, error)
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

func (r *VoteRepo) CalculateMaxMovieCount(votingID int64) (int64, int, error) {
	var result struct {
		MaxCount int64
		MovieID  int
	}
	err := r.db.Model(&model.Vote{}).
		Select("MAX(movie_count) as max_count, movie_id").
		Where("voting_id = ?", votingID).
		Group("voting_id").
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.MaxCount, result.MovieID, nil
}

func (r *VoteRepo) CalculateRatingMean(votingID int64) (float64, error) {
	var result struct {
		Mean float64
	}
	err := r.db.Model(&model.Vote{}).
		Select("AVG(rating) as mean").
		Where("voting_id = ?", votingID).
		Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return result.Mean, nil
}

func (r *VoteRepo) FindVotesByVotingID(id int64) (*model.Vote, error) {
	var vote model.Vote
	if err := r.db.First(&vote, "voting_id = ?", id).Error; err != nil {
		return nil, err
	}
	return &vote, nil
}
