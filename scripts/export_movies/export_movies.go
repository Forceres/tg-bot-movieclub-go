package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/Forceres/tg-bot-movieclub-go/internal/db"
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/ilyakaznacheev/cleanenv"
	"gorm.io/gorm"
)

type MovieExport struct {
    gorm.Model
    ID         int  `json:"id"`
    Title       string `json:"title"`
	Description string `json:"description"`
	Director    string `json:"director"`
	Year        int    `json:"year"`
	Countries		string `json:"countries"`
	Genres			string `json:"genres"`
	Link        string `json:"link"`
	Duration    int    `json:"duration"`
	IMDBRating  float64 `json:"imdb_rating"`
	Rating      float64 `json:"rating"`
    StartWatch string `json:"start_watch"`
    FinishWatch string `json:"finish_watch"`
    CreatedAt  string `json:"created_at"`
    UpdatedAt  string `json:"updated_at"`
    SuggestedAt string `json:"suggested_at"`
	SuggestedBy	string `json:"suggested_by"`
}

func main() {
    var cfg config.Config
    err := cleanenv.ReadConfig(".env", &cfg)
    if err != nil {	
        panic(err)
    }
    database, err := db.NewSqliteDB(cfg.Database)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    var movies []MovieExport
    if err := database.Table("movies").Find(&movies).Error; err != nil {
        log.Fatalf("Failed to fetch movies: %v", err)
    }

    // Map to export format
    var moviesExport []model.Movie
    for _, movie := range movies {
        moviesExport = append(moviesExport, model.Movie{
            ID:          movie.ID,
            Title:       movie.Title,
            Year:        movie.Year,
            Description: movie.Description,
            Directors:    movie.Director,
            Countries:    movie.Countries,
            Genres:       movie.Genres,
            Link: movie.Link,
            IMDBRating: movie.IMDBRating,
            Rating: movie.Rating,
            SuggestedAt: movie.SuggestedAt,
            SuggestedBy: movie.SuggestedBy,
            Duration:    movie.Duration,
            StartedAt:   movie.StartWatch,   // Map from StartWatch
            FinishedAt:  movie.FinishWatch,  // Map from FinishWatch
        })
    }

    file, err := os.Create("movies.json")
    if err != nil {
        log.Fatalf("Failed to create file: %v", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(moviesExport); err != nil {
        log.Fatalf("Failed to encode JSON: %v", err)
    }

    log.Printf("Exported %d movies to movies.json", len(moviesExport))
}