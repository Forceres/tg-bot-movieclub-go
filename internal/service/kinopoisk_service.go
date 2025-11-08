package service

import (
	"fmt"

	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/kinopoisk"
)

type MovieDTO struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Directors   []string `json:"director"`
	Year        int      `json:"year"`
	Countries   []string `json:"countries"`
	Genres      []string `json:"genres"`
	Link        string   `json:"link"`
	Duration    int      `json:"duration"`
	Imdb        float64  `json:"imdb"`
	SuggestedBy string   `json:"suggested_by"`
}

type KinopoiskService struct {
	kinopoiskAPI kinopoisk.IKinopoiskAPI
}

type IKinopoiskService interface {
	SearchMovies(ids []int, suggestedBy string) ([]MovieDTO, error)
	ParseMovies(response *[]kinopoisk.KinopoiskMovieWithStaff, suggestedBy *string) ([]MovieDTO, error)
}

func NewKinopoiskService(kinopoiskAPI kinopoisk.IKinopoiskAPI) *KinopoiskService {
	return &KinopoiskService{
		kinopoiskAPI: kinopoiskAPI,
	}
}

func (s *KinopoiskService) SearchMovies(ids []int, suggestedBy string) ([]MovieDTO, error) {
	movies, err := s.kinopoiskAPI.SearchMovies(ids)
	if err != nil {
		return nil, err
	}

	return s.ParseMovies(movies, &suggestedBy)
}

func (s *KinopoiskService) ParseMovies(response *[]kinopoisk.KinopoiskMovieWithStaff, suggestedBy *string) ([]MovieDTO, error) {
	var moviesDto []MovieDTO
	for _, item := range *response {
		var movieDto MovieDTO
		movieDto.Link = fmt.Sprintf("https://www.kinopoisk.ru/film/%d/", item.Movie.KinopoiskID)
		for _, person := range *item.Staff {
			if person.ProfessionKey == "DIRECTOR" {
				movieDto.Directors = append(movieDto.Directors, person.NameRu)
			}
		}
		movieDto.Description = item.Movie.Description
		movieDto.Title = item.Movie.NameRu
		for _, country := range item.Movie.Countries {
			movieDto.Countries = append(movieDto.Countries, country.Country)
		}
		for _, genre := range item.Movie.Genres {
			movieDto.Genres = append(movieDto.Genres, genre.Genre)
		}
		movieDto.Year = item.Movie.Year
		movieDto.Duration = item.Movie.FilmLength
		movieDto.Imdb = item.Movie.RatingImdb
		movieDto.SuggestedBy = *suggestedBy
		moviesDto = append(moviesDto, movieDto)
	}
	return moviesDto, nil
}
