package repository

import (
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type FinishVotingParams struct {
	VotingID int64
	Tx       *gorm.DB
}

type IVotingRepo interface {
	Transaction(txFunc func(tx *gorm.DB) error) error
	CreateVoting(voting *model.Voting) (*model.Voting, error)
	FindVotingByID(id int64) (*model.Voting, error)
	FindVotingsByStatus(status string) ([]*model.Voting, error)
	UpdateVotingStatus(voting *model.Voting) (*model.Voting, error)
	FinishVoting(params *FinishVotingParams) error
}

type VotingRepo struct {
	db        *gorm.DB
	pollRepo  IPollRepo
	movieRepo IMovieRepo
}

func NewVotingRepository(db *gorm.DB, pollRepo IPollRepo, movieRepo IMovieRepo) *VotingRepo {
	return &VotingRepo{db: db, pollRepo: pollRepo, movieRepo: movieRepo}
}

func (r *VotingRepo) Transaction(txFunc func(tx *gorm.DB) error) error {
	return r.db.Transaction(txFunc)
}

func (r *VotingRepo) CreateVoting(voting *model.Voting) (*model.Voting, error) {
	if err := r.db.Create(&voting).Error; err != nil {
		return nil, err
	}
	return voting, nil
}

func (r *VotingRepo) FinishVoting(params *FinishVotingParams) error {
	var tx *gorm.DB = r.db
	if params.Tx != nil {
		tx = params.Tx
	}
	err := tx.Model(&model.Voting{ID: params.VotingID}).Update("status", "inactive").Update("finished_at", time.Now().Unix()).Error
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
