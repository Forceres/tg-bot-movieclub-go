package main

import (
	"context"
	"log"

	"net/http"
	"os"
	"os/signal"

	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/Forceres/tg-bot-movieclub-go/internal/transport/telegram"
	"github.com/go-telegram/bot"
)

func main() {
	log.Println("Starting Telegram Movie Club Bot...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
		panic(err)
	}
	log.Printf("Loaded Telegram Bot Token: %s", cfg.Telegram.BotToken)

	nodeEnv := os.Getenv("NODE_ENV")

	if nodeEnv == "production" {

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

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
		b, err := bot.New(cfg.Telegram.BotToken, bot.WithDefaultHandler(telegram.DefaultHandler))
		if err != nil {
			log.Fatalf("Failed to create bot: %v", err)
			panic(err)
		}

		b.Start(context.TODO())
	}
}

