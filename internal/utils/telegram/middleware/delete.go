package middleware

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Delete(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil && update.Message.From != nil && update.Message.From.ID != b.ID() {
			_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			})
			if err != nil {
				log.Printf("Failed to delete message (by id) %d in chat %d: %v", update.Message.ID, update.Message.Chat.ID, err)
			} else {
				log.Printf("Deleted message (by id) %d in chat %d", update.Message.ID, update.Message.Chat.ID)
			}
		}
		next(ctx, b, update)
	}
}
