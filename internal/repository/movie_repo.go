package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UpdateRatingParams struct {
	MovieID int64
	Rating  float64
	Tx      *gorm.DB
}

type UpdateDatesParams struct {
	MovieID   int64
	StartedAt *string
	Tx        *gorm.DB
}

type UpdateParams struct {
	Movie *model.Movie
	Tx    *gorm.DB
}

type IMovieRepo interface {
	GetCurrentMovies() ([]*model.Movie, error)
	GetAlreadyWatchedMovies() ([]*model.Movie, error)
	GetSuggestedMovies() ([]*model.Movie, error)
	GetMovieByID(id int64) (*model.Movie, error)
	Create(movie *model.Movie) error
	Update(params *UpdateParams) error
	UpdateRating(params *UpdateRatingParams) error
	Upsert(movie *model.Movie) error
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

func (r *MovieRepo) Upsert(movie *model.Movie) error {
	return r.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&movie).Error
}

func (r *MovieRepo) Update(params *UpdateParams) error {
	tx := params.Tx
	if tx == nil {
		tx = r.db
	}
	return tx.Save(params.Movie).Error
}

func (r *MovieRepo) UpdateRating(params *UpdateRatingParams) error {
	tx := params.Tx
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&model.Movie{ID: params.MovieID}).Update("rating", params.Rating).Error
}

func (r *MovieRepo) GetMovieByID(id int64) (*model.Movie, error) {
	var movie model.Movie
	if err := r.db.Model(&model.Movie{}).Where(&model.Movie{ID: id}).First(&movie).Error; err != nil {
		return nil, err
	}
	return &movie, nil
}

func (r *MovieRepo) GetCurrentMovies() ([]*model.Movie, error) {
	var movies []*model.Movie

	sub := r.db.Model(&model.Session{}).
		Select("movies_sessions.movie_id").
		Joins("JOIN movies_sessions ON movies_sessions.session_id = sessions.id").
		Where(&model.Session{Status: model.SESSION_ONGOING_STATUS})

	if err := r.db.Model(&model.Movie{}).Where("id IN (?)", sub).Find(&movies).Error; err != nil {
		return nil, err
	}

	return movies, nil
}

func (r *MovieRepo) GetAlreadyWatchedMovies() ([]*model.Movie, error) {
	var movies []*model.Movie
	if err := r.db.Model(&model.Movie{}).Where("watch_count > 0").Find(&movies).Error; err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *MovieRepo) GetSuggestedMovies() ([]*model.Movie, error) {
	var movies []*model.Movie
	if err := r.db.Model(&model.Movie{}).Where(&model.Movie{Status: model.MOVIE_SUGGESTED_STATUS}).Find(&movies).Error; err != nil {
		return nil, err
	}
	return movies, nil
}
