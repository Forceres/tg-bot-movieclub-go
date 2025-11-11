package model

import "gorm.io/gorm"

type Vote struct {
	gorm.Model
	ID       int64 `gorm:"primaryKey"`
	VotingID int64
	Voting   Voting `gorm:"foreignKey:VotingID"`
	UserID   int64
	User     User `gorm:"foreignKey:UserID"`
	MovieID  *int
	Movie    *Movie `gorm:"foreignKey:MovieID"`
	Rating   *int   `gorm:"check:rating >= 1 AND rating <= 10"`
}
