package telegram

import (
	"context"

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
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: "Мы пока ничего не смотрим!",
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text: *movies,
			ParseMode: models.ParseModeHTML,
	})
}