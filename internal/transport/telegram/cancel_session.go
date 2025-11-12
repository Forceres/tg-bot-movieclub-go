package telegram

import (
	"context"
	"fmt"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/tasks"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/hibiken/asynq"
)

type CancelSessionHandler struct {
	service       service.ISessionService
	votingService service.IVotingService
	inspector     *asynq.Inspector
}

type ICancelSessionHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewCancelSessionHandler(service service.ISessionService, votingService service.IVotingService, inspector *asynq.Inspector) ICancelSessionHandler {
	return &CancelSessionHandler{service: service, votingService: votingService, inspector: inspector}
}

func (h *CancelSessionHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	session, err := h.service.CancelSession()
	if err != nil {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка при отмене сессии.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}
	activeTasks, err := h.inspector.ListActiveTasks(tasks.FinishSessionTaskType)
	if err != nil {
		log.Printf("Error listing active tasks: %v", err)
	}
	var finishSessionTask *asynq.TaskInfo
	for _, t := range activeTasks {
		if t.ID == fmt.Sprint(session.ID) {
			finishSessionTask = t
		}
	}
	err = h.inspector.DeleteTask(tasks.FinishSessionTaskType, finishSessionTask.ID)
	if err != nil {
		log.Printf("Error deleting finish session task: %v", err)
	}
	votings, err := h.votingService.FindVotingsBySessionID(session.ID)
	if err != nil {
		log.Printf("Error finding votings by session ID: %v", err)
	}
	for _, voting := range votings {
		if voting.Status == "ongoing" {
			activeTasks, err = h.inspector.ListActiveTasks(tasks.CloseRatingVotingTaskType)
			if err != nil {
				log.Printf("Error listing active tasks: %v", err)
				continue
			}
			var closeVotingTask *asynq.TaskInfo
			for _, t := range activeTasks {
				if t.ID == fmt.Sprint(voting.ID) {
					closeVotingTask = t
				}
			}
			if closeVotingTask == nil {
				continue
			}
			err = h.inspector.DeleteTask(tasks.CloseRatingVotingTaskType, closeVotingTask.ID)
			if err != nil {
				log.Printf("Error deleting close voting task: %v", err)
			}
		}
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Сессия успешно отменена.",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
