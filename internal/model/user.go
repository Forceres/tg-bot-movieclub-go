package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID          int64  `gorm:"primaryKey"`
	TelegramID  int64  `gorm:"uniqueIndex;not null"`
	Username    string
	FirstName   string
	LastName    string
	Role Role `gorm:"foreignKey:RoleID"`
	RoleID int64
}