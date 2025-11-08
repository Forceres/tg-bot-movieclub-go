package service

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
)

type IVotingService interface {
	CreateVoting(voting *model.Voting) (*model.Voting, error)
	FindVotingByID(id int64) (*model.Voting, error)
	UpdateVotingStatus(voting *model.Voting) (*model.Voting, error)
	FindVotingByStatus(status string) ([]*model.Voting, error)
}

type VotingService struct {
	repo repository.IVotingRepo
}

func NewVotingService(repo repository.IVotingRepo) *VotingService {
	return &VotingService{repo: repo}
}

func (s *VotingService) CreateVoting(voting *model.Voting) (*model.Voting, error) {
	return s.repo.CreateVoting(voting)
}

func (s *VotingService) FindVotingByID(id int64) (*model.Voting, error) {
	return s.repo.FindVotingByID(id)
}

func (s *VotingService) FindVotingByStatus(status string) ([]*model.Voting, error) {
	return s.repo.FindVotingsByStatus(status)
}

func (s *VotingService) UpdateVotingStatus(voting *model.Voting) (*model.Voting, error) {
	return s.repo.UpdateVotingStatus(voting)
}
