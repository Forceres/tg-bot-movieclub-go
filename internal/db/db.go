package db

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqliteDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
		db, err := gorm.Open(sqlite.Open(cfg.Name), &gorm.Config{})
		if err != nil {
			panic("Failed to connect database")
		}

		// Migrate the schema
		// db.AutoMigrate(&Product{})

    return db, nil
}
