package app

import (
	"log"
	"net/http"

	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/Forceres/tg-bot-movieclub-go/internal/db"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/transport/telegram"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/kinopoisk"
	permission "github.com/Forceres/tg-bot-movieclub-go/internal/utils/telegram"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/telegraph"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"github.com/hibiken/asynq"
)

const (
	statePrepareVotingType     fsm.StateID = "prepare_voting_type"
	statePrepareVotingTitle    fsm.StateID = "prepare_voting_title"
	statePrepareVotingDuration fsm.StateID = "prepare_voting_duration"
	statePrepareMovies         fsm.StateID = "prepare_movies"
	stateStartVoting           fsm.StateID = "start_voting"
	statePrepareCancelIDs      fsm.StateID = "prepare_cancel_ids"
	stateCancel                fsm.StateID = "cancel"
)

func PollAnswerMatchFunc() bot.MatchFunc {
	return func(update *models.Update) bool {
		return update != nil && update.PollAnswer != nil
	}
}

func LoadApp(cfg *config.Config, b *bot.Bot, f *fsm.FSM) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Redis.URL})
	defer client.Close()

	db, err := db.NewSqliteDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	telegraph, err := telegraph.InitTelegraph()
	if err != nil {
		log.Fatalf("Failed to initialize telegraph: %v", err)
	}

	movieRepo := repository.NewMovieRepository(db)
	movieService := service.NewMovieService(movieRepo)

	votingRepo := repository.NewVotingRepository(db)
	votingService := service.NewVotingService(votingRepo)

	pollRepo := repository.NewPollRepository(db)
	pollService := service.NewPollService(pollRepo)

	voteRepo := repository.NewVoteRepository(db)
	voteService := service.NewVoteService(voteRepo)

	kinopoiskClient := &http.Client{}
	kinopoiskAPI := kinopoisk.NewKinopoiskAPI(&cfg.Kinopoisk, kinopoiskClient)
	kinopoiskService := service.NewKinopoiskService(kinopoiskAPI)

	currentMoviesHandler := telegram.NewCurrentMoviesHandler(movieService)
	alreadyWatchedMoviesHandler := telegram.NewAlreadyWatchedMoviesHandler(movieService, telegraph)
	votingHandler := telegram.NewVotingHandler(movieService, votingService, pollService, voteService, f)
	suggestMovieHandler := telegram.NewSuggestMovieHandler(movieService, kinopoiskService)
	cancelHandler := telegram.NewCancelHandler(f)
	cancelVotingHandler := telegram.NewCancelVotingHandler(f, votingService)

	f.AddCallbacks(map[fsm.StateID]fsm.Callback{
		statePrepareVotingType:     votingHandler.PrepareVotingType,
		statePrepareVotingTitle:    votingHandler.PrepareVotingTitle,
		statePrepareVotingDuration: votingHandler.PrepareVotingDuration,
		statePrepareMovies:         votingHandler.PrepareMovies,
		stateStartVoting:           votingHandler.StartVoting,
		stateCancel:                cancelVotingHandler.Cancel,
		statePrepareCancelIDs:      cancelVotingHandler.PrepareCancelIDs,
	})

	b.RegisterHandlerMatchFunc(PollAnswerMatchFunc(), votingHandler.HandlePollAnswer)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, telegram.HelpHandler, permission.AdminOnly(cfg.Telegram.GroupID))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/now", bot.MatchTypeExact, currentMoviesHandler.Handle, permission.AdminOnly(cfg.Telegram.GroupID))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/already", bot.MatchTypeExact, alreadyWatchedMoviesHandler.Handle, permission.AdminOnly(cfg.Telegram.GroupID))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/voting", bot.MatchTypeExact, votingHandler.Handle, permission.AdminOnly(cfg.Telegram.GroupID))
	b.RegisterHandler(bot.HandlerTypeMessageText, "#предлагаю", bot.MatchTypePrefix, suggestMovieHandler.Handle)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/cancel", bot.MatchTypeExact, cancelHandler.Handle)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/cancel_voting", bot.MatchTypeExact, cancelVotingHandler.Handle)
}
