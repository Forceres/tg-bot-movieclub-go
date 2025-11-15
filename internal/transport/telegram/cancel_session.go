package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
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
	session, votings, err := h.service.CancelSession()
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
	taskId := fmt.Sprintf("%s-%d", tasks.FinishSessionTaskType, session.ID)
	_, err = h.inspector.GetTaskInfo(tasks.QUEUE, taskId)
	if err != nil {
		log.Printf("Error getting finish session task info: %v", err)
	}
	err = h.inspector.DeleteTask(tasks.QUEUE, taskId)
	if err != nil {
		log.Printf("Error deleting finish session task: %v", err)
	}
	for _, voting := range votings {
		if voting.Status == model.VOTING_ACTIVE_STATUS {
			taskInfo, err := h.inspector.GetTaskInfo(tasks.QUEUE, fmt.Sprintf("%s-%d", tasks.CloseRatingVotingTaskType, voting.ID))
			if err != nil {
				log.Printf("Error getting task with id: %s, %v", taskInfo.ID, err)
				continue
			}
			err = h.inspector.DeleteTask(tasks.QUEUE, taskInfo.ID)
			if err != nil {
				log.Printf("Error deleting close voting task: %v", err)
			}
			var payload tasks.CloseRatingVotingPayload
			err = json.Unmarshal([]byte(taskInfo.Payload), &payload)
			if err != nil {
				log.Printf("Error unmarshaling task payload: %v", err)
				continue
			}
			_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    payload.ChatID,
				MessageID: payload.MessageID,
			})
			if err != nil {
				log.Printf("Error deleting message: %v", err)
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
