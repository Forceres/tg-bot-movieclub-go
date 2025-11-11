package model

import (
	"gorm.io/gorm"
)

type Schedule struct {
	gorm.Model
	ID       int64  `gorm:"primaryKey"`
	Weekday  int    `gorm:"not null"` // 1=Monday, ..., 6=Saturday, 7=Sunday
	Hour     int    `gorm:"not null"` // 0-23
	Minute   int    `gorm:"not null"` // 0-59
	IsActive bool   `gorm:"default:true"`
	Location string `gorm:"default:'Europe/Moscow'"`
}
