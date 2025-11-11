package model

import "gorm.io/gorm"

type Voting struct {
	gorm.Model
	ID         int64  `gorm:"primaryKey"`
	Title      string `gorm:"not null"`
	Type       string `gorm:"not null;check:type IN ('selection', 'rating')"` // selection or rating
	Status     string `gorm:"default:'active'"`
	FinishedAt *int64
	MovieID    *int
	Movie      *Movie `gorm:"foreignKey:MovieID"`
	SessionID  *int64
	Session    *Session `gorm:"foreignKey:SessionID"`
	CreatedBy  int64    `gorm:"not null"`
	Creator    User     `gorm:"foreignKey:CreatedBy"`
	Votes      []Vote   `gorm:"foreignKey:VotingID"`
}
