package telegram

import (
	"context"
	"log"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"github.com/goodsign/monday"
)

const (
	stateDate         fsm.StateID = "date"
	stateTime         fsm.StateID = "time"
	stateLocation     fsm.StateID = "location"
	stateSaveSchedule fsm.StateID = "save_schedule"
)

type ScheduleHandler struct {
	f               *fsm.FSM
	scheduleService service.IScheduleService
}

type IScheduleHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
	HandleReschedule(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewScheduleHandler(scheduleService service.IScheduleService, f *fsm.FSM) IScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService, f: f}
}

func (h *ScheduleHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	schedule, err := h.scheduleService.GetActiveSchedule()
	if err != nil || schedule == nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Нет активного расписания.",
		})
		return
	}
	var weekday int = schedule.Weekday
	if weekday == 7 {
		weekday = 0
	}
	locale, err := time.LoadLocation(schedule.Location)
	if err != nil {
		log.Printf("Error loading location: %v", err)
		locale = time.FixedZone("UTC+3", 3*60*60)
	}
	now := time.Now().In(locale)
	base := time.Date(now.Year(), now.Month(), now.Day(), schedule.Hour, schedule.Minute, 0, 0, locale)
	daysUntil := (int(weekday) - int(base.Weekday()) + 7) % 7
	target := base.AddDate(0, 0, daysUntil)
	out := monday.Format(target, "Monday, 02-January-06 15:04:05 MST", monday.LocaleRuRU)
	if err != nil {
		log.Printf("Error parsing time: %v", err)
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Активное расписание: " + out,
	})
}

func (h *ScheduleHandler) HandleReschedule(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	h.f.Transition(userID, stateDate, userID, ctx, b, update)
}

func (h *ScheduleHandler) PrepareDate(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	// ctx := args[1].(context.Context)
	// b := args[2].(*bot.Bot)
	// update := args[3].(*models.Update)
}

func (h *ScheduleHandler) PrepareTime(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	// ctx := args[1].(context.Context)
	// b := args[2].(*bot.Bot)
	// update := args[3].(*models.Update)
}

func (h *ScheduleHandler) PrepareLocation(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	// ctx := args[1].(context.Context)
	// b := args[2].(*bot.Bot)
	// update := args[3].(*models.Update)
}

func (h *ScheduleHandler) SaveSchedule(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	// ctx := args[1].(context.Context)
	// b := args[2].(*bot.Bot)
	// update := args[3].(*models.Update)
}
