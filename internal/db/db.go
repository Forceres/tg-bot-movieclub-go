package db

import (
	"log"
	"os"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewSqliteDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
    newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
        logger.Config{
            SlowThreshold:              time.Second,   // Slow SQL threshold
            LogLevel:                   logger.Info, // Log level
            IgnoreRecordNotFoundError: true,           // Ignore ErrRecordNotFound error for logger
            ParameterizedQueries:      true,           // Don't include params in the SQL log
            Colorful:                  true,          // Disable color
        },
    )

    db, err := gorm.Open(sqlite.Open(cfg.Name), &gorm.Config{
        Logger: newLogger,
        PrepareStmt: true,
    })
    if err != nil {
        panic("Failed to connect database")
    }

    // Migrate the schema
    db.AutoMigrate(&model.Role{})
    db.AutoMigrate(&model.User{})
    db.AutoMigrate(&model.Movie{})
    db.AutoMigrate(&model.Session{})
    db.AutoMigrate(&model.Voting{})
    db.AutoMigrate(&model.Vote{})

    // Seed roles
    seedRoles(db)

    return db, nil
}

func seedRoles(db *gorm.DB) {
    roles := []model.Role{
        {Name: "ADMIN"},
        {Name: "USER"},
    }

    for _, role := range roles {
        var existingRole model.Role
        result := db.Where("name = ?", role.Name).First(&existingRole)
        if result.Error == gorm.ErrRecordNotFound {
            if err := db.Create(&role).Error; err != nil {
                log.Printf("Failed to create role %s: %v", role.Name, err)
            } else {
                log.Printf("Created role: %s", role.Name)
            }
        }
    }
}