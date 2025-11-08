package telegram

import (
	"context"

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
		return
	}
	messageID, ok := h.f.Get(userID, "messageID")
	if !ok {
		return
	}
	h.f.Reset(userID)
	_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    update.Message.Chat.ID,
		MessageID: messageID.(int),
	})
	if err != nil {
		return
	}
}
