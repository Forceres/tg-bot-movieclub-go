package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IVoteRepo interface {
	Create(vote *model.Vote) error
	Upsert(vote *model.Vote) error
	CalculateRatingMean(votingID int64) (float64, error)
	CalculateMaxMovieCount(votingID int64) (int64, int64, error)
}

type VoteRepo struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) *VoteRepo {
	return &VoteRepo{db: db}
}

func (r *VoteRepo) Upsert(vote *model.Vote) error {
	return r.db.Clauses(
		clause.OnConflict{
			UpdateAll: true,
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "voting_id"}},
		},
	).Create(&vote).Error
}

func (r *VoteRepo) Create(vote *model.Vote) error {
	if err := r.db.Create(&vote).Error; err != nil {
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
