package model

import "gorm.io/gorm"

type Vote struct {
	gorm.Model
	ID       int64  `gorm:"primaryKey"`
	VotingID int64  // Parent voting session
	Voting   Voting `gorm:"foreignKey:VotingID"`
	UserID   int64
	User     User `gorm:"foreignKey:UserID"`
	// For "selection" voting: MovieID contains the movie user voted for
	// For "rating" voting: MovieID should match Voting.MovieID (or be NULL)
	MovieID *int   // Required for selection votes, optional for rating
	Movie   *Movie `gorm:"foreignKey:MovieID"`

	// For "rating" voting: Rating contains the score (1-10)
	// For "selection" voting: Rating is NULL
	Rating *int `gorm:"check:rating >= 1 AND rating <= 10"`
}
