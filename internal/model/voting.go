package model

import "gorm.io/gorm"

const (
	VOTING_ACTIVE_STATUS    = "ACTIVE"
	VOTING_INACTIVE_STATUS  = "INACTIVE"
	VOTING_CANCELLED_STATUS = "CANCELLED"
	VOTING_RATING_TYPE      = "RATING"
	VOTING_SELECTION_TYPE   = "SELECTION"
)

type Voting struct {
	gorm.Model
	ID         int64  `gorm:"primaryKey"`
	Title      string `gorm:"not null"`
	Type       string `gorm:"not null;check:type IN ('SELECTION', 'RATING')"` // selection or rating
	Status     string `gorm:"default:'ACTIVE'"`
	FinishedAt *int64
	MovieID    *int64
	Movie      *Movie `gorm:"foreignKey:MovieID"`
	SessionID  *int64
	Session    *Session `gorm:"foreignKey:SessionID"`
	CreatedBy  int64    `gorm:"not null"`
	Creator    User     `gorm:"foreignKey:CreatedBy"`
	Votes      []Vote   `gorm:"foreignKey:VotingID"`
}
