package main

import (
	"log"
	"os"

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

	nodeEnv := os.Getenv("NODE_ENV")
	redisClientOpts := asynq.RedisClientOpt{Addr: cfg.Redis.URL}
	if nodeEnv == "PRODUCTION" {
		redisClientOpts.Password = cfg.Redis.Password
		redisClientOpts.Username = cfg.Redis.Username
		redisClientOpts.DB = 0
	}

	log.Printf("password: %s", redisClientOpts.Password)
	log.Printf("username: %s", redisClientOpts.Username)

	services := app.LoadServices(cfg)

	srv := asynq.NewServer(
		redisClientOpts,
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
