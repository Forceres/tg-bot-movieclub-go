package telegram

import (
	"context"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"github.com/go-telegram/ui/paginator"
)

type SuggestionsHandler struct {
	movieService service.IMovieService
	fsm          *fsm.FSM
}

type ISuggestionsHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

const (
	statePrepareMovieSuggestions fsm.StateID = "prepare_movie_suggestions"
)

func NewSuggestionsHandler(movieService service.IMovieService, f *fsm.FSM) *SuggestionsHandler {
	return &SuggestionsHandler{movieService: movieService, fsm: f}
}

func (h *SuggestionsHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	currentState := h.fsm.Current(userID)
	if currentState != stateDefault {
		return
	}
	h.fsm.Transition(userID, statePrepareMovieSuggestions, userID, ctx, b, update)
}

func (h *SuggestionsHandler) PrepareMovies(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	var movies [][]string
	var err error
	movies, err = h.movieService.GetSuggestedOrWatchedMovies(true)
	if err != nil || len(movies) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "📭 Увы фильмов в предложке нет.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		f.Reset(userID)
		return
	}
	opts := []paginator.Option{
		paginator.PerPage(5),
	}
	var paginatedMovies []string
	for _, movie := range movies {
		paginatedMovies = append(paginatedMovies, movie[1])
	}
	p := paginator.New(b, paginatedMovies, opts...)
	showOpts := []paginator.ShowOption{}
	_, err = p.Show(ctx, b, update.Message.Chat.ID, showOpts...)
	if err != nil {
		log.Printf("Error showing paginator: %v", err)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Ошибка при показе пагинатора.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		f.Reset(userID)
		return
	}
	f.Reset(userID)
}
