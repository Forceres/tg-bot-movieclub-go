package app

import (
	"log"
	"net/http"
	"strings"

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
	stateRescheduleSession     fsm.StateID = "reschedule_session"
	statePrepareMoviesToDelete fsm.StateID = "prepare_movies_to_delete"
	stateRemove                fsm.StateID = "remove"
	stateDescription           fsm.StateID = "description"
	stateSaveDescription       fsm.StateID = "save_description"
)

func PollAnswerMatchFunc() bot.MatchFunc {
	return func(update *models.Update) bool {
		return update != nil && update.PollAnswer != nil
	}
}

type Handlers struct {
	HelpHandler                     bot.HandlerFunc
	CurrentMoviesHandler            bot.HandlerFunc
	AlreadyWatchedMoviesHandler     bot.HandlerFunc
	VotingHandler                   bot.HandlerFunc
	PollAnswerHandler               bot.HandlerFunc
	SuggestMovieHandler             bot.HandlerFunc
	CancelHandler                   bot.HandlerFunc
	CancelVotingHandler             bot.HandlerFunc
	RegisterUserHandler             bot.HandlerFunc
	UpdateChatMemberHandler         bot.HandlerFunc
	ScheduleHandler                 bot.HandlerFunc
	RescheduleHandler               bot.HandlerFunc
	CancelSessionHandler            bot.HandlerFunc
	RescheduleSessionHandler        bot.HandlerFunc
	RemoveMovieFromSessionHandler   bot.HandlerFunc
	AddMovieToSessionHandler        bot.HandlerFunc
	CustomSessionDescriptionHandler bot.HandlerFunc
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
	cancelVotingHandler := telegram.NewCancelVotingHandler(f, services.VotingService, services.AsynqInspector)
	registerUserHandler := telegram.NewRegisterUserHandler(services.UserService)
	updateChatMemberHandler := telegram.NewUpdateChatMemberHandler(services.UserService)
	pollAnswerHandler := telegram.NewPollAnswerHandler(services.PollService, services.VoteService)
	scheduleHandler := telegram.NewScheduleHandler(services.ScheduleService, f, services.ScheduleDatepicker, services.SessionDatepicker)
	cancelSessionHandler := telegram.NewCancelSessionHandler(services.SessionService, services.VotingService, services.AsynqInspector)
	rescheduleSessionHandler := telegram.NewResheduleSessionHandler(f, services.SessionService, services.AsynqInspector, services.AsynqClient)
	removeMovieFromSessionHandler := telegram.NewRemoveMovieFromSessionHandler(services.SessionService, services.AsynqInspector, f)
	customSessionDescriptionHandler := telegram.NewCustomSessionDescriptionHandler(services.SessionService, f)
	addMovieToSessionHandler := telegram.NewAddMovieToSessionHandler(services.MovieService, services.KinopoiskService, services.SessionService, services.PollService, services.AsynqClient, services.AsynqInspector)

	handlers := &Handlers{
		HelpHandler:                     telegram.HelpHandler,
		CurrentMoviesHandler:            currentMoviesHandler.Handle,
		AlreadyWatchedMoviesHandler:     alreadyWatchedMoviesHandler.Handle,
		VotingHandler:                   votingHandler.Handle,
		PollAnswerHandler:               pollAnswerHandler.Handle,
		SuggestMovieHandler:             suggestMovieHandler.Handle,
		CancelHandler:                   cancelHandler.Handle,
		CancelVotingHandler:             cancelVotingHandler.Handle,
		RegisterUserHandler:             registerUserHandler.Handle,
		UpdateChatMemberHandler:         updateChatMemberHandler.Handle,
		ScheduleHandler:                 scheduleHandler.Handle,
		RescheduleHandler:               scheduleHandler.HandleReschedule,
		CancelSessionHandler:            cancelSessionHandler.Handle,
		RescheduleSessionHandler:        rescheduleSessionHandler.Handle,
		RemoveMovieFromSessionHandler:   removeMovieFromSessionHandler.Handle,
		AddMovieToSessionHandler:        addMovieToSessionHandler.Handle,
		CustomSessionDescriptionHandler: customSessionDescriptionHandler.Handle,
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
		stateRescheduleSession:     rescheduleSessionHandler.RescheduleSession,
		statePrepareMoviesToDelete: removeMovieFromSessionHandler.PrepareMoviesToDelete,
		stateRemove:                removeMovieFromSessionHandler.Remove,
		stateDescription:           customSessionDescriptionHandler.HandleDescriptionInput,
		stateSaveDescription:       customSessionDescriptionHandler.SaveDescription,
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
	connOpt, err := asynq.ParseRedisURI(cfg.Redis.URL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URI: %v", err)
	}

	client := asynq.NewClient(connOpt)
	inspector := asynq.NewInspector(connOpt)

	db, err := db.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	sessionRepo := repository.NewSessionRepository(db)
	movieRepo := repository.NewMovieRepository(db)
	pollRepo := repository.NewPollRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	votingRepo := repository.NewVotingRepository(db, pollRepo, movieRepo)
	voteRepo := repository.NewVoteRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)

	movieService := service.NewMovieService(movieRepo, sessionRepo)

	pollService := service.NewPollService(pollRepo)

	scheduleService := service.NewScheduleService(scheduleRepo)

	sessionService := service.NewSessionService(sessionRepo, movieRepo, votingRepo, scheduleService)

	votingService := service.NewVotingService(votingRepo, scheduleService, sessionRepo, movieRepo, pollRepo)

	voteService := service.NewVoteService(voteRepo)

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
		SessionService:   sessionService,
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
	b.RegisterHandlerMatchFunc(telegram.UpdateChatMemberMatchFunc(cfg.Telegram.GroupID), handlers.UpdateChatMemberHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "#предлагаю", bot.MatchTypePrefix, handlers.SuggestMovieHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "#расписание", bot.MatchTypeExact, handlers.ScheduleHandler, middleware.Delete)
	b.RegisterHandler(bot.HandlerTypeMessageText, "#перенос", bot.MatchTypeExact, handlers.RescheduleSessionHandler, middleware.Delete)
	registerCommandHandler(b, "help", handlers.HelpHandler, middleware.Delete)
	registerCommandHandler(b, "now", handlers.CurrentMoviesHandler, middleware.Delete)
	registerCommandHandler(b, "already", handlers.AlreadyWatchedMoviesHandler, middleware.AdminOnly(cfg.Telegram.GroupID, services.UserService), middleware.Delete)
	registerCommandHandler(b, "voting", handlers.VotingHandler, middleware.AdminOnly(cfg.Telegram.GroupID, services.UserService), middleware.Delete)
	registerCommandHandler(b, "cancel", handlers.CancelHandler, middleware.Delete)
	registerCommandHandler(b, "cancel_voting", handlers.CancelVotingHandler, middleware.Delete)
	registerCommandHandler(b, "cancel_session", handlers.CancelSessionHandler, middleware.Delete)
	registerCommandHandler(b, "register", handlers.RegisterUserHandler, middleware.Delete)
	registerCommandHandler(b, "schedule", handlers.RescheduleHandler, middleware.Delete)
	registerCommandHandler(b, "add", handlers.AddMovieToSessionHandler, middleware.Delete)
	registerCommandHandler(b, "start", handlers.RegisterUserHandler, middleware.Delete)
	registerCommandHandler(b, "rm", handlers.RemoveMovieFromSessionHandler, middleware.Delete)
	registerCommandHandler(b, "custom", handlers.CustomSessionDescriptionHandler, middleware.Delete)
}

func registerCommandHandler(bot *bot.Bot, command string, handler bot.HandlerFunc, middlewares ...bot.Middleware) {
	commandText := "/" + command
	commandTextPrefix := commandText + "@"
	matchFunc := func(update *models.Update) bool {
		if update.Message != nil {
			for _, e := range update.Message.Entities {
				if e.Offset == 0 {
					part := update.Message.Text[e.Offset : e.Offset+e.Length]
					if part == commandText || strings.HasPrefix(part, commandTextPrefix) {
						return true
					}
				}
			}
		}
		return false
	}
	bot.RegisterHandlerMatchFunc(matchFunc, handler, middlewares...)
}
