package main

import (
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/app"
	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/go-telegram/bot"
	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	connOpt, err := asynq.ParseRedisURI(cfg.Redis.URL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URI: %v", err)
	}

	services := app.LoadServices(cfg)

	srv := asynq.NewServer(
		connOpt,
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
		},
	)

	mux := asynq.NewServeMux()

	opts := []bot.Option{
		bot.WithAllowedUpdates([]string{}),
		bot.WithSkipGetMe(),
	}

	b, err := bot.New(cfg.Telegram.BotToken, opts...)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	app.RegisterTaskProcessors(services, b, mux)

	log.Println("Starting Telegram Movie Club Worker...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
