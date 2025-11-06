package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Telegram TelegramConfig
	Database DatabaseConfig
	App      AppConfig
	Kinopoisk KinopoiskConfig
	Redis    RedisConfig
}

func LoadConfig() (*Config, error) {
	var cfg Config

	nodeEnv := os.Getenv("NODE_ENV")

	if nodeEnv == "production" {
		err := cleanenv.ReadEnv(&cfg)
		if err != nil {	
			return nil, err
		}
		return &cfg, nil
	}
	
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {	
		return nil, err
	}
	return &cfg, nil
}
