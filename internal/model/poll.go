package model

import "gorm.io/gorm"

type Poll struct {
	gorm.Model
	ID        int64  `gorm:"primaryKey"`
	MessageID int    `gorm:"not null"`
	PollID    string `gorm:"uniqueIndex;not null"` // Telegram poll ID
	VotingID  int64  `gorm:"not null"`
	Voting    Voting `gorm:"foreignKey:VotingID"`
	MovieID   *int   // For rating polls
	Movie     *Movie `gorm:"foreignKey:MovieID"`
	Type      string `gorm:"not null"`         // "selection" or "rating"
	Status    string `gorm:"default:'active'"` // active, closed
}
