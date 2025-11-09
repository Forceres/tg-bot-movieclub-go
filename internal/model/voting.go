package model

import "gorm.io/gorm"

type Voting struct {
	gorm.Model
	ID         int64  `gorm:"primaryKey"`
	Title      string `gorm:"not null"`                                       // e.g., "Vote for next movie", "Rate the movie"
	Type       string `gorm:"not null;check:type IN ('selection', 'rating')"` // "selection" (choose movie) or "rating" (rate after session)
	Status     string `gorm:"default:'active'"`                               // active, closed, cancelled
	FinishedAt *int64 // Unix timestamp, nullable
	// For "selection" type: MovieID is NULL, users vote for different movies in Votes
	// For "rating" type: MovieID references the movie being rated
	MovieID   *int     // Optional: associated movie (for rating type)
	Movie     *Movie   `gorm:"foreignKey:MovieID"`
	SessionID *int64   // Optional: associated session
	Session   *Session `gorm:"foreignKey:SessionID"`
	CreatedBy int64    `gorm:"not null"`
	Creator   User     `gorm:"foreignKey:CreatedBy"`
	Votes     []Vote   `gorm:"foreignKey:VotingID"`
}
