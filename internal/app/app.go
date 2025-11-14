package app

import (
	"log"
	"net/http"
	"os"

	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/Forceres/tg-bot-movieclub-go/internal/db"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/tasks"
	"github.com/Forceres/tg-bot-movieclub-go/internal/transport/telegram"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/datepicker"
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
	stateDate                  fsm.StateID = "date"
	stateTime                  fsm.StateID = "time"
	stateLocation              fsm.StateID = "location"
	stateSaveSchedule          fsm.StateID = "save_schedule"
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
	ScheduleHandler             bot.HandlerFunc
	RescheduleHandler           bot.HandlerFunc
	CancelSessionHandler        bot.HandlerFunc
	AddsMovieHandler            bot.HandlerFunc
	RescheduleSessionHandler    bot.HandlerFunc
}

type Middlewares struct {
	Authentication bot.Middleware
	AdminOnly      bot.Middleware
	Log            bot.Middleware
	Delete         bot.Middleware
}

type Services struct {
	UserService        service.IUserService
	MovieService       service.IMovieService
	KinopoiskService   service.IKinopoiskService
	VotingService      service.IVotingService
	PollService        service.IPollService
	VoteService        service.IVoteService
	ScheduleService    service.IScheduleService
	SessionService     service.ISessionService
	AsynqClient        *asynq.Client
	AsynqInspector     *asynq.Inspector
	ScheduleDatepicker *datepicker.Datepicker
	SessionDatepicker  *datepicker.Datepicker
}

