package repository

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"gorm.io/gorm"
)

type IScheduleRepository interface {
	Create(schedule *model.Schedule) (*model.Schedule, error)
	FindActive() (*model.Schedule, error)
	Update(schedule *model.Schedule) error
	Replace(schedule *model.Schedule) (*model.Schedule, error)
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

func (r *ScheduleRepository) Replace(schedule *model.Schedule) (*model.Schedule, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Schedule{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return err
		}

		if err := tx.Create(schedule).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

func (r *ScheduleRepository) FindActive() (*model.Schedule, error) {
	var schedule model.Schedule
	err := r.db.Where(&model.Schedule{IsActive: true}).Limit(1).Find(&schedule).Error
	return &schedule, err
}

func (r *ScheduleRepository) Update(schedule *model.Schedule) error {
	return r.db.Save(schedule).Error
}
