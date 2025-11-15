package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/Forceres/tg-bot-movieclub-go/internal/db"
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/ilyakaznacheev/cleanenv"
)

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

	var movies []model.Movie
	if err := database.Table("movies").Find(&movies).Error; err != nil {
		log.Fatalf("Failed to fetch movies: %v", err)
	}

	var moviesExport []model.Movie
	for _, movie := range movies {
		var suggestedAt *int64
		if movie.SuggestedAt != nil && *movie.SuggestedAt < 0 {
			suggestedAt = nil
		}
		moviesExport = append(moviesExport, model.Movie{
			ID:          movie.ID,
			Title:       movie.Title,
			Year:        movie.Year,
			Description: movie.Description,
			Directors:   movie.Directors,
			Countries:   movie.Countries,
			Genres:      movie.Genres,
			Link:        movie.Link,
			IMDBRating:  movie.IMDBRating,
			Rating:      movie.Rating,
			SuggestedAt: suggestedAt,
			SuggestedBy: movie.SuggestedBy,
			Status:      movie.Status,
			Duration:    movie.Duration,
			WatchCount:  movie.WatchCount,
			FinishedAt:  movie.FinishedAt, // Map from FinishedAt
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
