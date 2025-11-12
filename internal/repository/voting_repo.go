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

type CreateVotingParams struct {
	Voting *model.Voting
	Tx     *gorm.DB
}

type IVotingRepo interface {
	Transaction(txFunc func(tx *gorm.DB) error) error
	CreateVoting(params *CreateVotingParams) (*model.Voting, error)
	FindVotingByID(id int64) (*model.Voting, error)
	FindVotingsByStatus(status string) ([]*model.Voting, error)
	UpdateVotingStatus(voting *model.Voting) (*model.Voting, error)
	FinishVoting(params *FinishVotingParams) error
	FindVotingsBySessionID(sessionID int64) ([]*model.Voting, error)
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

func (r *VotingRepo) FindVotingsBySessionID(sessionID int64) ([]*model.Voting, error) {
	var votings []*model.Voting
	if err := r.db.Where(&model.Voting{SessionID: &sessionID}).Find(&votings).Error; err != nil {
		return nil, err
	}
	return votings, nil
}

func (r *VotingRepo) CreateVoting(params *CreateVotingParams) (*model.Voting, error) {
	var tx *gorm.DB = r.db
	if params.Tx != nil {
		tx = params.Tx
	}
	voting := params.Voting
	if err := tx.Create(voting).Error; err != nil {
		return nil, err
	}
	return voting, nil
}

func (r *VotingRepo) FinishVoting(params *FinishVotingParams) error {
	var tx *gorm.DB = r.db
	if params.Tx != nil {
		tx = params.Tx
	}
	updates := map[string]interface{}{
		"status":      model.VOTING_INACTIVE_STATUS,
		"finished_at": time.Now().Unix(),
	}
	err := tx.Model(&model.Voting{}).Where(&model.Voting{ID: params.VotingID}).Updates(updates).Error
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
