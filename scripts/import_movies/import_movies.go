package main

import (
	"encoding/json"
	"flag"
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

	// Parse command line flags
	jsonFile := flag.String("file", "movies.json", "Path to JSON file with movies")
	flag.Parse()

	database, err := db.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Read JSON file
	file, err := os.Open(*jsonFile)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	var movies []model.Movie
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&movies); err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)
	}

	// Import movies
	imported := 0
	skipped := 0
	for _, movie := range movies {
		var existing model.Movie
		result := database.Where("title = ? AND year = ?", movie.Title, movie.Year).First(&existing)

		if result.Error != nil {
			// Movie doesn't exist, create it
			if err := database.Create(&movie).Error; err != nil {
				log.Printf("Failed to import movie '%s': %v", movie.Title, err)
			} else {
				imported++
				log.Printf("Imported: %s (%d)", movie.Title, movie.Year)
			}
		} else {
			// Movie already exists, skip
			skipped++
			log.Printf("Skipped (already exists): %s (%d)", movie.Title, movie.Year)
		}
	}

	log.Printf("\nImport completed: %d imported, %d skipped, %d total", imported, skipped, len(movies))
}
