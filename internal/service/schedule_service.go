package service

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/date"
)

type IScheduleService interface {
	CreateSchedule(schedule *model.Schedule) (*model.Schedule, error)
	GetActiveSchedule() (*model.Schedule, error)
	UpdateSchedule(schedule *model.Schedule) error
	GetNextScheduledTime() (int64, error)
}

type ScheduleService struct {
	scheduleRepo repository.IScheduleRepository
}

func NewScheduleService(scheduleRepo repository.IScheduleRepository) IScheduleService {
	return &ScheduleService{scheduleRepo: scheduleRepo}
}

func (s *ScheduleService) CreateSchedule(schedule *model.Schedule) (*model.Schedule, error) {
	return s.scheduleRepo.Create(schedule)
}

func (s *ScheduleService) GetActiveSchedule() (*model.Schedule, error) {
	return s.scheduleRepo.FindActive()
}

func (s *ScheduleService) UpdateSchedule(schedule *model.Schedule) error {
	return s.scheduleRepo.Update(schedule)
}

func (s *ScheduleService) GetNextScheduledTime() (int64, error) {
	schedule, err := s.scheduleRepo.FindActive()
	if err != nil || schedule == nil {
		return 0, err
	}

	return date.GetRelativeDate(&date.GetRelativeDateParams{
		Day:      &schedule.Weekday,
		Minute:   &schedule.Minute,
		Hour:     &schedule.Hour,
		Location: &schedule.Location,
	}), nil
}
