package service

import (
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"gorm.io/gorm"
)

type FinishSelectionVotingParams struct {
	VotingID  int64
	PollID    string
	MovieID   int
	CreatedBy int64
}

type FinishRatingVotingParams struct {
	VotingID  int64
	PollID    string
	MovieID   int
	Mean      float64
	CreatedBy int64
}

type IVotingService interface {
	CreateVoting(voting *model.Voting) (*model.Voting, error)
	FindVotingByID(id int64) (*model.Voting, error)
	UpdateVotingStatus(voting *model.Voting) (*model.Voting, error)
	FindVotingByStatus(status string) ([]*model.Voting, error)
	FinishRatingVoting(params *FinishRatingVotingParams) error
	FinishSelectionVoting(params *FinishSelectionVotingParams) (int64, error)
}

type VotingService struct {
	repo            repository.IVotingRepo
	sessionRepo     repository.ISessionRepo
	movieRepo       repository.IMovieRepo
	pollRepo        repository.IPollRepo
	scheduleService IScheduleService
}

func NewVotingService(repo repository.IVotingRepo, scheduleService IScheduleService, sessionRepo repository.ISessionRepo, movieRepo repository.IMovieRepo, pollRepo repository.IPollRepo) *VotingService {
	return &VotingService{repo: repo, scheduleService: scheduleService, sessionRepo: sessionRepo, movieRepo: movieRepo, pollRepo: pollRepo}
}

func (s *VotingService) CreateVoting(voting *model.Voting) (*model.Voting, error) {
	return s.repo.CreateVoting(voting)
}

func (s *VotingService) FindVotingByID(id int64) (*model.Voting, error) {
	return s.repo.FindVotingByID(id)
}

func (s *VotingService) FinishRatingVoting(params *FinishRatingVotingParams) error {
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		err := s.repo.FinishVoting(&repository.FinishVotingParams{
			VotingID: params.VotingID,
			Tx:       tx,
		})
		if err != nil {
			return err
		}
		err = s.pollRepo.UpdateStatus(&repository.UpdateStatusParams{
			PollID: params.PollID,
			Status: "closed",
			Tx:     tx,
		})
		if err != nil {
			return err
		}
		err = s.movieRepo.UpdateRating(&repository.UpdateRatingParams{
			MovieID: params.MovieID,
			Rating:  params.Mean,
			Tx:      tx,
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

func (s *VotingService) FinishSelectionVoting(params *FinishSelectionVotingParams) (int64, error) {
	var finishedAt int64
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		err := s.repo.FinishVoting(&repository.FinishVotingParams{
			VotingID: params.VotingID,
			Tx:       tx,
		})
		if err != nil {
			return err
		}
		err = s.pollRepo.UpdateStatus(&repository.UpdateStatusParams{
			PollID: params.PollID,
			Status: "closed",
			Tx:     tx,
		})
		if err != nil {
			return err
		}
		err = s.movieRepo.UpdateDates(&repository.UpdateDatesParams{
			MovieID:   params.MovieID,
			StartedAt: time.Now().String(),
			Tx:        tx,
		})
		if err != nil {
			return err
		}
		finishedAt, err = s.scheduleService.GetNextScheduledTime()
		if err != nil {
			return err
		}
		session, err := s.sessionRepo.FindOrCreateSession(&repository.FindOrCreateSessionParams{
			CreatedBy:  params.CreatedBy,
			FinishedAt: &finishedAt,
			Tx:         tx,
		})
		if err != nil {
			return err
		}
		err = s.sessionRepo.ConnectMoviesToSession(&repository.ConnectMoviesToSessionParams{
			SessionID: session.ID,
			MovieIDs:  []int{params.MovieID},
			Tx:        tx,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return finishedAt, nil
}

func (s *VotingService) FindVotingByStatus(status string) ([]*model.Voting, error) {
	return s.repo.FindVotingsByStatus(status)
}

func (s *VotingService) UpdateVotingStatus(voting *model.Voting) (*model.Voting, error) {
	return s.repo.UpdateVotingStatus(voting)
}
