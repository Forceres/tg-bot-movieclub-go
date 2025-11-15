package config

type RedisConfig struct {
	URL string `env:"REDIS_URL" required:"true"`
}
