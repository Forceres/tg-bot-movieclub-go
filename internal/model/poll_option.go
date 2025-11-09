package model

import "gorm.io/gorm"

type PollOption struct {
	gorm.Model
	ID          int64 `gorm:"primaryKey"`
	PollID      int64 `gorm:"not null"`
	Poll        Poll  `gorm:"foreignKey:PollID"`
	OptionIndex int   `gorm:"not null"` // 0, 1, 2, 3... (position in poll)
	MovieID     int   `gorm:"not null"`
	Movie       Movie `gorm:"foreignKey:MovieID"`
}
