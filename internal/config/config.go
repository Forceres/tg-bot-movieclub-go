package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DomainAddress string `yaml:"domain" env:"DOMAIN_ADDRESS" env-default:"http://localhost:2000"`
	Telegram      TelegramConfig
	Database      DatabaseConfig
	App           AppConfig
	Kinopoisk     KinopoiskConfig
	Redis         RedisConfig
}

func LoadConfig() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		log.Printf("No config file found, reading configuration from ENV variables: %v", err)
	}

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
