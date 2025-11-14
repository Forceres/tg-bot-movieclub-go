package service

import (
	"context"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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

type VotingOptions struct {
	Title      string
	Type       string
	CreatedBy  int64
	FinishedAt *int64
	MovieID    *int
	SessionID  *int64
}

type StartRatingVotingParams struct {
	Bot         *bot.Bot
	Context     context.Context
	ChatID      int64
	Options     VotingOptions
	Question    string
	PollOptions []models.InputPollOption
}

type IVotingService interface {
	UpdateVotingStatus(voting *model.Voting) (*model.Voting, error)
	FindVotingByStatus(status string) ([]*model.Voting, error)
	FinishRatingVoting(params *FinishRatingVotingParams) error
	FinishSelectionVoting(params *FinishSelectionVotingParams) (*model.Session, bool, error)
	StartVoting(params *StartRatingVotingParams) (*model.Poll, error)
	FindVotingsBySessionID(sessionID int64) ([]*model.Voting, error)
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

func (s *VotingService) FindVotingsBySessionID(sessionID int64) ([]*model.Voting, error) {
	return s.repo.FindVotingsBySessionID(sessionID)
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

func (s *VotingService) FinishSelectionVoting(params *FinishSelectionVotingParams) (*model.Session, bool, error) {
	var created bool = false
	var session *model.Session
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
		finishedAt, err := s.scheduleService.GetNextScheduledTime()
		if err != nil {
			return err
		}
		session, created, err = s.sessionRepo.FindOrCreateSession(&repository.FindOrCreateSessionParams{
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
		return nil, created, err
	}
	return session, created, nil
}

func (s *VotingService) StartVoting(params *StartRatingVotingParams) (*model.Poll, error) {
	var poll *model.Poll
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		voting := &model.Voting{
			Title:      params.Options.Title,
			Type:       params.Options.Type,
			CreatedBy:  params.Options.CreatedBy,
			FinishedAt: params.Options.FinishedAt,
		}
		if params.Options.MovieID != nil {
			voting.MovieID = params.Options.MovieID
		}
		if params.Options.SessionID != nil {
			voting.SessionID = params.Options.SessionID
		}
		createdVoting, err := s.repo.CreateVoting(&repository.CreateVotingParams{
			Voting: voting,
			Tx:     tx,
		})
		if err != nil {
			params.Bot.SendMessage(params.Context, &bot.SendMessageParams{
				ChatID: params.ChatID,
				Text:   "Ошибка при создании голосования.",
			})
			return err
		}
		pollMsg, err := params.Bot.SendPoll(params.Context, &bot.SendPollParams{
			ChatID:            params.ChatID,
			Question:          params.Question,
			Options:           params.PollOptions,
			IsAnonymous:       bot.False(),
			Type:              "regular",
			QuestionParseMode: models.ParseModeMarkdown,
		})
		if err != nil {
			log.Printf("Error sending poll: %v", err)
			return err
		}

		pollModel := &model.Poll{
			PollID:    pollMsg.Poll.ID,
			MessageID: pollMsg.ID,
			VotingID:  createdVoting.ID,
			Type:      params.Options.Type,
			Status:    "active",
		}

		if params.Options.MovieID != nil {
			pollModel.MovieID = params.Options.MovieID
		}

		poll, err = s.pollRepo.Create(&repository.CreatePollParams{
			Poll: pollModel,
			Tx:   tx,
		})
		if err != nil {
			log.Printf("Error saving poll: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return poll, nil
}

func (s *VotingService) FindVotingByStatus(status string) ([]*model.Voting, error) {
	return s.repo.FindVotingsByStatus(status)
}

func (s *VotingService) UpdateVotingStatus(voting *model.Voting) (*model.Voting, error) {
	return s.repo.UpdateVotingStatus(voting)
}
