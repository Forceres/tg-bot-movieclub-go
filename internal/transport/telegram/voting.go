package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/tasks"
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

const (
	RATING_TYPE    = "rating"
	SELECTION_TYPE = "selection"
)

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
		Button("Выбор фильма", []byte(SELECTION_TYPE), h.onInlineKeyboardSelect).
		Button("Оценка фильма", []byte(RATING_TYPE), h.onInlineKeyboardSelect).
		Row().
		Button("Отменить", []byte("cancel"), h.onCancelSelect)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выбери тип голосования",
		ReplyMarkup: kb,
	})
}

func (h *VotingHandler) onInlineKeyboardSelect(ctx context.Context, b *bot.Bot, update *models.Update, data []byte) {
	userID := update.CallbackQuery.From.ID
	currentState := h.fsm.Current(userID)
	if currentState == stateDefault {
		return
	}
	selection := string(data)
	switch selection {
	case SELECTION_TYPE:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   "Вы выбрали 'Выбор фильма'. Начинаем процесс выбора фильма.",
		})
		h.fsm.Set(userID, "type", selection)
		h.fsm.Transition(userID, statePrepareVotingTitle, userID, ctx, b, update)
	case RATING_TYPE:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   "Вы выбрали 'Оценка фильма'. Начинаем процесс оценки фильма.",
		})
		h.fsm.Set(userID, "type", selection)
		h.fsm.Transition(userID, statePrepareVotingTitle, userID, ctx, b, update)
	default:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   "Неизвестный выбор.",
		})
		h.fsm.Transition(userID, stateDefault)
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
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "Введите название для голосования!",
	})
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
		ChatID: update.Message.Chat.ID,
		Text:   "Введите длительность голосования (в часах)",
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
	votingType := data.(string)

	switch votingType {
	case SELECTION_TYPE:
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
		showOpts := []paginator.ShowOption{}
		p.Show(ctx, b, update.Message.Chat.ID, showOpts...)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Перечисли номера фильмов, которые должны быть в голосовании через запятую.",
		})
	case RATING_TYPE:
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
		showOpts := []paginator.ShowOption{}
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
	case SELECTION_TYPE:
		voting := &model.Voting{
			Title:      title.(string),
			Type:       votingType.(string),
			CreatedBy:  userID,
			FinishedAt: &finishedAt,
		}
		createdVoting, err := h.votingService.CreateVoting(voting)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.CallbackQuery.Message.Message.Chat.ID,
				Text:   "Ошибка при создании голосования.",
			})
			h.fsm.Transition(userID, stateDefault)
			return
		}
		movies, _ := f.Get(userID, "movies")
		selectedMovieIndexes, _ := f.Get(userID, "movieIndexes")
		pollOpts := []models.InputPollOption{}
		for _, index := range selectedMovieIndexes.([]int64) {
			movieData := getByIndex(movies.([][]string), index)
			if movieData == nil {
				continue
			}
			pollOpts = append(pollOpts, models.InputPollOption{Text: (*movieData)[1]})
		}

		poll, err := b.SendPoll(ctx, &bot.SendPollParams{
			ChatID:            update.Message.Chat.ID,
			Question:          createdVoting.Title,
			Options:           pollOpts,
			IsAnonymous:       bot.False(),
			Type:              "regular",
			QuestionParseMode: models.ParseModeMarkdown,
		})
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Ошибка при создании опроса.",
			})
			h.fsm.Reset(userID)
			return
		}

		createdPoll, err := h.pollService.CreatePoll(&model.Poll{
			PollID:   poll.Poll.ID,
			VotingID: createdVoting.ID,
			Type:     "selection",
			Status:   "active",
		})
		if err != nil {
			log.Printf("Error saving poll: %v", err)
		}
		for optionIndex, index := range selectedMovieIndexes.([]int64) {
			movieData := getByIndex(movies.([][]string), index)
			if movieData == nil {
				continue
			}
			movieID, _ := strconv.Atoi((*movieData)[0])

			err := h.pollService.CreatePollOption(&model.PollOption{
				PollID:      createdPoll.ID,
				OptionIndex: optionIndex,
				MovieID:     movieID,
			})
			if err != nil {
				log.Printf("Error saving poll option: %v", err)
			}
		}
	case RATING_TYPE:
		movies, _ := f.Get(userID, "movies")
		selectedMovieIndexes, _ := f.Get(userID, "movieIndexes")
		for _, index := range selectedMovieIndexes.([]int64) {
			movieData := getByIndex(movies.([][]string), index)
			if movieData == nil {
				continue
			}
			movieID, _ := strconv.Atoi((*movieData)[0])

			voting := &model.Voting{
				Title:      title.(string),
				Type:       votingType.(string),
				CreatedBy:  userID,
				FinishedAt: &finishedAt,
				MovieID:    &movieID,
			}
			createdVoting, err := h.votingService.CreateVoting(voting)
			if err != nil {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.CallbackQuery.Message.Message.Chat.ID,
					Text:   "Ошибка при создании голосования.",
				})
				h.fsm.Reset(userID)
				return
			}

			pollMsg, err := b.SendPoll(ctx, &bot.SendPollParams{
				ChatID:            update.Message.Chat.ID,
				Question:          fmt.Sprintf("Оцените фильм: %s", (*movieData)[1]),
				Options:           RATING_VOTING_OPTIONS,
				IsAnonymous:       bot.False(),
				Type:              "regular",
				QuestionParseMode: models.ParseModeMarkdown,
			})
			if err != nil {
				log.Printf("Error sending poll: %v", err)
				continue
			}

			_, err = h.pollService.CreatePoll(&model.Poll{
				PollID:   pollMsg.Poll.ID,
				VotingID: createdVoting.ID,
				MovieID:  &movieID,
				Type:     "rating",
				Status:   "active",
			})
			if err != nil {
				log.Printf("Error saving poll: %v", err)
			}
			task, err := tasks.NewCloseRatingVotingTask(pollMsg.Poll.ID, update.Message.Chat.ID, createdVoting.ID, movieID)
			if err != nil {
				log.Printf("Error creating close rating voting task: %v", err)
				continue
			}
			scheduleOpts := []asynq.Option{asynq.MaxRetry(1), asynq.ProcessIn(time.Duration(duration.(int)) * time.Second)}
			taskInfo, err := h.scheduler.Enqueue(task, scheduleOpts...)
			if err != nil {
				log.Printf("Error scheduling voting end task: %v", err)
			}
			log.Printf("Scheduled voting end task: %s", taskInfo.ID)
		}
	}

	h.fsm.Reset(userID)
}

