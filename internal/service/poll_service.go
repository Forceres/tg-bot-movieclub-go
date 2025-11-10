package service

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
)

type IPollService interface {
	CreatePoll(poll *model.Poll) (*model.Poll, error)
	CreatePollOption(option *model.PollOption) error
	GetPollByPollID(pollID string) (*model.Poll, error)
	GetPollOptionsByPollID(pollID int64) ([]model.PollOption, error)
	GetActivePolls() ([]model.Poll, error)
	ClosePoll(pollID string) error
}

type PollService struct {
	pollRepo repository.IPollRepo
}

func NewPollService(pollRepo repository.IPollRepo) IPollService {
	return &PollService{pollRepo: pollRepo}
}

func (s *PollService) CreatePoll(poll *model.Poll) (*model.Poll, error) {
	return s.pollRepo.Create(&repository.CreatePollParams{Poll: poll})
}

func (s *PollService) CreatePollOption(option *model.PollOption) error {
	return s.pollRepo.CreatePollOption(option)
}

func (s *PollService) GetPollByPollID(pollID string) (*model.Poll, error) {
	return s.pollRepo.FindByPollID(pollID)
}

func (s *PollService) GetPollOptionsByPollID(pollID int64) ([]model.PollOption, error) {
	return s.pollRepo.FindPollOptionsByPollID(pollID)
}

func (s *PollService) GetActivePolls() ([]model.Poll, error) {
	return s.pollRepo.FindActivePolls()
}

func (s *PollService) ClosePoll(pollID string) error {
	return s.pollRepo.UpdateStatus(&repository.UpdateStatusParams{
		PollID: pollID,
		Status: "closed",
	})
}
