package model

import "gorm.io/gorm"

const (
	MOVIE_SUGGESTED_STATUS = "SUGGESTED"
	MOVIE_WATCHED_STATUS   = "WATCHED"
)

type Movie struct {
	gorm.Model
	ID          int64
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
	Status      string  `gorm:"default:'SUGGESTED'"`
	WatchCount  int     `gorm:"default:0"`
	FinishedAt  *string `gorm:"default:null"`
	SuggestedAt *int64
	SuggestedBy *int64    `gorm:"default:null"`
	Suggester   *User     `gorm:"foreignKey:SuggestedBy"`
	Sessions    []Session `gorm:"many2many:movies_sessions;"`
}
