package telegram

import (
	"context"
	"fmt"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/telegram/keyboard"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"github.com/go-telegram/ui/paginator"
)

type VotingHandler struct {
	movieService service.IMovieService
	fsm *fsm.FSM
}

type IVotingHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}


const (
	stateDefault fsm.StateID = "default"
	statePrepareVotingType  fsm.StateID = "prepare_voting_type"
	statePrepareVotingDuration  fsm.StateID = "prepare_voting_duration"
	statePrepareMovies  fsm.StateID = "prepare_movies"
)

func NewVotingHandler(movieService service.IMovieService, f *fsm.FSM) *VotingHandler {
	return &VotingHandler{movieService: movieService, fsm: f}
}

func (h *VotingHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	fmt.Println(userID)
	h.fsm.Transition(userID, statePrepareVotingType, userID, ctx, b, update)
}

func (h *VotingHandler) PrepareVotingType(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	opts := []keyboard.Option{}
	kb := keyboard.New(b, opts...).
		Row().
		Button("Выбор фильма", []byte("Выбор фильма"), h.onInlineKeyboardSelect).
		Button("Оценка фильма", []byte("Оценка фильма"), h.onInlineKeyboardSelect).
		Row().
		Button("Cancel", []byte("cancel"), h.onCancelSelect)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: "Выбери тип голосования",
		ReplyMarkup: kb,
	})
}

func (h *VotingHandler) onInlineKeyboardSelect(ctx context.Context, b *bot.Bot, update *models.Update, data []byte) {
	userID := update.CallbackQuery.From.ID
	currentState := h.fsm.Current(userID)
	fmt.Println(userID)
	if currentState == stateDefault {
		return
	}
	selection := string(data)
	switch selection {
		case "Выбор фильма":
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.CallbackQuery.Message.Message.Chat.ID,
				Text:   "Вы выбрали 'Выбор фильма'. Начинаем процесс выбора фильма.",
			})
			h.fsm.Set(userID, "type", data)
			h.fsm.Transition(userID, statePrepareVotingDuration, userID, ctx, b, update)
		case "Оценка фильма":
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.CallbackQuery.Message.Message.Chat.ID,
				Text:   "Вы выбрали 'Оценка фильма'. Начинаем процесс оценки фильма.",
			})
			h.fsm.Set(userID, "type", data)
			h.fsm.Transition(userID, statePrepareVotingDuration, userID, ctx, b, update)
		default:
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.CallbackQuery.Message.Message.Chat.ID,
				Text:   "Неизвестный выбор.",
			})
			h.fsm.Transition(userID, stateDefault)
	}
}

func (h *VotingHandler) PrepareVotingDuration(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		Text:        "Введите длительность голосования (в часах)",
	})
}

func (h *VotingHandler) onCancelSelect(ctx context.Context, b *bot.Bot, update *models.Update, data []byte) {
	userID := update.CallbackQuery.From.ID
	currentState := h.fsm.Current(userID)
	if currentState == stateDefault {
		return
	}
	h.fsm.Transition(userID, stateDefault)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "Отменено.",
	})
}

func (h *VotingHandler) PrepareMovies(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	data, _ := f.Get(userID, "type")
	votingType := string(data.([]byte))

	switch votingType {
		case "Выбор фильма":
			movies, err := h.movieService.GetSuggestedOrWatchedMovies(true)
			if err != nil || len(movies) == 0 {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Увы в предложке ничего нет.",
				})
				f.Transition(userID, stateDefault)
				return
			}
			f.Set(userID, "movies", movies)
			opts := []paginator.Option{
				paginator.PerPage(5),
			}

			var paginatedMovies []string
			for _, movie := range movies {
				paginatedMovies = append(paginatedMovies, movie[1])
			}

			p := paginator.New(b, paginatedMovies, opts...)
			showOpts := []paginator.ShowOption{paginator.ShowWithThreadID(1)}
			p.Show(ctx, b, update.Message.Chat.ID, showOpts...)
			b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Перечисли номера фильмов, которые должны быть в голосовании через запятую.",
				})
		case "Оценка фильма":
			movies, err := h.movieService.GetSuggestedOrWatchedMovies(false)
			if err != nil || len(movies) == 0 {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Увы фильмов нет.",
				})
				f.Transition(userID, stateDefault)
				return
			}
			f.Set(userID, "movies", movies)
			opts := []paginator.Option{
				paginator.PerPage(5),
			}

			var paginatedMovies []string
			for _, movie := range movies {
				paginatedMovies = append(paginatedMovies, movie[1])
			}

			p := paginator.New(b, paginatedMovies, opts...)
			showOpts := []paginator.ShowOption{paginator.ShowWithThreadID(1)}
			_, err = p.Show(ctx, b, update.Message.Chat.ID, showOpts...)
			if err != nil {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Ошибка при показе пагинатора.",
				})
				log.Fatalln(err)
				f.Transition(userID, stateDefault)
				return
			}
			b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Перечисли номера фильмов, которые должны быть оценены через запятую",
				})
		default:
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Неизвестный тип голосования.",
			})
			f.Transition(userID, stateDefault)
			return
	}
}

func (h *VotingHandler) StartVoting(ctx context.Context, b *bot.Bot, update *models.Update, data []byte) {
	userID := update.CallbackQuery.From.ID
	currentState := h.fsm.Current(userID)
	if currentState == stateDefault {
		return
	}
}