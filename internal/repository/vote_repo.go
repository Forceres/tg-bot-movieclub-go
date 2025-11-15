package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type DeleteByUserIdAndVotingIdParams struct {
	UserID   int64
	VotingID int64
	Tx       *gorm.DB
}

type CreateVoteParams struct {
	Vote *model.Vote
	Tx   *gorm.DB
}

type IVoteRepo interface {
	Create(params *CreateVoteParams) error
	DeleteByUserIdAndVotingId(params *DeleteByUserIdAndVotingIdParams) error
	CalculateRatingMean(votingID int64) (float64, error)
	CalculateMaxMovieCount(votingID int64) (int64, int64, error)
	Transaction(func(tx *gorm.DB) error) error
}

type VoteRepo struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) IVoteRepo {
	return &VoteRepo{db: db}
}

func (r *VoteRepo) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *VoteRepo) DeleteByUserIdAndVotingId(params *DeleteByUserIdAndVotingIdParams) error {
	var tx *gorm.DB = r.db
	if params.Tx != nil {
		tx = params.Tx
	}
	return tx.Where(&model.Vote{UserID: params.UserID, VotingID: params.VotingID}).Delete(&model.Vote{}).Error
}

func (r *VoteRepo) Create(params *CreateVoteParams) error {
	var tx *gorm.DB = r.db
	if params.Tx != nil {
		tx = params.Tx
	}
	if err := tx.Create(params.Vote).Error; err != nil {
		return err
	}
	return nil
}

func (r *VoteRepo) CalculateMaxMovieCount(votingID int64) (int64, int64, error) {
	var result struct {
		MovieCount int64
		MovieID    int64
	}
	err := r.db.Model(&model.Vote{}).
		Select("COUNT(*) as movie_count, movie_id").
		Where("voting_id = ?", votingID).
		Group("movie_id").
		Order("movie_count DESC").
		Limit(1).
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.MovieCount, result.MovieID, nil
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
