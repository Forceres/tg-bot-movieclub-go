package config

type RedisConfig struct {
	URL      string `env:"REDIS_URL" envDefault:"127.0.0.1:6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	Username string `env:"REDIS_USERNAME" envDefault:""`
}
