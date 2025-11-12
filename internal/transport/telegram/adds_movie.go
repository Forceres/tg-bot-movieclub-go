package telegram

import (
	"context"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type AddsMovieHandler struct {
}

func (h *AddsMovieHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	movie_name := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/adds"))

	if movie_name == "" {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Пожалуйста, введите название фильма, например, /adds Звездные Войны",
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Название фильма: " + movie_name,
	})
}
