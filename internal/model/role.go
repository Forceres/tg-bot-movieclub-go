package model

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	ID   int64
	Name string `gorm:"default:'user';check:name IN ('USER', 'ADMIN')"`
}
