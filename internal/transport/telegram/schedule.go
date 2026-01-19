package telegram

import (
	"context"
	"log"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/datepicker"
	fsmutils "github.com/Forceres/tg-bot-movieclub-go/internal/utils/fsm"
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
	f                 *fsm.FSM
	scheduleService   service.IScheduleService
	datepicker        *datepicker.Datepicker
	sessionDatepicker *datepicker.Datepicker
}

type IScheduleHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
	HandleReschedule(ctx context.Context, b *bot.Bot, update *models.Update)
	PrepareDate(f *fsm.FSM, args ...any)
	PrepareTime(f *fsm.FSM, args ...any)
	PrepareLocation(f *fsm.FSM, args ...any)
	SaveSchedule(f *fsm.FSM, args ...any)
}

func NewScheduleHandler(scheduleService service.IScheduleService, f *fsm.FSM, datepicker *datepicker.Datepicker, sessionDatepicker *datepicker.Datepicker) IScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService, f: f, datepicker: datepicker, sessionDatepicker: sessionDatepicker}
}

func (h *ScheduleHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	schedule, err := h.scheduleService.GetActiveSchedule()
	if err != nil || schedule == nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "ℹ️ Нет активного расписания.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
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
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "📅 Активное расписание: " + out,
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *ScheduleHandler) HandleReschedule(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	currentState := h.f.Current(userID)
	if currentState != stateDefault {
		return
	}
	h.f.Transition(userID, stateDate, userID, ctx, b, update)
}

func (h *ScheduleHandler) PrepareDate(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	// reschedule_session
	var title string
	dp, _ := f.Get(userID, "datepicker")
	if dp == "session" {
		title = "Обновление даты сессии..."
	} else {
		title = "Изменение расписания..."
	}
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   title,
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
		f.Reset(userID)
		return
	}
	fsmutils.AppendMessageID(f, userID, msg.ID)
	var datepicker *datepicker.Datepicker
	if dp == "session" {
		datepicker = h.sessionDatepicker
	} else {
		datepicker = h.datepicker
	}
	msg, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выбери дату",
		ReplyMarkup: datepicker.Datepicker,
	})
	if err != nil {
		log.Printf("Error sending datepicker: %v", err)
		f.Reset(userID)
		return
	}
	fsmutils.AppendMessageID(f, userID, msg.ID)
}

func (h *ScheduleHandler) PrepareTime(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	callbackQuery := args[3].(*models.CallbackQuery)
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: callbackQuery.Message.Message.Chat.ID,
		Text:   "🕐 Введите время в формате ЧЧ:ММ (например, 18:30)",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
		f.Reset(userID)
		return
	}
	fsmutils.AppendMessageID(f, userID, msg.ID)
}

func (h *ScheduleHandler) PrepareLocation(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "🌍 Введите локацию (например, Europe/Moscow)",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
		f.Reset(userID)
		return
	}
	fsmutils.AppendMessageID(f, userID, msg.ID)
}

func (h *ScheduleHandler) SaveSchedule(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	hour, _ := f.Get(userID, "hour")
	minute, _ := f.Get(userID, "minute")
	location, _ := f.Get(userID, "location")
	date, _ := f.Get(userID, "date")
	schedule := &model.Schedule{
		Weekday:  int(date.(time.Time).Weekday()),
		Hour:     hour.(int),
		Minute:   minute.(int),
		IsActive: true,
		Location: location.(string),
	}
	_, err := h.scheduleService.ReplaceSchedule(schedule)
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Ошибка при сохранении расписания.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		f.Reset(userID)
		return
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "✅ Расписание успешно обновлено.",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
	fsmutils.DeleteMessages(ctx, b, f, userID, update.Message.Chat.ID)
	f.Reset(userID)
}
