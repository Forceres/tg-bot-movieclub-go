package telegram

import (
	"context"

	fsmutils "github.com/Forceres/tg-bot-movieclub-go/internal/utils/fsm"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
)

type CancelHandler struct {
	f *fsm.FSM
}

type ICancelHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewCancelHandler(f *fsm.FSM) ICancelHandler {
	return &CancelHandler{
		f: f,
	}
}

func (h *CancelHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	currentState := h.f.Current(userID)
	if currentState == stateDefault {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Нечего отменять!",
		})
		return
	}
	messageIDs, ok := fsmutils.GetMessageIDs(h.f, userID)
	if !ok {
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.ID,
		})
		return
	}
	_, err := b.DeleteMessages(ctx, &bot.DeleteMessagesParams{
		ChatID:     update.Message.Chat.ID,
		MessageIDs: append(messageIDs, update.Message.ID),
	})
	if err != nil {
		return
	}
	h.f.Reset(userID)
}
