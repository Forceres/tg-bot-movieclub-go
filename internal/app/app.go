package app

import (
	"log"
	"net/http"

	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/Forceres/tg-bot-movieclub-go/internal/db"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/tasks"
	"github.com/Forceres/tg-bot-movieclub-go/internal/transport/telegram"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/kinopoisk"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/telegram/middleware"
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

type Handlers struct {
	HelpHandler                 bot.HandlerFunc
	CurrentMoviesHandler        bot.HandlerFunc
	AlreadyWatchedMoviesHandler bot.HandlerFunc
	VotingHandler               bot.HandlerFunc
	PollAnswerHandler           bot.HandlerFunc
	SuggestMovieHandler         bot.HandlerFunc
	CancelHandler               bot.HandlerFunc
	CancelVotingHandler         bot.HandlerFunc
	RegisterUserHandler         bot.HandlerFunc
	UpdateChatMemberHandler     bot.HandlerFunc
}

type Middlewares struct {
	Authentication bot.Middleware
	AdminOnly      bot.Middleware
	Log            bot.Middleware
	Delete         bot.Middleware
}

type Services struct {
	UserService      service.IUserService
	MovieService     service.IMovieService
	KinopoiskService service.IKinopoiskService
	VotingService    service.IVotingService
	PollService      service.IPollService
	VoteService      service.IVoteService
	AsynqClient      *asynq.Client
}

func LoadApp(cfg *config.Config, f *fsm.FSM) (*Handlers, *Middlewares, *Services) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Redis.URL})

	telegraph, err := telegraph.InitTelegraph()
	if err != nil {
		log.Fatalf("Failed to initialize telegraph: %v", err)
	}

	services := LoadServices(cfg)

	currentMoviesHandler := telegram.NewCurrentMoviesHandler(services.MovieService)
	alreadyWatchedMoviesHandler := telegram.NewAlreadyWatchedMoviesHandler(services.MovieService, telegraph)
	votingHandler := telegram.NewVotingHandler(services.MovieService, services.VotingService, services.PollService, services.VoteService, f, client)
	suggestMovieHandler := telegram.NewSuggestMovieHandler(services.MovieService, services.KinopoiskService)
	cancelHandler := telegram.NewCancelHandler(f)
	cancelVotingHandler := telegram.NewCancelVotingHandler(f, services.VotingService)
	registerUserHandler := telegram.NewRegisterUserHandler(services.UserService)
	updateChatMemberHandler := telegram.NewUpdateChatMemberHandler(services.UserService)

	handlers := &Handlers{
		HelpHandler:                 telegram.HelpHandler,
		CurrentMoviesHandler:        currentMoviesHandler.Handle,
		AlreadyWatchedMoviesHandler: alreadyWatchedMoviesHandler.Handle,
		VotingHandler:               votingHandler.Handle,
		PollAnswerHandler:           votingHandler.HandlePollAnswer,
		SuggestMovieHandler:         suggestMovieHandler.Handle,
		CancelHandler:               cancelHandler.Handle,
		CancelVotingHandler:         cancelVotingHandler.Handle,
		RegisterUserHandler:         registerUserHandler.Handle,
		UpdateChatMemberHandler:     updateChatMemberHandler.Handle,
	}

	f.AddCallbacks(map[fsm.StateID]fsm.Callback{
		statePrepareVotingType:     votingHandler.PrepareVotingType,
		statePrepareVotingTitle:    votingHandler.PrepareVotingTitle,
		statePrepareVotingDuration: votingHandler.PrepareVotingDuration,
		statePrepareMovies:         votingHandler.PrepareMovies,
		stateStartVoting:           votingHandler.StartVoting,
		stateCancel:                cancelVotingHandler.Cancel,
		statePrepareCancelIDs:      cancelVotingHandler.PrepareCancelIDs,
	})

	middlewares := &Middlewares{
		Authentication: middleware.Authentication(cfg.Telegram.GroupID, services.UserService),
		AdminOnly:      middleware.AdminOnly(cfg.Telegram.GroupID, services.UserService),
		Log:            middleware.Log,
		Delete:         middleware.Delete,
	}

	return handlers, middlewares, services
}

func LoadServices(cfg *config.Config) *Services {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Redis.URL})

	db, err := db.NewSqliteDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	movieRepo := repository.NewMovieRepository(db)
	movieService := service.NewMovieService(movieRepo)

	pollRepo := repository.NewPollRepository(db)
	pollService := service.NewPollService(pollRepo)

	votingRepo := repository.NewVotingRepository(db, pollRepo, movieRepo)
	votingService := service.NewVotingService(votingRepo)

	voteRepo := repository.NewVoteRepository(db)
	voteService := service.NewVoteService(voteRepo)

	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, roleRepo)

	kinopoiskClient := &http.Client{}
	kinopoiskAPI := kinopoisk.NewKinopoiskAPI(&cfg.Kinopoisk, kinopoiskClient)
	kinopoiskService := service.NewKinopoiskService(kinopoiskAPI)

	services := &Services{
		UserService:      userService,
		MovieService:     movieService,
		KinopoiskService: kinopoiskService,
		VotingService:    votingService,
		PollService:      pollService,
		VoteService:      voteService,
		AsynqClient:      client,
	}

	return services
}

func RegisterTaskProcessors(services *Services, b *bot.Bot, mux *asynq.ServeMux) {
	closeRatingVotingProcessor := tasks.NewCloseRatingVotingTaskProcessor(b, services.VotingService, services.VoteService, services.MovieService)
	closeSelectionVotingProcessor := tasks.NewCloseSelectionVotingTaskProcessor(b, services.VotingService, services.VoteService, services.MovieService)
	mux.HandleFunc(tasks.CloseRatingVotingTaskType, closeRatingVotingProcessor.Process)
	mux.HandleFunc(tasks.CloseSelectionVotingTaskType, closeSelectionVotingProcessor.Process)

}

func RegisterHandlers(b *bot.Bot, handlers *Handlers, services *Services, cfg *config.Config) {
	b.RegisterHandlerMatchFunc(PollAnswerMatchFunc(), handlers.PollAnswerHandler)
	b.RegisterHandlerMatchFunc(telegram.UpdateChatMemberMatchFunc(), handlers.UpdateChatMemberHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "help", bot.MatchTypeCommand, handlers.HelpHandler, middleware.AdminOnly(cfg.Telegram.GroupID, services.UserService))
	b.RegisterHandler(bot.HandlerTypeMessageText, "now", bot.MatchTypeCommand, handlers.CurrentMoviesHandler, middleware.AdminOnly(cfg.Telegram.GroupID, services.UserService))
	b.RegisterHandler(bot.HandlerTypeMessageText, "already", bot.MatchTypeCommand, handlers.AlreadyWatchedMoviesHandler, middleware.AdminOnly(cfg.Telegram.GroupID, services.UserService))
	b.RegisterHandler(bot.HandlerTypeMessageText, "voting", bot.MatchTypeCommand, handlers.VotingHandler, middleware.AdminOnly(cfg.Telegram.GroupID, services.UserService))
	b.RegisterHandler(bot.HandlerTypeMessageText, "#предлагаю", bot.MatchTypePrefix, handlers.SuggestMovieHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "cancel", bot.MatchTypeCommand, handlers.CancelHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "cancel_voting", bot.MatchTypeCommand, handlers.CancelVotingHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "register", bot.MatchTypeCommand, handlers.RegisterUserHandler)
}
