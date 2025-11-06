package config

type DatabaseConfig struct {
	Name string `env:"DATABASE_NAME" env-default:"movieclub.db"`
}