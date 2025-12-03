package service

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"gorm.io/gorm"
)

type IVoteService interface {
	CreateMultiple(votingID int64, userID int64, votes []*model.Vote) error
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

func (s *VoteService) CreateMultiple(votingID int64, userID int64, votes []*model.Vote) error {
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		deleteParams := &repository.DeleteByUserIdAndVotingIdParams{
			UserID:   userID,
			VotingID: votingID,
			Tx:       tx,
		}
		if err := s.repo.DeleteByUserIdAndVotingId(deleteParams); err != nil {
			return err
		}

		for _, vote := range votes {
			createParams := &repository.CreateVoteParams{
				Vote: vote,
				Tx:   tx,
			}
			if err := s.repo.Create(createParams); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
