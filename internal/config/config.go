package serverConfig

import "github.com/caarlos0/env"

type Config struct {
	FilePath     string `env:"FILE_STORAGE_PATH"`
	ServerAddres string `env:"SERVER_ADDRESS"`
	BaseUrl      string `env:"BASE_URL"`
}

func GetConfig() (*Config, error) {
	var cfg *Config
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
