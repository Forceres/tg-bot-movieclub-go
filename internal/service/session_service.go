package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"gorm.io/gorm"
)

type ISessionService interface {
	FinishSession(sessionID int64) error
	CancelSession() (*model.Session, error)
	AddMoviesToSession(createdBy int64, movieIDs []int) (*model.Session, []int, bool, error)
	FindOngoingSession() (*model.Session, error)
	RescheduleSession(sessionID int64, finishedAt int64) error
	FindOrCreateSession(finishedAt int64, createdAt int64) (*model.Session, bool, error)
}

type SessionService struct {
	repo            repository.ISessionRepo
	movieRepo       repository.IMovieRepo
	scheduleService IScheduleService
}

func NewSessionService(repo repository.ISessionRepo, movieRepo repository.IMovieRepo, scheduleService IScheduleService) ISessionService {
	return &SessionService{repo: repo, movieRepo: movieRepo, scheduleService: scheduleService}
}

func (s *SessionService) FindOrCreateSession(finishedAt int64, createdAt int64) (*model.Session, bool, error) {
	return s.repo.FindOrCreateSession(&repository.FindOrCreateSessionParams{
		CreatedBy:  createdAt,
		FinishedAt: &finishedAt,
	})
}

func (s *SessionService) RescheduleSession(sessionID int64, finishedAt int64) error {
	return s.repo.RescheduleSession(sessionID, finishedAt)
}

func (s *SessionService) CancelSession() (*model.Session, error) {
	return s.repo.CancelSession()
}

func (s *SessionService) FindOngoingSession() (*model.Session, error) {
	return s.repo.FindOngoingSession()
}

func (s *SessionService) FinishSession(sessionID int64) error {
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		session, err := s.repo.FinishSession(&repository.FinishSessionParams{
			SessionID: int64(sessionID),
			Tx:        tx,
		})
		if err != nil {
			return err
		}
		movies, err := s.movieRepo.GetCurrentMovies()
		if err != nil {
			return err
		}
		finishedAt := time.Unix(session.FinishedAt, 0).String()
		for _, movie := range movies {
			movie.WatchCount += 1
			movie.FinishedAt = &finishedAt
			if err := s.movieRepo.Update(&repository.UpdateParams{Movie: movie, Tx: tx}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *SessionService) AddMoviesToSession(createdBy int64, movieIDs []int) (*model.Session, []int, bool, error) {
	if len(movieIDs) == 0 {
		return nil, nil, false, fmt.Errorf("movieIDs cannot be empty")
	}
	unique := uniqueInts(movieIDs)
	var session *model.Session
	var newMovieIDs []int
	var sessionCreated bool
	err := s.repo.Transaction(func(tx *gorm.DB) error {
		var err error
		session, err = s.repo.GetOngoingSession(tx)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if s.scheduleService == nil {
				return fmt.Errorf("schedule service is not configured")
			}
			nextFinishedAt, schErr := s.scheduleService.GetNextScheduledTime()
			if schErr != nil {
				return schErr
			}
			if nextFinishedAt == 0 {
				return fmt.Errorf("no active schedule configured for session creation")
			}
			session, err = s.repo.Create(&repository.CreateSessionParams{
				Session: &model.Session{
					Status:     model.SESSION_ONGOING_STATUS,
					CreatedBy:  createdBy,
					FinishedAt: nextFinishedAt,
				},
				Tx: tx,
			})
			if err != nil {
				return err
			}
			sessionCreated = true
		}
		var existingMovies []model.Movie
		if err := tx.Model(session).Association("Movies").Find(&existingMovies); err != nil {
			return err
		}
		existing := make(map[int]struct{}, len(existingMovies))
		for _, movie := range existingMovies {
			existing[movie.ID] = struct{}{}
		}
		var moviesToAttach []model.Movie
		for _, id := range unique {
			if _, ok := existing[id]; ok {
				continue
			}
			var movie model.Movie
			if err := tx.First(&movie, id).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			moviesToAttach = append(moviesToAttach, movie)
			newMovieIDs = append(newMovieIDs, id)
		}
		if len(moviesToAttach) == 0 {
			return nil
		}
		if err := tx.Model(session).Association("Movies").Append(moviesToAttach); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, false, err
	}
	return session, newMovieIDs, sessionCreated, nil
}

func uniqueInts(values []int) []int {
	seen := make(map[int]struct{}, len(values))
	result := make([]int, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
