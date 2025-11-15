package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/tasks"
	fsmutils "github.com/Forceres/tg-bot-movieclub-go/internal/utils/fsm"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"github.com/hibiken/asynq"
)

const stateRescheduleSession fsm.StateID = "reschedule_session"

type ResheduleSessionHandler struct {
	f              *fsm.FSM
	sessionService service.ISessionService
	inspector      *asynq.Inspector
	client         *asynq.Client
}

type IRescheduleSessionHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
	RescheduleSession(f *fsm.FSM, args ...any)
}

func NewResheduleSessionHandler(f *fsm.FSM, sessionService service.ISessionService, inspector *asynq.Inspector, client *asynq.Client) IRescheduleSessionHandler {
	return &ResheduleSessionHandler{f: f, sessionService: sessionService, inspector: inspector, client: client}
}

func (h *ResheduleSessionHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	currentState := h.f.Current(userID)
	if currentState != stateDefault {
		return
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Обновление даты сессии просмотра...",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}
	h.f.Set(userID, "datepicker", "session")
	h.f.Transition(userID, stateDate, userID, ctx, b, update)
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
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Нет текущей сессии просмотра.",
		})
		if err != nil {
			log.Printf("Error sending error message: %v", err)
		}
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
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка при обновлении сессии просмотра.",
		})
		if err != nil {
			log.Printf("Error sending error message: %v", err)
		}
		f.Reset(userID)
		return
	}
	taskId := fmt.Sprintf("%s-%d", tasks.FinishSessionTaskType, session.ID)
	_, err = h.inspector.GetTaskInfo(tasks.QUEUE, taskId)
	if err != nil {
		log.Printf("Error getting task info: %v", err)
	}
	if err := h.inspector.DeleteTask(tasks.QUEUE, taskId); err != nil {
		log.Printf("Error deleting task: %v", err)
	} else {
		log.Printf("Deleted scheduled finish session task: %s", taskId)
	}
	err = tasks.EnqueueFinishSessionTask(h.client, &tasks.EnqueueFinishSessionParams{
		SessionID: session.ID,
		Duration:  time.Duration(finishedAt) - time.Duration(time.Now().Unix()),
	})
	if err != nil {
		log.Printf("Error scheduling new finish session task: %v", err)
	} else {
		log.Printf("Scheduled new finish session task for session: %d", session.ID)
	}
	taskIds := make([]string, 0)
	for _, movie := range session.Movies {
		taskIds = append(taskIds, fmt.Sprintf("%s-%d-%d", tasks.OpenRatingVotingTaskType, session.ID, movie.ID))
	}
	openRatingVotingTasks := make([]*asynq.TaskInfo, 0)
	for _, tId := range taskIds {
		taskInfo, err := h.inspector.GetTaskInfo(tasks.QUEUE, taskId)
		if err != nil {
			log.Printf("Error getting task info: %v", err)
		}
		openRatingVotingTasks = append(openRatingVotingTasks, taskInfo)
		if err := h.inspector.DeleteTask(tasks.QUEUE, tId); err != nil {
			log.Printf("Error deleting task: %v", err)
		} else {
			log.Printf("Deleted scheduled open rating voting task: %s", tId)
		}
	}
	for _, t := range openRatingVotingTasks {
		var p tasks.OpenRatingVotingPayload
		if err := json.Unmarshal(t.Payload, &p); err != nil {
			log.Printf("Error unmarshaling task payload: %v", err)
			continue
		}
		err := tasks.EnqueueOpenRatingVotingTask(h.client, &tasks.EnqueueOpenRatingVotingParams{
			ChatID:    p.ChatID,
			SessionID: session.ID,
			Movie:     p.Movie,
			UserID:    p.UserID,
			TaskID:    t.ID,
			Duration:  time.Duration(finishedAt) - time.Duration(time.Now().Unix()),
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
