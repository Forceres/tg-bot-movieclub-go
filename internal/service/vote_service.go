package service

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
)

type IVoteService interface {
	Create(vote *model.Vote) error
	CalculateRatingMean(votingID int64) (float64, error)
	CalculateMaxMovieCount(votingID int64) (int64, int, error)
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

func (s *VoteService) CalculateMaxMovieCount(votingID int64) (int64, int, error) {
	return s.repo.CalculateMaxMovieCount(votingID)
}

func (s *VoteService) Create(vote *model.Vote) error {
	return s.repo.Create(vote)
}
