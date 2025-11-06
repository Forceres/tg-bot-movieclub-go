package model

import "gorm.io/gorm"

type Session struct {
	gorm.Model
	ID          int64  `gorm:"primaryKey"`
	SessionDate int64  `gorm:"not null"` // Unix timestamp
	Location    string
	Status      string `gorm:"default:'planned'"` // planned, ongoing, finished, cancelled
	CreatedBy   int64  `gorm:"not null"`
	Creator     User   `gorm:"foreignKey:CreatedBy"`
	Movies      []Movie   `gorm:"many2many:movies_sessions;"`
	Votings     []Voting  `gorm:"foreignKey:SessionID"`
}