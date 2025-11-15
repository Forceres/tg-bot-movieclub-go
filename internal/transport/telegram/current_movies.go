package telegram

import (
	"context"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CurrentMoviesHandler struct {
	movieService service.IMovieService
}

type ICurrentMoviesHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewCurrentMoviesHandler(movieService service.IMovieService) *CurrentMoviesHandler {
	return &CurrentMoviesHandler{movieService: movieService}
}

func (h *CurrentMoviesHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	movies, err := h.movieService.GetCurrentMovies()
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "🎬 Мы пока ничего не смотрим!",
		})
		if err != nil {
			log.Printf("Error sending the message: %v", err)
		}
		return
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      *movies,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		log.Printf("Error sending the message: %v", err)
	}
}
