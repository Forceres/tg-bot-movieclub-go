package model

import "gorm.io/gorm"

type Movie struct {
	gorm.Model
	ID          int
	Title       string
	Description string
	Directors   string
	Year        int
	Countries   string
	Genres      string
	Link        string
	Duration    int
	IMDBRating  float64
	Rating      float64
	Status      string  `gorm:"default:'suggested'"`
	WatchCount  int     `gorm:"default:0"`
	FinishedAt  *string `gorm:"default:null"`
	SuggestedAt *int64
	SuggestedBy *int64    `gorm:"default:null"`
	Sessions    []Session `gorm:"many2many:movies_sessions;"`
}
