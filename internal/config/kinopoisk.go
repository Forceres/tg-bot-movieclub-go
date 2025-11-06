package config

type KinopoiskConfig struct {
	APIKey string `env:"KINOPOISK_API_KEY" env-required:"true"`
	APIURL string `env:"KINOPOISK_API_URL" env-default:"https://api.kinopoisk.dev/%s/movie/"`
	APIVersion string `env:"KINOPOISK_API_VERSION" env-default:"1.4"`
}