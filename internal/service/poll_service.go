package service

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
)

type IPollService interface {
	CreatePollOption(option *model.PollOption) error
	GetPollByPollID(pollID string) (*model.Poll, error)
	GetOpenedPollByMovieID(movieID int64) (*model.Poll, error)
	GetPollOptionsByPollID(pollID int64) ([]*model.PollOption, error)
}

type PollService struct {
	pollRepo repository.IPollRepo
}

func NewPollService(pollRepo repository.IPollRepo) IPollService {
	return &PollService{pollRepo: pollRepo}
}

func (s *PollService) CreatePollOption(option *model.PollOption) error {
	return s.pollRepo.CreatePollOption(option)
}

func (s *PollService) GetPollByPollID(pollID string) (*model.Poll, error) {
	return s.pollRepo.FindByPollID(pollID)
}

func (s *PollService) GetOpenedPollByMovieID(movieID int64) (*model.Poll, error) {
	return s.pollRepo.FindOpenedByMovieID(movieID)
}

func (s *PollService) GetPollOptionsByPollID(pollID int64) ([]*model.PollOption, error) {
	return s.pollRepo.FindPollOptionsByPollID(pollID)
}
