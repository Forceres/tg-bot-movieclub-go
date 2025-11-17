package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"github.com/go-telegram/bot"
	"github.com/goodsign/monday"
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

const ALREADY_WATCHED_MOVIES_FORMAT = `<p>
<b>#%d: %s (%d)</b>
<b>Режиссер(ы): %s.</b>
<b>Страны выпуска: %s.</b>
<b>Жанры: %s.</b>
<b>Длительность в минутах: %d.</b>
<b>Рейтинг IMDb: %f.</b>
<b>Рейтинг КиноКласса: %s.</b>
<i>Дата просмотра: %s.</i>
<i>Предложен: %s.</i>
<a href=%s><i>Ссылка</i></a>
</p>`

const ALREADY_WATCHED_MOVIES_PAGE_SIZE = 50

type IMovieService interface {
	GetCurrentMovies() (*string, error)
	GetAlreadyWatchedMovies() ([]string, error)
	GetSuggestedOrWatchedMovies(suggested bool) ([][]string, error)
	GetMovieByID(id int64) (*model.Movie, error)
	Create(movie *MovieDTO, suggestedBy int64) error
	Upsert(movie *MovieDTO, suggestedBy int64) error
	generateHTMLForWatchedMovies(movies []*model.Movie) []string
}

type MovieService struct {
	repo        repository.IMovieRepo
	sessionRepo repository.ISessionRepo
}

func NewMovieService(repo repository.IMovieRepo, sessionRepo repository.ISessionRepo) *MovieService {
	return &MovieService{repo: repo, sessionRepo: sessionRepo}
}

func (s *MovieService) Upsert(movie *MovieDTO, suggestedBy int64) error {
	suggestedAt := time.Now().Unix()
	newMovie := model.Movie{
		ID:          movie.KinopoiskID,
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
	return s.repo.Upsert(&newMovie)
}

func (s *MovieService) Create(movie *MovieDTO, suggestedBy int64) error {
	suggestedAt := time.Now().Unix()
	newMovie := model.Movie{
		ID:          movie.KinopoiskID,
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

func (s *MovieService) GetMovieByID(id int64) (*model.Movie, error) {
	movie, err := s.repo.GetMovieByID(id)
	if err != nil {
		return nil, err
	}
	return movie, nil
}

func (s *MovieService) GetCurrentMovies() (*string, error) {
	session, err := s.sessionRepo.FindOngoingSession()
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("no ongoing session found")
	}
	if len(session.Movies) == 0 {
		return nil, fmt.Errorf("no current movies found")
	}
	finishedAt := time.Unix(session.FinishedAt, 0)
	month := monday.Format(finishedAt, "January", monday.LocaleRuRU)
	day := monday.Format(finishedAt, "Monday", monday.LocaleRuRU)
	month = strings.ToUpper(month)
	day = strings.ToLower(day)
	dateStr := finishedAt.Format("02.01.2006")
	timeStr := finishedAt.Format("15:04")
	schedule := fmt.Sprintf("%s | %s 🗓️\n%s | %s 🕤", month, dateStr, day, timeStr)
	movies := session.Movies
	var offset int
	var formattedMovies []string
	if session.Description != "" {
		offset = 3
		formattedMovies = make([]string, len(movies)+offset)
		formattedMovies[0] = session.Description
		formattedMovies[1] = schedule
		formattedMovies[2] = "<b>#смотрим</b>"
	} else {
		offset = 2
		formattedMovies = make([]string, len(movies)+offset)
		formattedMovies[0] = schedule
		formattedMovies[1] = "<b>#смотрим</b>"
	}
	for i, movie := range movies {
		var suggestedBy string = "Неизвестно"
		if movie.Suggester != nil {
			suggestedBy = fmt.Sprintf("%s %s", movie.Suggester.FirstName, movie.Suggester.LastName)
		}
		formattedMovies[i+offset] = fmt.Sprintf(MOVIE_FORMAT,
			movie.ID,
			movie.Title,
			movie.Genres,
			movie.Countries,
			movie.IMDBRating,
			movie.Directors,
			movie.Year,
			movie.Duration,
			suggestedBy,
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
	var movies []*model.Movie
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
		movieString := fmt.Sprintf(`%d. %s (%d)`, i+1, movie.Title, movie.Year)
		if movie.Suggester != nil {
			movieString += fmt.Sprintf(" - предложил: %s %s", movie.Suggester.FirstName, movie.Suggester.LastName)
		}
		list[i] = []string{fmt.Sprint(movie.ID), bot.EscapeMarkdownUnescaped(movieString)}
	}
	return list, nil
}

func (s *MovieService) generateHTMLForWatchedMovies(movies []*model.Movie) []string {
	var pages []string
	var html strings.Builder

	for i, movie := range movies {
		var rating string = "N/A"
		if movie.Rating != 0 {
			rating = fmt.Sprintf("%.1f", movie.Rating)
		}
		var suggestedBy string = "Неизвестно"
		if movie.Suggester != nil {
			suggestedBy = fmt.Sprintf("%s %s", movie.Suggester.FirstName, movie.Suggester.LastName)
		}
		var finishedAt string = "N/A"
		if movie.FinishedAt != nil {
			tm, err := time.Parse("2006-01-02 15:04:05", *movie.FinishedAt)
			if err != nil {
				fmt.Println("Error parsing time:", err)
			} else {
				finishedAt = monday.Format(tm, "02 January 2006", monday.LocaleRuRU)
			}
		}
		html.WriteString(fmt.Sprintf(ALREADY_WATCHED_MOVIES_FORMAT, i+1, movie.Title, movie.Year, movie.Directors, movie.Countries, movie.Genres, movie.Duration, movie.IMDBRating, rating, finishedAt, suggestedBy, movie.Link))

		if (i+1)%ALREADY_WATCHED_MOVIES_PAGE_SIZE == 0 || i == len(movies)-1 {
			pages = append(pages, html.String())
			html.Reset()
		}
	}
	return pages
}