func LoadApp(cfg *config.Config, f *fsm.FSM) (*Handlers, *Middlewares, *Services) {
	telegraph, err := telegraph.InitTelegraph()
	if err != nil {
		log.Fatalf("Failed to initialize telegraph: %v", err)
	}

	services := LoadServices(cfg)

	sessionDatepicker := datepicker.NewDatepicker(f)
	scheduleDatepicker := datepicker.NewDatepicker(f)

	services.SessionDatepicker = sessionDatepicker
	services.ScheduleDatepicker = scheduleDatepicker

	currentMoviesHandler := telegram.NewCurrentMoviesHandler(services.MovieService)
	alreadyWatchedMoviesHandler := telegram.NewAlreadyWatchedMoviesHandler(services.MovieService, telegraph)
	votingHandler := telegram.NewVotingHandler(services.MovieService, services.VotingService, services.PollService, services.VoteService, f, services.AsynqClient)
	suggestMovieHandler := telegram.NewSuggestMovieHandler(services.MovieService, services.KinopoiskService)
	cancelHandler := telegram.NewCancelHandler(f)
	cancelVotingHandler := telegram.NewCancelVotingHandler(f, services.VotingService)
	registerUserHandler := telegram.NewRegisterUserHandler(services.UserService)
	updateChatMemberHandler := telegram.NewUpdateChatMemberHandler(services.UserService)
	pollAnswerHandler := telegram.NewPollAnswerHandler(services.PollService, services.VoteService)
	scheduleHandler := telegram.NewScheduleHandler(services.ScheduleService, f, services.ScheduleDatepicker)
	cancelSessionHandler := telegram.NewCancelSessionHandler(services.SessionService, services.VotingService, services.AsynqInspector)
	addsMovieHandler := &telegram.AddsMovieHandler{}
	rescheduleSessionHandler := telegram.NewResheduleSessionHandler(f, services.SessionDatepicker, services.SessionService, services.AsynqInspector, services.AsynqClient)

	handlers := &Handlers{
		HelpHandler:                 telegram.HelpHandler,
		CurrentMoviesHandler:        currentMoviesHandler.Handle,
		AlreadyWatchedMoviesHandler: alreadyWatchedMoviesHandler.Handle,
		VotingHandler:               votingHandler.Handle,
		PollAnswerHandler:           pollAnswerHandler.Handle,
		SuggestMovieHandler:         suggestMovieHandler.Handle,
		CancelHandler:               cancelHandler.Handle,
		CancelVotingHandler:         cancelVotingHandler.Handle,
		RegisterUserHandler:         registerUserHandler.Handle,
		UpdateChatMemberHandler:     updateChatMemberHandler.Handle,
		ScheduleHandler:             scheduleHandler.Handle,
		RescheduleHandler:           scheduleHandler.HandleReschedule,
		CancelSessionHandler:        cancelSessionHandler.Handle,
		AddsMovieHandler:            addsMovieHandler.Handle,
		RescheduleSessionHandler:    rescheduleSessionHandler.Handle,
	}

	f.AddCallbacks(map[fsm.StateID]fsm.Callback{
		statePrepareVotingType:     votingHandler.PrepareVotingType,
		statePrepareVotingTitle:    votingHandler.PrepareVotingTitle,
		statePrepareVotingDuration: votingHandler.PrepareVotingDuration,
		statePrepareMovies:         votingHandler.PrepareMovies,
		stateStartVoting:           votingHandler.StartVoting,
		stateCancel:                cancelVotingHandler.Cancel,
		statePrepareCancelIDs:      cancelVotingHandler.PrepareCancelIDs,
		stateDate:                  scheduleHandler.PrepareDate,
		stateTime:                  scheduleHandler.PrepareTime,
		stateLocation:              scheduleHandler.PrepareLocation,
		stateSaveSchedule:          scheduleHandler.SaveSchedule,
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
	nodeEnv := os.Getenv("NODE_ENV")
	redisClientOpts := asynq.RedisClientOpt{Addr: cfg.Redis.URL}
	if nodeEnv == "PRODUCTION" {
		redisClientOpts.Password = cfg.Redis.Password
	}

	client := asynq.NewClient(redisClientOpts)
	inspector := asynq.NewInspector(redisClientOpts)

	db, err := db.NewSqliteDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	movieRepo := repository.NewMovieRepository(db)
	movieService := service.NewMovieService(movieRepo)

	pollRepo := repository.NewPollRepository(db)
	pollService := service.NewPollService(pollRepo)

	scheduleRepo := repository.NewScheduleRepository(db)
	scheduleService := service.NewScheduleService(scheduleRepo)

	sessionRepo := repository.NewSessionRepository(db)

	votingRepo := repository.NewVotingRepository(db, pollRepo, movieRepo)
	votingService := service.NewVotingService(votingRepo, scheduleService, sessionRepo, movieRepo, pollRepo)

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
		ScheduleService:  scheduleService,
		AsynqClient:      client,
		AsynqInspector:   inspector,
	}

	return services
}

func RegisterTaskProcessors(services *Services, b *bot.Bot, mux *asynq.ServeMux) {
	closeRatingVotingProcessor := tasks.NewCloseRatingVotingTaskProcessor(b, services.VotingService, services.VoteService, services.MovieService)
	closeSelectionVotingProcessor := tasks.NewCloseSelectionVotingTaskProcessor(b, services.VotingService, services.VoteService, services.MovieService, services.AsynqInspector, services.AsynqClient)
	openRatingVotingProcessor := tasks.NewOpenRatingVotingTaskProcessor(b, services.VotingService, services.MovieService, services.AsynqClient)
	finishSessionProcessor := tasks.NewFinishSessionTaskProcessor(services.SessionService)
	mux.HandleFunc(tasks.CloseRatingVotingTaskType, closeRatingVotingProcessor.Process)
	mux.HandleFunc(tasks.CloseSelectionVotingTaskType, closeSelectionVotingProcessor.Process)
	mux.HandleFunc(tasks.OpenRatingVotingTaskType, openRatingVotingProcessor.Process)
	mux.HandleFunc(tasks.FinishSessionTaskType, finishSessionProcessor.Process)
}

func RegisterHandlers(b *bot.Bot, handlers *Handlers, services *Services, cfg *config.Config) {
	b.RegisterHandlerMatchFunc(PollAnswerMatchFunc(), handlers.PollAnswerHandler)
	b.RegisterHandlerMatchFunc(telegram.UpdateChatMemberMatchFunc(), handlers.UpdateChatMemberHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "#предлагаю", bot.MatchTypePrefix, handlers.SuggestMovieHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "#расписание", bot.MatchTypeExact, handlers.ScheduleHandler, middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "#перенос", bot.MatchTypeExact, handlers.RescheduleHandler, middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "help", bot.MatchTypeCommand, handlers.HelpHandler, middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "now", bot.MatchTypeCommand, handlers.CurrentMoviesHandler, middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "already", bot.MatchTypeCommand, handlers.AlreadyWatchedMoviesHandler, middleware.AdminOnly(cfg.Telegram.GroupID, services.UserService), middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "voting", bot.MatchTypeCommand, handlers.VotingHandler, middleware.AdminOnly(cfg.Telegram.GroupID, services.UserService), middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "cancel", bot.MatchTypeCommand, handlers.CancelHandler, middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "cancel_voting", bot.MatchTypeCommand, handlers.CancelVotingHandler, middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "cancel_session", bot.MatchTypeCommand, handlers.CancelSessionHandler, middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "register", bot.MatchTypeCommand, handlers.RegisterUserHandler, middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "schedule", bot.MatchTypeCommand, handlers.RescheduleHandler, middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "adds", bot.MatchTypeCommand, handlers.AddsMovieHandler, middleware.Delete)
}
