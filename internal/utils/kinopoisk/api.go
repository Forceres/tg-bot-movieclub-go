package kinopoisk

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
)

const MOVIES = "/films"
const STAFF = "/staff"

type Country struct {
	Country string `json:"country"`
}

type Genre struct {
	Genre string `json:"genre"`
}

type KinopoiskMovie struct {
	KinopoiskID                int64     `json:"kinopoiskId"`
	KinopoiskHDID              string    `json:"kinopoiskHDId"`
	ImdbID                     string    `json:"imdbId"`
	NameRu                     *string   `json:"nameRu"`
	NameEn                     string    `json:"nameEn"`
	NameOriginal               string    `json:"nameOriginal"`
	PosterURL                  string    `json:"posterUrl"`
	PosterURLPreview           string    `json:"posterUrlPreview"`
	CoverURL                   string    `json:"coverUrl"`
	LogoURL                    string    `json:"logoUrl"`
	ReviewsCount               int       `json:"reviewsCount"`
	RatingGoodReview           float64   `json:"ratingGoodReview"`
	RatingGoodReviewVoteCount  int       `json:"ratingGoodReviewVoteCount"`
	RatingKinopoisk            float64   `json:"ratingKinopoisk"`
	RatingKinopoiskVoteCount   int       `json:"ratingKinopoiskVoteCount"`
	RatingImdb                 float64   `json:"ratingImdb"`
	RatingImdbVoteCount        int       `json:"ratingImdbVoteCount"`
	RatingFilmCritics          float64   `json:"ratingFilmCritics"`
	RatingFilmCriticsVoteCount int       `json:"ratingFilmCriticsVoteCount"`
	RatingAwait                float64   `json:"ratingAwait"`
	RatingAwaitCount           int       `json:"ratingAwaitCount"`
	RatingRfCritics            float64   `json:"ratingRfCritics"`
	RatingRfCriticsVoteCount   int       `json:"ratingRfCriticsVoteCount"`
	WebURL                     string    `json:"webUrl"`
	Year                       int       `json:"year"`
	FilmLength                 int       `json:"filmLength"`
	Slogan                     string    `json:"slogan"`
	Description                string    `json:"description"`
	ShortDescription           string    `json:"shortDescription"`
	EditorAnnotation           string    `json:"editorAnnotation"`
	IsTicketsAvailable         bool      `json:"isTicketsAvailable"`
	ProductionStatus           string    `json:"productionStatus"`
	Type                       string    `json:"type"`
	RatingMpaa                 string    `json:"ratingMpaa"`
	RatingAgeLimits            string    `json:"ratingAgeLimits"`
	HasImax                    bool      `json:"hasImax"`
	Has3D                      bool      `json:"has3D"`
	LastSync                   string    `json:"lastSync"`
	Countries                  []Country `json:"countries"`
	Genres                     []Genre   `json:"genres"`
	StartYear                  int       `json:"startYear"`
	EndYear                    int       `json:"endYear"`
	Serial                     bool      `json:"serial"`
	ShortFilm                  bool      `json:"shortFilm"`
	Completed                  bool      `json:"completed"`
}

type KinopoiskStaff struct {
	StaffID        int     `json:"staffId"`
	NameRu         string  `json:"nameRu"`
	NameEn         string  `json:"nameEn"`
	Description    *string `json:"description"`
	PosterURL      string  `json:"posterUrl"`
	ProfessionText string  `json:"professionText"`
	ProfessionKey  string  `json:"professionKey"`
}

type KinopoiskMovieWithStaff struct {
	Movie *KinopoiskMovie
	Staff *[]KinopoiskStaff
}

type KinopoiskAPI struct {
	Client     *http.Client
	APIUrl     string
	APIKey     string
	APIVersion string
}

func NewKinopoiskAPI(cfg *config.KinopoiskConfig, client *http.Client) *KinopoiskAPI {
	return &KinopoiskAPI{
		Client:     client,
		APIUrl:     cfg.APIURL,
		APIKey:     cfg.APIKey,
		APIVersion: cfg.APIVersion,
	}
}

type IKinopoiskAPI interface {
	SearchMovie(id int64) (*KinopoiskMovie, error)
	SearchMovies(ids []int64) (*[]KinopoiskMovieWithStaff, error)
	SearchStaff(movieId int64) (*[]KinopoiskStaff, error)
	APIGetCall(url string) ([]byte, error)
}

func (k *KinopoiskAPI) APIGetCall(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}
	req.Header.Set("X-API-KEY", k.APIKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := k.Client.Do(req)
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}
	return body, nil
}

func (k *KinopoiskAPI) SearchMovie(movieId int64) (*KinopoiskMovie, error) {
	var url string = fmt.Sprintf(k.APIUrl, k.APIVersion) + fmt.Sprintf(MOVIES+"/%d", movieId)
	body, err := k.APIGetCall(url)
	if err != nil {
		log.Printf("Error fetching movie: %v", err)
		return nil, err
	}
	var movie KinopoiskMovie
	err = json.Unmarshal(body, &movie)
	if err != nil {
		log.Printf("Error unmarshalling response body: %v", err)
		return nil, err
	}
	return &movie, nil
}

func (k *KinopoiskAPI) SearchMovies(ids []int64) (*[]KinopoiskMovieWithStaff, error) {
	var responses []KinopoiskMovieWithStaff
	for _, id := range ids {
		movie, err := k.SearchMovie(id)
		if err != nil {
			log.Printf("Error fetching movie: %v", err)
			continue
		}
		var staff *[]KinopoiskStaff
		staff, err = k.SearchStaff(movie.KinopoiskID)
		if err != nil {
			log.Printf("Error fetching staff: %v", err)
			continue
		}
		responses = append(responses, KinopoiskMovieWithStaff{
			Movie: movie,
			Staff: staff,
		})
	}
	return &responses, nil
}

func (k *KinopoiskAPI) SearchStaff(movieId int64) (*[]KinopoiskStaff, error) {
	var url string = fmt.Sprintf(k.APIUrl+"%s?filmId=%d", "v1", STAFF, movieId)
	var staff []KinopoiskStaff
	body, err := k.APIGetCall(url)
	if err != nil {
		log.Printf("Error fetching staff: %v", err)
		return nil, err
	}
	err = json.Unmarshal(body, &staff)
	if err != nil {
		log.Printf("Error unmarshalling response body: %v", err)
		return nil, err
	}
	return &staff, nil
}
