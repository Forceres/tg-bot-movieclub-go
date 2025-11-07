package main

import (
	"context"
	"log"

	"net/http"
	"os"
	"os/signal"

	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/Forceres/tg-bot-movieclub-go/internal/db"
	"github.com/Forceres/tg-bot-movieclub-go/internal/repository"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/transport/telegram"
	permission "github.com/Forceres/tg-bot-movieclub-go/internal/utils/telegram"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/telegraph"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/fsm"
	"github.com/hibiken/asynq"
)

const (
	stateDefault fsm.StateID = "default"
	statePrepareVotingType  fsm.StateID = "prepare_voting_type"
	statePrepareVotingDuration  fsm.StateID = "prepare_voting_duration"
	statePrepareMovies  fsm.StateID = "prepare_movies"
)

func main() {
	log.Println("Starting Telegram Movie Club Bot...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
		panic(err)
	}
	log.Printf("Loaded Telegram Bot Token: %s", cfg.Telegram.BotToken)

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Redis.URL})
	defer client.Close()

	db, err := db.NewSqliteDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	nodeEnv := os.Getenv("NODE_ENV")

	telegraph, err := telegraph.InitTelegraph()
	if err != nil {
		log.Fatalf("Failed to initialize telegraph: %v", err)
		panic(err)
	}

	f := fsm.New(
		stateDefault,
		map[fsm.StateID]fsm.Callback{},
	)

	movieRepo := repository.NewMovieRepository(db)
	movieService := service.NewMovieService(movieRepo)

	defaultHandler := telegram.NewDefaultHandler(f)
	currentMoviesHandler := telegram.NewCurrentMoviesHandler(movieService)
	alreadyWatchedMoviesHandler := telegram.NewAlreadyWatchedMoviesHandler(movieService, telegraph)
	votingHandler := telegram.NewVotingHandler(movieService, f)

	f.AddCallbacks(map[fsm.StateID]fsm.Callback{
		statePrepareVotingType: votingHandler.PrepareVotingType,
		statePrepareVotingDuration: votingHandler.PrepareVotingDuration,
		statePrepareMovies: votingHandler.PrepareMovies,
	})
	
	if nodeEnv == "production" {
		opts := []bot.Option{
			bot.WithDefaultHandler(defaultHandler.Handle),
			bot.WithWebhookSecretToken(cfg.Telegram.WebhookSecretToken),
		}

		b, _ := bot.New(cfg.Telegram.BotToken, opts...)

		go b.StartWebhook(ctx)

		err := http.ListenAndServe(":2000", b.WebhookHandler())

		if err != nil {
			log.Fatalf("Failed to start webhook server: %v", err)
			panic(err)
		}
	} else {
		opts := []bot.Option{
			bot.WithDefaultHandler(defaultHandler.Handle),
			bot.WithMiddlewares(permission.Authentication(cfg.Telegram.GroupID)),
		}
		b, err := bot.New(cfg.Telegram.BotToken, opts...)
		if err != nil {
			log.Fatalf("Failed to create bot: %v", err)
			panic(err)
		}


		b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, telegram.HelpHandler, permission.AdminOnly(cfg.Telegram.GroupID))
		b.RegisterHandler(bot.HandlerTypeMessageText, "/now", bot.MatchTypeExact, currentMoviesHandler.Handle, permission.AdminOnly(cfg.Telegram.GroupID))
		b.RegisterHandler(bot.HandlerTypeMessageText, "/already", bot.MatchTypeExact, alreadyWatchedMoviesHandler.Handle, permission.AdminOnly(cfg.Telegram.GroupID))
		b.RegisterHandler(bot.HandlerTypeMessageText, "/voting", bot.MatchTypeExact, votingHandler.Handle, permission.AdminOnly(cfg.Telegram.GroupID))
		b.Start(ctx)
	}
}

