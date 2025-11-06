package config

type TelegramConfig struct {
	BotToken string `env:"TELEGRAM_BOT_TOKEN" env-required:"true"`
	WebhookSecretToken string `env:"TELEGRAM_WEBHOOK_SECRET_TOKEN"`
	GroupID int64 `env:"TELEGRAM_GROUP_ID" env-required:"true"`
}