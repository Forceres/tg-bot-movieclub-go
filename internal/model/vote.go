package model

import "gorm.io/gorm"

type Vote struct {
	gorm.Model
	ID       int64  `gorm:"primaryKey"`
	VotingID int64  `gorm:"index:,unique,composite:idx_user_voting"`
	Voting   Voting `gorm:"foreignKey:VotingID"`
	UserID   int64  `gorm:"index:,unique,composite:idx_user_voting"`
	User     User   `gorm:"foreignKey:UserID"`
	MovieID  *int64
	Movie    *Movie `gorm:"foreignKey:MovieID"`
	Rating   *int   `gorm:"check:rating >= 1 AND rating <= 10"`
}
