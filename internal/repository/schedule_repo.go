package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type IScheduleRepository interface {
	Create(schedule *model.Schedule) (*model.Schedule, error)
	FindActive() (*model.Schedule, error)
	Update(schedule *model.Schedule) error
}

type ScheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) IScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) Create(schedule *model.Schedule) (*model.Schedule, error) {
	err := r.db.Create(schedule).Error
	return schedule, err
}

func (r *ScheduleRepository) FindActive() (*model.Schedule, error) {
	var schedule *model.Schedule
	err := r.db.Where("is_active = ?", true).Find(&schedule).Limit(1).Error
	return schedule, err
}

func (r *ScheduleRepository) Update(schedule *model.Schedule) error {
	return r.db.Save(schedule).Error
}
