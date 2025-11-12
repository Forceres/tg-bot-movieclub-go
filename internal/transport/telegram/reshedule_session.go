package telegram

import (
	"context"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/datepicker"
	fsmutils "github.com/Forceres/tg-bot-movieclub-go/internal/utils/fsm"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
)

type ResheduleSessionHandler struct {
	f          *fsm.FSM
	datepicker *datepicker.Datepicker
}

type IRescheduleSessionHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
	PrepareDate(f *fsm.FSM, args ...any)
	PrepareTime(f *fsm.FSM, args ...any)
	PrepareLocation(f *fsm.FSM, args ...any)
	RescheduleSession(f *fsm.FSM, args ...any)
}

func NewResheduleSessionHandler(f *fsm.FSM, datepicker *datepicker.Datepicker) IRescheduleSessionHandler {
	return &ResheduleSessionHandler{f: f, datepicker: datepicker}
}

func (h *ResheduleSessionHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	currentState := h.f.Current(userID)
	if currentState != stateDefault {
		return
	}
	h.f.Transition(userID, stateDate, userID, ctx, b, update)
}

func (h *ResheduleSessionHandler) PrepareDate(f *fsm.FSM, args ...any) {
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
		Text:   "Обновление даты сессии просмотра...",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
		f.Reset(userID)
		return
	}
	fsmutils.AppendMessageID(f, userID, msg.ID)
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выбери дату",
		ReplyMarkup: h.datepicker.Datepicker,
	})
	if err != nil {
		log.Printf("Error sending datepicker: %v", err)
		f.Reset(userID)
	}
}

func (h *ResheduleSessionHandler) PrepareTime(f *fsm.FSM, args ...any) {
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
		Text:   "Введите время в формате ЧЧ:ММ (например, 18:30)",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
		f.Reset(userID)
		return
	}
	fsmutils.AppendMessageID(f, userID, msg.ID)
}

func (h *ResheduleSessionHandler) PrepareLocation(f *fsm.FSM, args ...any) {
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
		Text:   "Введите локацию (например, Europe/Moscow)",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
		f.Reset(userID)
		return
	}
	fsmutils.AppendMessageID(f, userID, msg.ID)
}

func (h *ResheduleSessionHandler) RescheduleSession(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	// hour, _ := f.Get(userID, "hour")
	// minute, _ := f.Get(userID, "minute")
	// location, _ := f.Get(userID, "location")
	// date, _ := f.Get(userID, "date")
	// schedule := &model.Schedule{
	// 	Weekday:  int(date.(time.Time).Weekday()) + 1,
	// 	Hour:     hour.(int),
	// 	Minute:   minute.(int),
	// 	IsActive: true,
	// 	Location: location.(string),
	// }

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Сессия просмотра успешно обновлена.",
	})
	fsmutils.DeleteMessages(ctx, b, f, userID, update.Message.Chat.ID)
	f.Reset(userID)
}
