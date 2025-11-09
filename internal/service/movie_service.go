package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"github.com/go-telegram/bot"
)

const MOVIE_FORMAT = `
<b>%d. Фильм: %s.</b>
<i>Жанры: %s.</i>
<i>Страны производства: %s.</i>
<i>Рейтинг IMDb: %f.</i>
<i>Режиссер: %s.</i>
<i>Год: %d.</i>
<i>Длительность в минутах: %d.</i>
<i>Предложен: %d.</i>
<i>Ссылка на кинопоиск: %s.</i>
`

const ALREADY_WATCHED_MOVIES_FORMAT = `<p><b>#%d: %s (%d)</b>
<b>Режиссер: %s <br>Страны выпуска: %s</b>
<b>Жанры: %s</b>
<b>Длительность в минутах: %d</b>
<b>Рейтинг IMDb: %f</b>
<b>Рейтинг КиноКласса: %s</b>
<i>Дата просмотра: %s</i>
<i>Предложен: %s</i>
<a href=%s><i>Ссылка</i></a>
</p>`

const ALREADY_WATCHED_MOVIES_PAGE_SIZE = 50

type IMovieService interface {
	GetCurrentMovies() (*string, error)
	GetAlreadyWatchedMovies() ([]string, error)
	GetSuggestedOrWatchedMovies(suggested bool) ([][]string, error)
	GetMovieByID(id int) (*model.Movie, error)
	Create(movie *MovieDTO, suggestedBy int64) error
	generateHTMLForWatchedMovies(movies []model.Movie) []string
}

type MovieService struct {
	repo repository.IMovieRepo
}

func NewMovieService(repo repository.IMovieRepo) *MovieService {
	return &MovieService{repo: repo}
}

func (s *MovieService) Create(movie *MovieDTO, suggestedBy int64) error {
	suggestedAt := time.Now().Unix()
	newMovie := model.Movie{
		ID:          0,
		Title:       movie.Title,
		Description: movie.Description,
		Directors:   strings.Join(movie.Directors, ", "),
		Year:        movie.Year,
		Countries:   strings.Join(movie.Countries, ", "),
		Genres:      strings.Join(movie.Genres, ", "),
		Link:        movie.Link,
		Duration:    movie.Duration,
		IMDBRating:  movie.IMDBRating,
		SuggestedBy: &suggestedBy,
		SuggestedAt: &suggestedAt,
	}
	return s.repo.Create(&newMovie)
}

func (s *MovieService) GetMovieByID(id int) (*model.Movie, error) {
	movie, err := s.repo.GetMovieByID(id)
	if err != nil {
		return nil, err
	}
	return movie, nil
}

func (s *MovieService) GetCurrentMovies() (*string, error) {
	movies, err := s.repo.GetCurrentMovies()
	if err != nil {
		return nil, err
	}
	if len(movies) == 0 {
		return nil, fmt.Errorf("no current movies found")
	}
	formattedMovies := make([]string, len(movies)+1)
	formattedMovies[0] = "<b>#смотрим</b>"
	for i, movie := range movies {
		formattedMovies[i+1] = fmt.Sprintf(MOVIE_FORMAT,
			movie.ID,
			movie.Title,
			movie.Genres,
			movie.Countries,
			movie.IMDBRating,
			movie.Directors,
			movie.Year,
			movie.Duration,
			movie.SuggestedBy,
			movie.Link,
		)
	}
	result := strings.Join(formattedMovies, "\n")
	return &result, nil
}

func (s *MovieService) GetAlreadyWatchedMovies() ([]string, error) {
	movies, err := s.repo.GetAlreadyWatchedMovies()
	if err != nil {
		return nil, err
	}
	result := s.generateHTMLForWatchedMovies(movies)
	return result, nil
}

func (s *MovieService) GetSuggestedOrWatchedMovies(suggested bool) ([][]string, error) {
	var movies []model.Movie
	var err error

	if suggested {
		movies, err = s.repo.GetSuggestedMovies()
		if err != nil {
			return nil, err
		}
	} else {
		movies, err = s.repo.GetAlreadyWatchedMovies()
		if err != nil {
			return nil, err
		}
	}
	list := make([][]string, len(movies))
	for i, movie := range movies {
		movieString := fmt.Sprintf(`%d. %s (%d)`, movie.ID, movie.Title, movie.Year)
		if movie.SuggestedBy != nil {
			movieString += fmt.Sprintf(" - предложил: %d", *movie.SuggestedBy)
		}
		list[i] = []string{fmt.Sprint(movie.ID), bot.EscapeMarkdown(movieString)}
	}
	return list, nil
}

func (s *MovieService) generateHTMLForWatchedMovies(movies []model.Movie) []string {
	var pages []string
	var html strings.Builder

	for i, movie := range movies {
		var rating string = "N/A"
		if movie.Rating != 0 {
			rating = fmt.Sprintf("%.1f", movie.Rating)
		}
		var suggestedBy string = "Неизвестно"
		if movie.SuggestedBy != nil {
			suggestedBy = fmt.Sprintf("%d", *movie.SuggestedBy)
		}
		html.WriteString(fmt.Sprintf(ALREADY_WATCHED_MOVIES_FORMAT, i+1, movie.Title, movie.Year, movie.Directors, movie.Countries, movie.Genres, movie.Duration, movie.IMDBRating, rating, movie.StartedAt, suggestedBy, movie.Link))

		if (i+1)%ALREADY_WATCHED_MOVIES_PAGE_SIZE == 0 || i == len(movies)-1 {
			pages = append(pages, html.String())
			html.Reset()
		}
	}
	return pages
}
