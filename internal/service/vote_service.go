package service

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
)

type IVoteService interface {
	Create(vote *model.Vote) error
	FindVotesByVotingID(id int64) (*model.Vote, error)
}

type VoteService struct {
	repo repository.IVoteRepo
}

func NewVoteService(repo repository.IVoteRepo) *VoteService {
	return &VoteService{repo: repo}
}

func (s *VoteService) Create(vote *model.Vote) error {
	return s.repo.Create(vote)
}

func (s *VoteService) FindVotesByVotingID(id int64) (*model.Vote, error) {
	return s.repo.FindVotesByVotingID(id)
}
