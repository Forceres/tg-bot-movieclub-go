package model

import "gorm.io/gorm"

type Movie struct {
	gorm.Model
	ID          int
	Title       string
	Description string
	Directors    string
	Year        int
	Countries		string
	Genres			string
	Link        string
	Duration    int
	IMDBRating  float64
	Rating      float64
	StartedAt  string
	FinishedAt string
	SuggestedAt string
	SuggestedBy	string
	Sessions	 []Session `gorm:"many2many:movies_sessions;"` 
}