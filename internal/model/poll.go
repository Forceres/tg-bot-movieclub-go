package model

import "gorm.io/gorm"

const (
	POLL_OPENED_STATUS = "opened"
	POLL_CLOSED_STATUS = "closed"
)

type Poll struct {
	gorm.Model
	ID        int64  `gorm:"primaryKey"`
	MessageID int    `gorm:"not null"`
	PollID    string `gorm:"uniqueIndex;not null"` // Telegram poll ID
	VotingID  int64  `gorm:"not null"`
	Voting    Voting `gorm:"foreignKey:VotingID"`
	MovieID   *int64 // For rating polls
	Movie     *Movie `gorm:"foreignKey:MovieID"`
	Type      string `gorm:"not null"`         // "SELECTION" or "RATING"
	Status    string `gorm:"default:'active'"` // active, closed
}
