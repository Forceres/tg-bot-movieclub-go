package service

import (
	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/date"
)

type IScheduleService interface {
	CreateSchedule(schedule *model.Schedule) (*model.Schedule, error)
	GetActiveSchedule() (*model.Schedule, error)
	ReplaceSchedule(schedule *model.Schedule) (*model.Schedule, error)
	UpdateSchedule(schedule *model.Schedule) error
	GetNextScheduledTime() (int64, error)
}

type ScheduleService struct {
	repo repository.IScheduleRepository
}

func NewScheduleService(scheduleRepo repository.IScheduleRepository) IScheduleService {
	return &ScheduleService{repo: scheduleRepo}
}

func (s *ScheduleService) CreateSchedule(schedule *model.Schedule) (*model.Schedule, error) {
	return s.repo.Create(schedule)
}

func (s *ScheduleService) ReplaceSchedule(schedule *model.Schedule) (*model.Schedule, error) {
	return s.repo.Replace(schedule)
}

func (s *ScheduleService) GetActiveSchedule() (*model.Schedule, error) {
	return s.repo.FindActive()
}

func (s *ScheduleService) UpdateSchedule(schedule *model.Schedule) error {
	return s.repo.Update(schedule)
}

func (s *ScheduleService) GetNextScheduledTime() (int64, error) {
	schedule, err := s.repo.FindActive()
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
