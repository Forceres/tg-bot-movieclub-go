package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/tasks"
	fsmutils "github.com/Forceres/tg-bot-movieclub-go/internal/utils/fsm"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"github.com/go-telegram/ui/paginator"
	"github.com/hibiken/asynq"
)

const (
	statePrepareMoviesToDelete fsm.StateID = "prepare_movies_to_delete"
	stateRemove                fsm.StateID = "remove"
)

type RemoveMovieFromSessionHandler struct {
	f              *fsm.FSM
	sessionService service.ISessionService
	inspector      *asynq.Inspector
}

type IRemoveMovieFromSessionHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
	PrepareMoviesToDelete(f *fsm.FSM, args ...any)
	Remove(f *fsm.FSM, args ...any)
}

func NewRemoveMovieFromSessionHandler(sessionService service.ISessionService, inspector *asynq.Inspector, f *fsm.FSM) IRemoveMovieFromSessionHandler {
	return &RemoveMovieFromSessionHandler{
		f:              f,
		sessionService: sessionService,
		inspector:      inspector,
	}
}

func (h *RemoveMovieFromSessionHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	currentState := h.f.Current(userID)
	if currentState != stateDefault {
		return
	}
	session, err := h.sessionService.FindOngoingSession()
	if err != nil || len(session.Movies) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "ℹ️ Нет фильмов в текущей сессии.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}
	opts := []paginator.Option{
		paginator.PerPage(5),
	}
	var formattedMovies []string
	for idx, movie := range session.Movies {
		formattedMovies = append(formattedMovies, bot.EscapeMarkdown(fmt.Sprintf("%d. %s", idx+1, movie.Title)))
	}
	h.f.Set(userID, "movies", session.Movies)
	p := paginator.New(b, formattedMovies, opts...)
	showOpts := []paginator.ShowOption{}
	paginatorMsg, err := p.Show(ctx, b, update.Message.Chat.ID, showOpts...)
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Ошибка при отображении фильмов.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}
	fsmutils.AppendMessageID(h.f, userID, update.Message.ID)
	fsmutils.AppendMessageID(h.f, userID, paginatorMsg.ID)
	h.f.Transition(userID, statePrepareMoviesToDelete, userID, ctx, b, update, paginatorMsg.ID)
}

func (h *RemoveMovieFromSessionHandler) PrepareMoviesToDelete(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	paginatorMsgID := args[4].(int)
	f.Set(userID, "paginatorMsgID", paginatorMsgID)
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "📝 Перечислите номера фильмов, которые хотите удалить из сессии, через запятую.",
	})
	if err != nil {
		return
	}
	fsmutils.AppendMessageID(f, userID, msg.ID)
}

func (h *RemoveMovieFromSessionHandler) Remove(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	ids, ok := f.Get(userID, "movieIDs")
	if !ok {
		f.Reset(userID)
		return
	}

	session, err := h.sessionService.FindOngoingSession()
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "ℹ️ Нет активной сессии.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		h.f.Reset(userID)
		return
	}

	votings, err := h.sessionService.RemoveMoviesFromSession(ids.([]int64), session.ID)
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Ошибка при удалении фильмов из сессии.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		h.f.Reset(userID)
		return
	}

	for _, voting := range votings {
		var taskId string
		switch voting.Type {
		case model.VOTING_RATING_TYPE:
			taskId = fmt.Sprintf("%s-%d", tasks.CloseRatingVotingTaskType, voting.ID)
		case model.VOTING_SELECTION_TYPE:
			taskId = fmt.Sprintf("%s-%d", tasks.CloseSelectionVotingTaskType, voting.ID)
		}
		taskInfo, err := h.inspector.GetTaskInfo(tasks.QUEUE, taskId)
		if err != nil {
			continue
		}
		err = h.inspector.DeleteTask(taskInfo.Queue, taskInfo.ID)
		if err != nil {
			continue
		}
		var payload map[string]interface{}
		err = json.Unmarshal([]byte(taskInfo.Payload), &payload)
		if err != nil {
			log.Printf("Error unmarshaling task payload: %v", err)
			continue
		}
		_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    int64(payload["chat_id"].(float64)),
			MessageID: int(payload["message_id"].(float64)),
		})
		if err != nil {
			log.Printf("Error deleting message: %v", err)
		}
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "✅ Выбранные фильмы были успешно удалены из сессии.",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
	h.f.Reset(userID)
}
