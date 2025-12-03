package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/tasks"
	fsmutils "github.com/Forceres/tg-bot-movieclub-go/internal/utils/fsm"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/telegram/keyboard"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"github.com/go-telegram/ui/paginator"
	"github.com/hibiken/asynq"
)

var RATING_VOTING_OPTIONS = []models.InputPollOption{
	{Text: "1"},
	{Text: "2"},
	{Text: "3"},
	{Text: "4"},
	{Text: "5"},
	{Text: "6"},
	{Text: "7"},
	{Text: "8"},
	{Text: "9"},
	{Text: "10"},
}

type VotingHandler struct {
	movieService  service.IMovieService
	votingService service.IVotingService
	pollService   service.IPollService
	voteService   service.IVoteService
	fsm           *fsm.FSM
	scheduler     *asynq.Client
}

type IVotingHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

const (
	stateDefault               fsm.StateID = "default"
	statePrepareVotingType     fsm.StateID = "prepare_voting_type"
	statePrepareVotingTitle    fsm.StateID = "prepare_voting_title"
	statePrepareVotingDuration fsm.StateID = "prepare_voting_duration"
	statePrepareMovies         fsm.StateID = "prepare_movies"
	stateStartVoting           fsm.StateID = "start_voting"
)

func NewVotingHandler(movieService service.IMovieService, votingService service.IVotingService, pollService service.IPollService, voteService service.IVoteService, f *fsm.FSM, scheduler *asynq.Client) *VotingHandler {
	return &VotingHandler{movieService: movieService, votingService: votingService, pollService: pollService, voteService: voteService, fsm: f, scheduler: scheduler}
}

func (h *VotingHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
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
		Button("Выбор фильма", []byte(model.VOTING_SELECTION_TYPE), h.onInlineKeyboardSelect).
		Button("Оценка фильма", []byte(model.VOTING_RATING_TYPE), h.onInlineKeyboardSelect).
		Row().
		Button("Отменить", []byte("cancel"), h.onCancelSelect)
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выбери тип голосования",
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *VotingHandler) onInlineKeyboardSelect(ctx context.Context, b *bot.Bot, update *models.Update, data []byte) {
	userID := update.CallbackQuery.From.ID
	currentState := h.fsm.Current(userID)
	if currentState == stateDefault {
		return
	}
	selection := string(data)
	switch selection {
	case model.VOTING_SELECTION_TYPE:
		msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   "🎬 Вы выбрали 'Выбор фильма'. Начинаем процесс выбора фильма.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		} else {
			fsmutils.AppendMessageID(h.fsm, userID, msg.ID)
		}
		h.fsm.Set(userID, "type", selection)
		h.fsm.Transition(userID, statePrepareVotingTitle, userID, ctx, b, update)
	case model.VOTING_RATING_TYPE:
		msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   "⭐ Вы выбрали 'Оценка фильма'. Начинаем процесс оценки фильма.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		} else {
			fsmutils.AppendMessageID(h.fsm, userID, msg.ID)
		}
		h.fsm.Set(userID, "type", selection)
		h.fsm.Transition(userID, statePrepareMovies, userID, ctx, b, update)
	default:
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   "❓ Неизвестный выбор.",
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		h.fsm.Reset(userID)
	}
}

func (h *VotingHandler) PrepareVotingTitle(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "📝 Введите название для голосования!",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	} else {
		fsmutils.AppendMessageID(f, userID, msg.ID)
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
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "⏱️ Введите длительность голосования (в часах)",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	} else {
		fsmutils.AppendMessageID(f, userID, msg.ID)
	}
}

func (h *VotingHandler) onCancelSelect(ctx context.Context, b *bot.Bot, update *models.Update, data []byte) {
	userID := update.CallbackQuery.From.ID
	currentState := h.fsm.Current(userID)
	if currentState == stateDefault {
		return
	}
	h.fsm.Reset(userID)
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "🚫 Отменено.",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
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
	votingType := data.(string)
	var movies [][]string
	var err error
	switch votingType {
	case model.VOTING_SELECTION_TYPE:
		movies, err = h.movieService.GetSuggestedOrWatchedMovies(true)
	case model.VOTING_RATING_TYPE:
		movies, err = h.movieService.GetSuggestedOrWatchedMovies(false)
	}
	if err != nil || len(movies) == 0 {
		var errorMsg string
		if votingType == model.VOTING_SELECTION_TYPE {
			errorMsg = "📭 Увы фильмов в предложке нет."
		} else {
			errorMsg = "📭 Увы просмотренных фильмов нет."
		}
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   errorMsg,
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		f.Reset(userID)
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
	showOpts := []paginator.ShowOption{}
	msg, err := p.Show(ctx, b, update.Message.Chat.ID, showOpts...)
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
	} else {
		fsmutils.AppendMessageID(f, userID, msg.ID)
	}
	msg, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "📝 Перечисли номера фильмов, которые должны быть оценены или выбраны через запятую",
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	} else {
		fsmutils.AppendMessageID(f, userID, msg.ID)
	}
}

