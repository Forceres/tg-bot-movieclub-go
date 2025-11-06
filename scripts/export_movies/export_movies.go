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
    if err := database.Find(&movies).Error; err != nil {
        log.Fatalf("Failed to fetch movies: %v", err)
    }

    file, err := os.Create("movies.json")
    if err != nil {
        log.Fatalf("Failed to create file: %v", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(movies); err != nil {
        log.Fatalf("Failed to encode JSON: %v", err)
    }

    log.Printf("Exported %d movies to movies.json", len(movies))
}