func getByIndex(slice [][]string, index int64) *[]string {
	for _, item := range slice {
		itemIndex, err := strconv.ParseInt(item[0], 10, 64)
		if err != nil {
			continue
		}
		if itemIndex == index {
			return &item
		}
	}
	return nil
}

func (h *VotingHandler) HandlePollAnswer(ctx context.Context, b *bot.Bot, update *models.Update) {
	poll, err := h.pollService.GetPollByPollID(update.PollAnswer.PollID)
	if err != nil {
		log.Printf("Poll not found: %s", update.PollAnswer.PollID)
		return
	}

	log.Printf("User %d voted in poll %s (type: %s)\n",
		update.PollAnswer.User.ID,
		update.PollAnswer.PollID,
		poll.Type)

	for _, optionID := range update.PollAnswer.OptionIDs {
		vote := &model.Vote{
			VotingID: poll.VotingID,
			UserID:   update.PollAnswer.User.ID,
		}

		if poll.Type == RATING_TYPE {
			rating := optionID + 1
			vote.Rating = &rating
			vote.MovieID = poll.MovieID
		} else if poll.Type == SELECTION_TYPE {
			options, err := h.pollService.GetPollOptionsByPollID(poll.ID)
			if err != nil {
				log.Printf("Error getting poll options: %v", err)
				continue
			}

			if optionID < len(options) {
				vote.MovieID = &options[optionID].MovieID
			} else {
				log.Printf("Invalid option ID: %d", optionID)
				continue
			}
		}

		if err := h.voteService.Create(vote); err != nil {
			log.Printf("Error saving vote: %v", err)
		}
	}
}
