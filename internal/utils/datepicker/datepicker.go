package datepicker

import (
	"context"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/telegram/datepicker"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
)

const stateDefault fsm.StateID = "default"
const stateTime fsm.StateID = "time"

type Datepicker struct {
	f          *fsm.FSM
	Datepicker *datepicker.DatePicker
}

func NewDatepicker(f *fsm.FSM) *Datepicker {
	return &Datepicker{f: f}
}

func (d *Datepicker) OnDatepickerCancel(ctx context.Context, b *bot.Bot, callbackQuery *models.CallbackQuery) {
	userID := callbackQuery.From.ID
	currentState := d.f.Current(userID)
	if currentState == stateDefault {
		return
	}
	d.f.Reset(userID)
}

func (d *Datepicker) OnDatepickerSelect(ctx context.Context, b *bot.Bot, callbackQuery *models.CallbackQuery, date time.Time) {
	userID := callbackQuery.From.ID
	currentState := d.f.Current(userID)
	if currentState == stateDefault {
		return
	}
	d.f.Set(userID, "date", date)
	d.f.Transition(userID, stateTime, userID, ctx, b, callbackQuery)
}

func ScheduleDatepicker(b *bot.Bot, d *Datepicker) {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	date := time.Date(year, month, 0, 0, 0, 0, 0, time.UTC)
	curr := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	curr = curr.AddDate(0, 0, -1)
	daysToAdd := curr.Day() - 1
	opts := []datepicker.Option{
		datepicker.From(date),
		datepicker.OnCancel(d.OnDatepickerCancel),
		datepicker.To(date.AddDate(0, 0, daysToAdd)),
		datepicker.Language("ru"),
		datepicker.WithPrefix("datepicker"),
	}
	d.Datepicker = datepicker.New(
		b,
		d.OnDatepickerSelect,
		opts...,
	)
}

func SessionDatepicker(b *bot.Bot, d *Datepicker) {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	date := time.Date(year, month, 0, 0, 0, 0, 0, time.UTC)
	opts := []datepicker.Option{
		datepicker.From(date),
		datepicker.OnCancel(d.OnDatepickerCancel),
		datepicker.Language("ru"),
		datepicker.WithPrefix("datepicker"),
	}
	d.Datepicker = datepicker.New(
		b,
		d.OnDatepickerSelect,
		opts...,
	)
}
