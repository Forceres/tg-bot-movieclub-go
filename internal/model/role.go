package model

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	ID   int64
	Name      string
}
