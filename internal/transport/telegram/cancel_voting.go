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

const statePrepareCancelIDs fsm.StateID = "prepare_cancel_ids"
const stateCancel fsm.StateID = "cancel"

type CancelVotingHandler struct {
	f             *fsm.FSM
	votingService service.IVotingService
	inspector     *asynq.Inspector
}

type ICancelVotingHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
	PrepareCancelIDs(f *fsm.FSM, args ...any)
	Cancel(f *fsm.FSM, args ...any)
}

func NewCancelVotingHandler(f *fsm.FSM, votingService service.IVotingService, inspector *asynq.Inspector) ICancelVotingHandler {
	return &CancelVotingHandler{
		f:             f,
		votingService: votingService,
		inspector:     inspector,
	}
}

func (h *CancelVotingHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	currentState := h.f.Current(userID)
	if currentState != stateDefault {
		return
	}
	votings, err := h.votingService.FindVotingByStatus(model.VOTING_ACTIVE_STATUS)
	if err != nil || len(votings) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "ℹ️ Нет активных голосований.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}
	opts := []paginator.Option{
		paginator.PerPage(5),
	}
	var formattedVotings []string
	for idx, voting := range votings {
		formattedVotings = append(formattedVotings, bot.EscapeMarkdown(fmt.Sprintf("%d. %s", idx+1, voting.Title)))
	}
	h.f.Set(userID, "votings", votings)
	p := paginator.New(b, formattedVotings, opts...)
	showOpts := []paginator.ShowOption{}
	paginatorMsg, err := p.Show(ctx, b, update.Message.Chat.ID, showOpts...)
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Ошибка при отображении голосований.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}
	fsmutils.AppendMessageID(h.f, userID, update.Message.ID)
	fsmutils.AppendMessageID(h.f, userID, paginatorMsg.ID)
	h.f.Transition(userID, statePrepareCancelIDs, userID, ctx, b, update, paginatorMsg.ID)
}

func (h *CancelVotingHandler) PrepareCancelIDs(f *fsm.FSM, args ...any) {
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
		Text:   "📝 Перечислите номера голосований, которые хотите отменить, через запятую.",
	})
	if err != nil {
		return
	}
	fsmutils.AppendMessageID(f, userID, msg.ID)
}

func (h *CancelVotingHandler) Cancel(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	ids, ok := f.Get(userID, "votingIDs")
	if !ok {
		f.Reset(userID)
		return
	}

	votings, err := h.votingService.CancelByVotingID(ids.([]int64))
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Ошибка при отмене голосований.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		f.Reset(userID)
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
		Text:   "✅ Выбранные голосования были отменены.",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
	h.f.Reset(userID)
}
