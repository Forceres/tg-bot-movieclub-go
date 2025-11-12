package main

import (
	"context"
	"log"
	"time"

	"net/http"
	"os"
	"os/signal"

	"github.com/Forceres/tg-bot-movieclub-go/internal/app"
	"github.com/Forceres/tg-bot-movieclub-go/internal/config"
	"github.com/Forceres/tg-bot-movieclub-go/internal/transport/telegram"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/date"
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

var allowedUpdates = []string{
	AllowedUpdateMessage,
	AllowedUpdateEditedMessage,
	AllowedUpdateChannelPost,
	AllowedUpdateEditedChannelPost,
	AllowedUpdateCallbackQuery,
	AllowedUpdatePoll,
	AllowedUpdatePollAnswer,
}

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
	defer services.AsynqInspector.Close()
	defaultHandler := telegram.NewDefaultHandler(f)
	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler.Handle),
		bot.WithMiddlewares(
			middlewares.Authentication,
			middlewares.Log,
		),
		bot.WithAllowedUpdates(allowedUpdates),
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
	date.InitDatePicker(b, handlers.OnDatepickerCancel, handlers.OnDatepickerSelect)
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
	// Cleanup function to delete webhook on exit
	defer func() {
		log.Println("Deleting webhook...")
		deleteCtx := context.Background()
		ok, err := b.DeleteWebhook(deleteCtx, &bot.DeleteWebhookParams{
			DropPendingUpdates: true,
		})
		if err != nil {
			log.Printf("Failed to delete webhook: %v", err)
		} else if ok {
			log.Println("Webhook deleted successfully")
		}
	}()
	date.InitDatePicker(b, handlers.OnDatepickerCancel, handlers.OnDatepickerSelect)
	app.RegisterHandlers(b, handlers, services, cfg)
	ok, err := b.SetWebhook(ctx, &bot.SetWebhookParams{
		URL:                cfg.DomainAddress,
		DropPendingUpdates: true,
		SecretToken:        cfg.Telegram.WebhookSecretToken,
		AllowedUpdates:     allowedUpdates,
	})
	if err != nil || !ok {
		log.Printf("Failed to set webhook: %v", err)
		return err
	}
	go b.StartWebhook(ctx)
	server := &http.Server{
		Addr:    ":2000",
		Handler: b.WebhookHandler(),
	}
	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		log.Println("Shutting down webhook server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Printf("Failed to start webhook server: %v", err)
		return err
	}
	return nil
}
