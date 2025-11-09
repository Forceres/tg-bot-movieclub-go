package repository

import (
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type FinishRatingVotingParams struct {
	VotingID int64
	PollID   string
	MovieID  int
	Mean     float64
	Tx       *gorm.DB
}

type FinishSelectionVotingParams struct {
	VotingID   int64
	PollID     string
	MovieID    int
	FinishedAt string
	Tx         *gorm.DB
}

type IVotingRepo interface {
	CreateVoting(voting *model.Voting) (*model.Voting, error)
	FindVotingByID(id int64) (*model.Voting, error)
	FindVotingsByStatus(status string) ([]*model.Voting, error)
	UpdateVotingStatus(voting *model.Voting) (*model.Voting, error)
	FinishRatingVoting(params *FinishRatingVotingParams) error
}

type VotingRepo struct {
	db        *gorm.DB
	pollRepo  IPollRepo
	movieRepo IMovieRepo
}

func NewVotingRepository(db *gorm.DB, pollRepo IPollRepo, movieRepo IMovieRepo) *VotingRepo {
	return &VotingRepo{db: db, pollRepo: pollRepo, movieRepo: movieRepo}
}

func (r *VotingRepo) CreateVoting(voting *model.Voting) (*model.Voting, error) {
	if err := r.db.Create(&voting).Error; err != nil {
		return nil, err
	}
	return voting, nil
}

func (r *VotingRepo) FinishRatingVoting(params *FinishRatingVotingParams) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.Voting{ID: params.VotingID}).Update("status", "inactive").Update("finished_at", time.Now().Unix()).Error
		if err != nil {
			return err
		}
		err = r.pollRepo.UpdateStatus(&UpdateStatusParams{
			PollID: params.PollID,
			Status: "closed",
			Tx:     tx,
		})
		if err != nil {
			return err
		}
		err = r.movieRepo.UpdateRating(&UpdateRatingParams{
			MovieID: params.MovieID,
			Rating:  params.Mean,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *VotingRepo) FinishSelectionVoting(params *FinishSelectionVotingParams) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.Voting{ID: params.VotingID}).Update("status", "inactive").Update("finished_at", time.Now().Unix()).Error
		if err != nil {
			return err
		}
		err = r.pollRepo.UpdateStatus(&UpdateStatusParams{
			PollID: params.PollID,
			Status: "closed",
			Tx:     tx,
		})
		if err != nil {
			return err
		}
		// err = r.movieRepo.UpdateDates(&UpdateDatesParams{
		// 	MovieID:    params.MovieID,
		// 	StartedAt:  time.Now().String(),
		// 	FinishedAt: params.FinishedAt,
		// 	Tx:         tx,
		// })
		// if err != nil {
		// 	return err
		// }
		return nil
	})
	if err != nil {
		return err
	}
	return nil
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
