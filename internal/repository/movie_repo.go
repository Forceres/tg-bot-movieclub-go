package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type UpdateRatingParams struct {
	MovieID int
	Rating  float64
	Tx      *gorm.DB
}

type UpdateDatesParams struct {
	MovieID    int
	StartedAt  string
	FinishedAt string
	Tx         *gorm.DB
}

type IMovieRepo interface {
	GetCurrentMovies() ([]model.Movie, error)
	GetAlreadyWatchedMovies() ([]model.Movie, error)
	GetSuggestedMovies() ([]model.Movie, error)
	GetMovieByID(id int) (*model.Movie, error)
	Create(movie *model.Movie) error
	UpdateRating(params *UpdateRatingParams) error
	UpdateDates(params *UpdateDatesParams) error
}

type MovieRepo struct {
	db *gorm.DB
}

func NewMovieRepository(db *gorm.DB) *MovieRepo {
	return &MovieRepo{db: db}
}

func (r *MovieRepo) Create(movie *model.Movie) error {
	return r.db.Create(movie).Error
}

func (r *MovieRepo) UpdateRating(params *UpdateRatingParams) error {
	tx := params.Tx
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&model.Movie{ID: params.MovieID}).Update("rating", params.Rating).Error
}

func (r *MovieRepo) UpdateDates(params *UpdateDatesParams) error {
	tx := params.Tx
	if tx == nil {
		tx = r.db
	}
	updates := map[string]string{}
	if params.StartedAt != "" {
		updates["started_at"] = params.StartedAt
	}
	if params.FinishedAt != "" {
		updates["finished_at"] = params.FinishedAt
	}
	return tx.Model(&model.Movie{ID: params.MovieID}).Updates(updates).Error
}

func (r *MovieRepo) GetMovieByID(id int) (*model.Movie, error) {
	var movie model.Movie
	if err := r.db.Model(&model.Movie{}).Where(&model.Movie{ID: id}).First(&movie).Error; err != nil {
		return nil, err
	}
	return &movie, nil
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
	if err := r.db.Model(&model.Movie{}).Where(r.db.Not(&model.Movie{FinishedAt: ""})).Find(&movies).Error; err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *MovieRepo) GetSuggestedMovies() ([]model.Movie, error) {
	var movies []model.Movie
	if err := r.db.Model(&model.Movie{}).Where(&model.Movie{SuggestedAt: nil}).Find(&movies).Error; err != nil {
		return nil, err
	}
	return movies, nil
}
