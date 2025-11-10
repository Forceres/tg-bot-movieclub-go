package main

import (
	"context"
	"log"

	"net/http"
	"os"
	"os/signal"

	"github.com/Forceres/tg-bot-movieclub-go/internal/app"
	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/Forceres/tg-bot-movieclub-go/internal/transport/telegram"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/fsm"
)

const stateDefault fsm.StateID = "default"

const PRODUCTION = "PRODUCTION"

const (
	AllowedUpdateMessage                 string = "message"
	AllowedUpdateEditedMessage           string = "edited_message"
	AllowedUpdateChannelPost             string = "channel_post"
	AllowedUpdateEditedChannelPost       string = "edited_channel_post"
	AllowedUpdateBusinessConnection      string = "business_connection"
	AllowedUpdateBusinessMessage         string = "business_message"
	AllowedUpdateEditedBusinessMessage   string = "edited_business_message"
	AllowedUpdateDeletedBusinessMessages string = "deleted_business_messages"
	AllowedUpdateMessageReaction         string = "message_reaction"
	AllowedUpdateMessageReactionCount    string = "message_reaction_count"
	AllowedUpdateInlineQuery             string = "inline_query"
	AllowedUpdateChosenInlineResult      string = "chosen_inline_result"
	AllowedUpdateCallbackQuery           string = "callback_query"
	AllowedUpdateShippingQuery           string = "shipping_query"
	AllowedUpdatePreCheckoutQuery        string = "pre_checkout_query"
	AllowedUpdatePurchasedPaidMedia      string = "purchased_paid_media"
	AllowedUpdatePoll                    string = "poll"
	AllowedUpdatePollAnswer              string = "poll_answer"
	AllowedUpdateMyChatMember            string = "my_chat_member"
	AllowedUpdateChatMember              string = "chat_member"
	AllowedUpdateChatJoinRequest         string = "chat_join_request"
	AllowedUpdateChatBoost               string = "chat_boost"
	AllowedUpdateRemovedChatBoost        string = "removed_chat_boost"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	log.Println("Starting Telegram Movie Club Bot...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	nodeEnv := os.Getenv("NODE_ENV")
	f := fsm.New(
		stateDefault,
		map[fsm.StateID]fsm.Callback{},
	)
	handlers, middlewares, services := app.LoadApp(cfg, f)
	defer services.AsynqClient.Close()
	defaultHandler := telegram.NewDefaultHandler(f)
	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler.Handle),
		bot.WithMiddlewares(
			middlewares.Authentication,
			middlewares.Log,
		),
		bot.WithAllowedUpdates([]string{
			AllowedUpdateMessage,
			AllowedUpdateEditedMessage,
			AllowedUpdateChannelPost,
			AllowedUpdateEditedChannelPost,
			AllowedUpdateCallbackQuery,
			AllowedUpdatePoll,
			AllowedUpdatePollAnswer,
		}),
	}
	if nodeEnv == PRODUCTION {
		err := startWebhook(ctx, opts, cfg, handlers, services)
		if err != nil {
			log.Fatalf("Failed to start webhook: %v", err)
		}
	} else {
		err := startLongPolling(ctx, opts, cfg, handlers, services)
		if err != nil {
			log.Fatalf("Failed to start long polling: %v", err)
		}
	}
}

func startLongPolling(ctx context.Context, opts []bot.Option, cfg *config.Config, handlers *app.Handlers, services *app.Services) error {
	b, err := bot.New(cfg.Telegram.BotToken, opts...)
	if err != nil {
		log.Printf("Failed to create bot: %v", err)
		return err
	}
	app.RegisterHandlers(b, handlers, services, cfg)
	b.Start(ctx)
	return nil
}

func startWebhook(ctx context.Context, opts []bot.Option, cfg *config.Config, handlers *app.Handlers, services *app.Services) error {
	opts = append(opts, bot.WithWebhookSecretToken(cfg.Telegram.WebhookSecretToken))
	b, err := bot.New(cfg.Telegram.BotToken, opts...)
	if err != nil {
		log.Printf("Failed to create bot: %v", err)
		return err
	}
	app.RegisterHandlers(b, handlers, services, cfg)
	go b.StartWebhook(ctx)
	err = http.ListenAndServe(":2000", b.WebhookHandler())
	if err != nil {
		log.Printf("Failed to start webhook server: %v", err)
		return err
	}
	return nil
}
