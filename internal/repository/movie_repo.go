package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type IMovieRepo interface {
	GetCurrentMovies() ([]model.Movie, error)
	GetAlreadyWatchedMovies() ([]model.Movie, error)
}

type MovieRepo struct {
	db *gorm.DB
}

func NewMovieRepository(db *gorm.DB) *MovieRepo {
	return &MovieRepo{db: db}
}

func (r *MovieRepo) GetCurrentMovies() ([]model.Movie, error) {
	var movies []model.Movie
	
	sub := r.db.Model(&model.Session{}).
		Select("movies_sessions.session_id").
		Joins("JOIN movies_sessions ON movies_sessions.session_id = sessions.id").
		Where(&model.Session{Status: "ACTIVE"})

	if err := r.db.Model(&model.Movie{}).Where("id IN (?)", sub).Find(&movies).Error; err != nil {
		return nil, err
	}
	
	return movies, nil
}

func (r *MovieRepo) GetAlreadyWatchedMovies() ([]model.Movie, error) {
	var movies []model.Movie
	if err := r.db.Model(&model.Movie{}).Where(r.db.Not(&model.Movie{FinishedAt: "NULL"})).Find(&movies).Error; err != nil {
		return nil, err
	}
	return movies, nil
}