package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Movie struct {
	CreatedAt   time.Time  `json:"CreatedAt"`
	UpdatedAt   time.Time  `json:"UpdatedAt"`
	DeletedAt   *time.Time `json:"DeletedAt"`
	ID          int64      `json:"ID"`
	Title       string     `json:"Title"`
	Description string     `json:"Description"`
	Directors   string     `json:"Directors"`
	Year        int        `json:"Year"`
	Countries   string     `json:"Countries"`
	Genres      string     `json:"Genres"`
	Link        string     `json:"Link"`
	Duration    int        `json:"Duration"`
	IMDBRating  float64    `json:"IMDBRating"`
	Rating      float64    `json:"Rating"`
	Status      string     `json:"Status"`
	WatchCount  int        `json:"WatchCount"`
	FinishedAt  string     `json:"FinishedAt"`
	SuggestedAt *int64     `json:"SuggestedAt"`
	SuggestedBy *int64     `json:"SuggestedBy"`
}

func escapeSQLString(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	return s
}

func main() {
	data, err := os.ReadFile("movies.json")
	if err != nil {
		log.Fatalf("Failed to read movies.json: %v", err)
	}

	var movies []Movie
	if err := json.Unmarshal(data, &movies); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	sqlFile, err := os.Create("movies_insert.sql")
	if err != nil {
		log.Fatalf("Failed to create SQL file: %v", err)
	}
	defer sqlFile.Close()

	sqlFile.WriteString("-- Movies INSERT statements\n")
	sqlFile.WriteString("-- Generated at: " + time.Now().Format(time.RFC3339) + "\n\n")

	for _, movie := range movies {
		var createdAt, updatedAt, deletedAt, finishedAt, suggestedAt, suggestedBy string

		if !movie.CreatedAt.IsZero() && movie.CreatedAt.Year() > 1 {
			createdAt = fmt.Sprintf("'%s'", movie.CreatedAt.Format("2006-01-02 15:04:05"))
		} else {
			createdAt = "CURRENT_TIMESTAMP"
		}

		if !movie.UpdatedAt.IsZero() && movie.UpdatedAt.Year() > 1 {
			updatedAt = fmt.Sprintf("'%s'", movie.UpdatedAt.Format("2006-01-02 15:04:05"))
		} else {
			updatedAt = "CURRENT_TIMESTAMP"
		}

		if movie.DeletedAt != nil {
			deletedAt = fmt.Sprintf("'%s'", movie.DeletedAt.Format("2006-01-02 15:04:05"))
		} else {
			deletedAt = "NULL"
		}

		if movie.FinishedAt != "" {
			parsedTime, err := time.Parse(time.RFC3339, movie.FinishedAt)
			if err == nil {
				finishedAt = fmt.Sprintf("'%s'", parsedTime.Format("2006-01-02 15:04:05"))
			} else {
				finishedAt = "NULL"
			}
		} else {
			finishedAt = "NULL"
		}

		if movie.SuggestedAt != nil {
			suggestedAt = fmt.Sprintf("%d", *movie.SuggestedAt)
		} else {
			suggestedAt = "NULL"
		}

		if movie.SuggestedBy != nil {
			suggestedBy = fmt.Sprintf("%d", *movie.SuggestedBy)
		} else {
			suggestedBy = "NULL"
		}

		title := escapeSQLString(movie.Title)
		description := escapeSQLString(movie.Description)
		directors := escapeSQLString(movie.Directors)
		countries := escapeSQLString(movie.Countries)
		genres := escapeSQLString(movie.Genres)
		link := escapeSQLString(movie.Link)
		status := escapeSQLString(movie.Status)

		sql := fmt.Sprintf(
			`INSERT INTO movies (created_at, updated_at, deleted_at, id, title, description, directors, year, countries, genres, link, duration, imdb_rating, rating, status, watch_count, finished_at, suggested_at, suggested_by)
VALUES (%s, %s, %s, %d, '%s', '%s', '%s', %d, '%s', '%s', '%s', %d, %.1f, %.1f, '%s', %d, %s, %s, %s);

`,
			createdAt, updatedAt, deletedAt, movie.ID, title, description, directors, movie.Year,
			countries, genres, link, movie.Duration, movie.IMDBRating, movie.Rating,
			status, movie.WatchCount, finishedAt, suggestedAt, suggestedBy,
		)

		sqlFile.WriteString(sql)
	}

	fmt.Printf("✅ Successfully generated SQL file with %d INSERT statements: movies_insert.sql\n", len(movies))
}
