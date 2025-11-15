package service

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"gorm.io/gorm"
)

type IVoteService interface {
	Create(vote *model.Vote) error
	CalculateRatingMean(votingID int64) (float64, error)
	CalculateMaxMovieCount(votingID int64) (int64, int64, error)
}

type VoteService struct {
	repo repository.IVoteRepo
}

func NewVoteService(repo repository.IVoteRepo) *VoteService {
	return &VoteService{repo: repo}
}

func (s *VoteService) CalculateRatingMean(votingID int64) (float64, error) {
	return s.repo.CalculateRatingMean(votingID)
}

func (s *VoteService) CalculateMaxMovieCount(votingID int64) (int64, int64, error) {
	return s.repo.CalculateMaxMovieCount(votingID)
}

func (s *VoteService) Create(vote *model.Vote) error {
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		deleteParams := &repository.DeleteByUserIdAndVotingIdParams{
			UserID:   vote.UserID,
			VotingID: vote.VotingID,
			Tx:       tx,
		}
		if err := s.repo.DeleteByUserIdAndVotingId(deleteParams); err != nil {
			return err
		}
		createParams := &repository.CreateVoteParams{
			Vote: vote,
			Tx:   tx,
		}
		if err := s.repo.Create(createParams); err != nil {
			return err
		}
		return nil
	})
	return err
}
