package config

type AppConfig struct {
	LogLevel string `env:"APP_LOG_LEVEL" env-default:"DEBUG"`
}