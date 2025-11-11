package service

import (
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"gorm.io/gorm"
)

type ISessionService interface {
	FinishSession(sessionID int64) error
}

type SessionService struct {
	repo      repository.ISessionRepo
	movieRepo repository.IMovieRepo
}

func NewSessionService(repo repository.ISessionRepo, movieRepo repository.IMovieRepo) ISessionService {
	return &SessionService{repo: repo, movieRepo: movieRepo}
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
			if err := s.movieRepo.Update(&repository.UpdateParams{Movie: &movie, Tx: tx}); err != nil {
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
