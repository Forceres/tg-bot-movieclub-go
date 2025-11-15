package model

import "gorm.io/gorm"

const (
	SESSION_ONGOING_STATUS   = "ONGOING"
	SESSION_FINISHED_STATUS  = "FINISHED"
	SESSION_CANCELLED_STATUS = "CANCELLED"
)

type Session struct {
	gorm.Model
	ID          int64 `gorm:"primaryKey"`
	FinishedAt  int64
	Status      string   `gorm:"default:'planned'"` // ongoing, finished, cancelled
	Description string   // Custom description for the session
	CreatedBy   int64    `gorm:"not null"`
	Creator     User     `gorm:"foreignKey:CreatedBy"`
	Movies      []Movie  `gorm:"many2many:movies_sessions;"`
	Votings     []Voting `gorm:"foreignKey:SessionID"`
}
