package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/Forceres/tg-bot-movieclub-go/internal/app"
	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/go-telegram/bot"
	"github.com/hibiken/asynq"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.URL},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
		},
	)

	mux := asynq.NewServeMux()

	services := app.LoadServices(cfg)

	opts := []bot.Option{
		bot.WithAllowedUpdates([]string{}),
		bot.WithSkipGetMe(),
	}

	b, err := bot.New(cfg.Telegram.BotToken, opts...)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	go b.Start(ctx)

	app.RegisterTaskProcessors(services, b, mux)

	log.Println("Starting Telegram Movie Club Worker...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