func (h *VotingHandler) StartVoting(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	currentState := h.fsm.Current(userID)
	if currentState == stateDefault {
		return
	}
	duration, _ := h.fsm.Get(userID, "duration")
	votingType, _ := h.fsm.Get(userID, "type")
	title, _ := h.fsm.Get(userID, "title")
	finishedAt := time.Now().Add(time.Duration(duration.(int)) * time.Hour).Unix()
	switch votingType.(string) {
	case model.VOTING_SELECTION_TYPE:
		movies, _ := f.Get(userID, "movies")
		movieIDs := []int64{}
		selectedMovieIndexes, _ := f.Get(userID, "movieIndexes")
		pollOpts := []models.InputPollOption{}
		moviesArray := movies.([][]string)
		for _, index := range selectedMovieIndexes.([]int64) {
			// User enters 1-based index (1, 2, 3...), convert to 0-based array index
			arrayIndex := int(index) - 1
			if arrayIndex < 0 || arrayIndex >= len(moviesArray) {
				log.Printf("Index out of bounds: user entered %d, array length is %d", index, len(moviesArray))
				continue
			}
			movieData := moviesArray[arrayIndex]
			movieID, err := strconv.ParseInt(movieData[0], 10, 64)
			if err != nil {
				log.Printf("Error converting movie ID: %v", err)
				continue
			}
			movieIDs = append(movieIDs, movieID)
			var title string = movieData[1]
			parts := strings.SplitN(movieData[1], ". ", 2)
			if len(parts) == 2 {
				title = parts[1]
			}
			pollOpts = append(pollOpts, models.InputPollOption{Text: title, TextParseMode: models.ParseModeMarkdown})
		}
		multi := new(bool)
		*multi = true
		poll, err := h.votingService.StartVoting(&service.StartRatingVotingParams{
			Bot:     b,
			Context: ctx,
			ChatID:  update.Message.Chat.ID,
			Options: service.VotingOptions{
				Title:      title.(string),
				Type:       votingType.(string),
				CreatedBy:  userID,
				FinishedAt: &finishedAt,
			},
			Multi:       multi,
			PollOptions: pollOpts,
			Question:    title.(string),
		})
		if err != nil {
			log.Printf("Error while starting a voting: %v", err)
			f.Reset(userID)
			return
		}
		for optionIndex, id := range movieIDs {
			err := h.pollService.CreatePollOption(&model.PollOption{
				PollID:      poll.ID,
				OptionIndex: optionIndex,
				MovieID:     id,
			})
			if err != nil {
				log.Printf("Error saving poll option: %v", err)
				f.Reset(userID)
				return
			}
		}
		duration := time.Duration(duration.(int)) * time.Hour
		err = tasks.EnqueueCloseSelectionVotingTask(h.scheduler, duration, &tasks.CloseSelectionVotingPayload{
			PollID:    poll.PollID,
			MessageID: poll.MessageID,
			ChatID:    update.Message.Chat.ID,
			VotingID:  poll.VotingID,
			UserID:    userID,
		})
		if err != nil {
			log.Printf("Error scheduling close rating voting task: %v", err)
		}
	case model.VOTING_RATING_TYPE:
		movies, _ := f.Get(userID, "movies")
		selectedMovieIndexes, _ := f.Get(userID, "movieIndexes")
		moviesArray := movies.([][]string)
		for _, index := range selectedMovieIndexes.([]int64) {
			arrayIndex := int(index) - 1
			if arrayIndex < 0 || arrayIndex >= len(moviesArray) {
				log.Printf("Index out of bounds: user entered %d, array length is %d", index, len(moviesArray))
				continue
			}
			movieData := moviesArray[arrayIndex]
			movieID, _ := strconv.ParseInt(movieData[0], 10, 64)
			var title string = movieData[1]
			parts := strings.SplitN(movieData[1], ". ", 2)
			if len(parts) == 2 {
				title = parts[1]
			}
			title = fmt.Sprintf("Оцените фильм: %s", title)
			poll, err := h.votingService.StartVoting(&service.StartRatingVotingParams{
				Bot:     b,
				Context: ctx,
				ChatID:  update.Message.Chat.ID,
				Options: service.VotingOptions{
					Title:      title,
					Type:       votingType.(string),
					CreatedBy:  userID,
					FinishedAt: &finishedAt,
					MovieID:    &movieID,
				},
				PollOptions: RATING_VOTING_OPTIONS,
				Question:    title,
			})
			if err != nil {
				log.Printf("Error while starting a voting: %v", err)
				f.Reset(userID)
				return
			}
			duration := time.Duration(duration.(int)) * time.Hour
			err = tasks.EnqueueCloseRatingVotingTask(h.scheduler, duration, &tasks.CloseRatingVotingPayload{
				PollID:    poll.PollID,
				MessageID: poll.MessageID,
				ChatID:    update.Message.Chat.ID,
				VotingID:  poll.VotingID,
				MovieID:   movieID,
				UserID:    userID,
			})
			if err != nil {
				log.Printf("Error scheduling close rating voting task: %v", err)
			}
		}
	}
	fsmutils.DeleteMessages(ctx, b, f, userID, update.Message.From.ID)
	h.fsm.Reset(userID)
}
