package model

import "gorm.io/gorm"

const (
	ROLE_USER  = "USER"
	ROLE_ADMIN = "ADMIN"
)

type Role struct {
	gorm.Model
	ID   int64
	Name string `gorm:"default:'user';check:name IN ('USER', 'ADMIN')"`
}
