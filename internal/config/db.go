package config

type DatabaseConfig struct {
	Name string `env:"DATABASE_NAME" env-default:"movieclub"`
	Host string `env:"DATABASE_HOST" env-default:"localhost"`
	Port string `env:"DATABASE_PORT" env-default:"5432"`
	User string `env:"DATABASE_USER" env-default:"movieclub_user"`
	Pass string `env:"DATABASE_PASS" env-default:"movieclub_pass"`
}
