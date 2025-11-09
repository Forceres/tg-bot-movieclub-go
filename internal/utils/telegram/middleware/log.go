package middleware

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Log(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			if update.Message.Text != "" {
				log.Printf("Received message from %s (%d): %s", update.Message.From.Username, update.Message.From.ID, update.Message.Text)
			}
		}
		next(ctx, b, update)
	}
}
