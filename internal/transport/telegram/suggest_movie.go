package telegram

import (
	"context"
	"log"
	"strconv"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/kinopoisk"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type SuggestMovieHandler struct {
	movieService     service.IMovieService
	kinopoiskService service.IKinopoiskService
}

type ISuggestMovieHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewSuggestMovieHandler(movieService service.IMovieService, kinopoiskService service.IKinopoiskService) *SuggestMovieHandler {
	return &SuggestMovieHandler{movieService: movieService, kinopoiskService: kinopoiskService}
}

func (h *SuggestMovieHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	ids := kinopoisk.ParseIDsOrRefs(update.Message.Text)
	if len(ids) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Не найдено ссылок на фильмы Кинопоиска в сообщении.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}
	var idsToFind []int
	for _, id := range ids {
		intId, err := strconv.Atoi(id)
		if err != nil {
			continue
		}
		_, err = h.movieService.GetMovieByID(intId)
		if err != nil {
			idsToFind = append(idsToFind, intId)
		}
	}
	if len(idsToFind) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Все фильмы из вашего сообщения уже предложены ранее.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}
	if len(ids) > 5 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Слишком много фильмов в одном сообщении. Пожалуйста, отправляйте не более 5 фильмов за раз.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}
	moviesDto, err := h.kinopoiskService.SearchMovies(idsToFind, update.Message.From.FirstName)
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка при поиске фильмов на Кинопоиске.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}
	if len(moviesDto) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Не удалось найти фильмы по предоставленным ссылкам.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}
	for _, movieDto := range moviesDto {
		err := h.movieService.Create(&movieDto, update.Message.From.ID)
		if err != nil {
			log.Printf("Error while creating movie: %v", err)
			continue
		}
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Фильмы успешно добавлены в предложку!",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
