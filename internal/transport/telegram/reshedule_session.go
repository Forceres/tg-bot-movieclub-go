package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/tasks"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/datepicker"
	fsmutils "github.com/Forceres/tg-bot-movieclub-go/internal/utils/fsm"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"github.com/hibiken/asynq"
)

type ResheduleSessionHandler struct {
	f              *fsm.FSM
	datepicker     *datepicker.Datepicker
	sessionService service.ISessionService
	inspector      *asynq.Inspector
	client         *asynq.Client
}

type IRescheduleSessionHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
	PrepareDate(f *fsm.FSM, args ...any)
	PrepareTime(f *fsm.FSM, args ...any)
	PrepareLocation(f *fsm.FSM, args ...any)
	RescheduleSession(f *fsm.FSM, args ...any)
}

func NewResheduleSessionHandler(f *fsm.FSM, datepicker *datepicker.Datepicker, sessionService service.ISessionService, inspector *asynq.Inspector, client *asynq.Client) IRescheduleSessionHandler {
	return &ResheduleSessionHandler{f: f, datepicker: datepicker, sessionService: sessionService, inspector: inspector, client: client}
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
	session, err := h.sessionService.FindOngoingSession()
	if err != nil {
		log.Printf("Error finding ongoing session: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Нет текущей сессии просмотра.",
		})
		f.Reset(userID)
		return
	}
	hour, _ := f.Get(userID, "hour")
	minute, _ := f.Get(userID, "minute")
	location, _ := f.Get(userID, "location")
	date, _ := f.Get(userID, "date")
	finishedAt := date.(time.Time).In(time.FixedZone(location.(string), 0)).Add(time.Duration(hour.(int)) * time.Hour).Add(time.Duration(minute.(int)) * time.Minute).Unix()
	err = h.sessionService.RescheduleSession(session.ID, finishedAt)
	if err != nil {
		log.Printf("Error rescheduling session: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка при обновлении сессии просмотра.",
		})
		f.Reset(userID)
		return
	}
	activeTasks, err := h.inspector.ListActiveTasks(tasks.FinishSessionTaskType)
	if err != nil {
		log.Printf("Error listing active tasks: %v", err)
	}
	for _, t := range activeTasks {
		var p tasks.FinishSessionTaskPayload
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			log.Printf("Error unmarshaling task payload: %v", err)
			continue
		}
		if p.SessionID == session.ID {
			if err := h.inspector.DeleteTask(tasks.FinishSessionTaskType, t.ID); err != nil {
				log.Printf("Error deleting task: %v", err)
			} else {
				log.Printf("Deleted scheduled finish session task: %s", t.ID)
			}
			break
		}
	}
	err = tasks.EnqueueFinishSessionTask(h.client, &tasks.EnqueueFinishSessionParams{
		SessionID: session.ID,
		Duration:  time.Duration(finishedAt),
	})
	if err != nil {
		log.Printf("Error scheduling new finish session task: %v", err)
	} else {
		log.Printf("Scheduled new finish session task for session: %d", session.ID)
	}
	activeTasks, err = h.inspector.ListActiveTasks(tasks.OpenRatingVotingTaskType)
	if err != nil {
		log.Printf("Error listing active tasks: %v", err)
	}
	openRatingVotingTasks := make([]*asynq.TaskInfo, 0)
	for idx, t := range activeTasks {
		var p tasks.OpenRatingVotingPayload
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			log.Printf("Error unmarshaling task payload: %v", err)
			continue
		}
		if t.ID == fmt.Sprintf("%d-%d", session.ID, idx) {
			openRatingVotingTasks = append(openRatingVotingTasks, t)
			if err := h.inspector.DeleteTask(tasks.OpenRatingVotingTaskType, t.ID); err != nil {
				log.Printf("Error deleting task: %v", err)
			} else {
				log.Printf("Deleted scheduled open rating voting task: %s", t.ID)
			}
			break
		}
	}
	for idx, t := range openRatingVotingTasks {
		var p tasks.OpenRatingVotingPayload
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			log.Printf("Error unmarshaling task payload: %v", err)
			continue
		}
		err := tasks.EnqueueOpenRatingVotingTask(h.client, &tasks.EnqueueOpenRatingVotingParams{
			PollID:    p.PollID,
			MessageID: p.MessageID,
			ChatID:    p.ChatID,
			VotingID:  p.VotingID,
			Movie:     p.Movie,
			UserID:    p.UserID,
			TaskID:    fmt.Sprintf("%d-%d", session.ID, idx),
			Duration:  time.Duration(finishedAt),
		})
		if err != nil {
			log.Printf("Error scheduling new open rating voting task: %v", err)
		} else {
			log.Printf("Scheduled new open rating voting task for session: %d", session.ID)
		}
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Сессия просмотра успешно обновлена.",
	})
	if err != nil {
		log.Printf("Error sending confirmation message: %v", err)
	}
	fsmutils.DeleteMessages(ctx, b, f, userID, update.Message.Chat.ID)
	f.Reset(userID)
}
