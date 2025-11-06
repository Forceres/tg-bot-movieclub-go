package service

import (
	"fmt"
	"strings"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
)

const MOVIE_FORMAT = `
<b>%d. Фильм: %s.</b>
<i>Жанры: %s.</i>
<i>Страны производства: %s.</i>
<i>Рейтинг IMDb: %f.</i>
<i>Режиссер: %s.</i>
<i>Год: %d.</i>
<i>Длительность в минутах: %d.</i>
<i>Предложен: %s.</i>
<i>Ссылка на кинопоиск: %s.</i>
`

type IMovieService interface {
	GetCurrentMovies() (*string, error)
	GetAlreadyWatchedMovies() ([]model.Movie, error)
}

type MovieService struct {
	repo repository.IMovieRepo
}

func NewMovieService(repo repository.IMovieRepo) *MovieService {
	return &MovieService{repo: repo}
}

func (s *MovieService) GetCurrentMovies() (*string, error) {
	movies, err := s.repo.GetCurrentMovies()
	if err != nil {
		return nil, err
	}
	if len(movies) == 0 {
		return nil, fmt.Errorf("no current movies found")
	}
	formattedMovies := make([]string, len(movies) + 1)
	formattedMovies[0] = "<b>#смотрим</b>"
	for i, movie := range movies {
		formattedMovies[i+1] = fmt.Sprintf(MOVIE_FORMAT,
			movie.ID,
			movie.Title,
			movie.Genres,
			movie.Countries,
			movie.IMDBRating,
			movie.Director,
			movie.Year,
			movie.Duration,
			movie.SuggestedBy,
			movie.Link,
		)
	}
	result := strings.Join(formattedMovies, "\n")
	return &result, nil
}

func (s *MovieService) GetAlreadyWatchedMovies() ([]model.Movie, error) {
	movies, err := s.repo.GetAlreadyWatchedMovies()
	if err != nil {
		return nil, err
	}
	return movies, nil
}