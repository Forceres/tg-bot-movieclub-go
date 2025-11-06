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
	"github.com/go-telegram/bot"
	"github.com/hibiken/asynq"
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

	// Initialize database
	db, err := db.NewSqliteDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		panic(err)
	}

	println(db)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	nodeEnv := os.Getenv("NODE_ENV")

	movieRepo := repository.NewMovieRepository(db)
	movieService := service.NewMovieService(movieRepo)

	currentMoviesHandler := telegram.NewCurrentMoviesHandler(movieService)

	if nodeEnv == "production" {
		opts := []bot.Option{
			bot.WithDefaultHandler(telegram.DefaultHandler),
			bot.WithWebhookSecretToken(cfg.Telegram.WebhookSecretToken),
		}

		b, _ := bot.New(cfg.Telegram.BotToken, opts...)

		// call methods.SetWebhook if needed

		go b.StartWebhook(ctx)

		err := http.ListenAndServe(":2000", b.WebhookHandler())

		if err != nil {
			log.Fatalf("Failed to start webhook server: %v", err)
			panic(err)
		}
	} else {
		opts := []bot.Option{
			bot.WithDefaultHandler(telegram.DefaultHandler),
		}
		b, err := bot.New(cfg.Telegram.BotToken, opts...)
		if err != nil {
			log.Fatalf("Failed to create bot: %v", err)
			panic(err)
		}

		b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, permission.AdminOnly(cfg.Telegram.GroupID, telegram.HelpHandler))
		b.RegisterHandler(bot.HandlerTypeMessageText, "/now", bot.MatchTypeExact, permission.AdminOnly(cfg.Telegram.GroupID, currentMoviesHandler.Handle))
		b.Start(ctx)
	}
}

