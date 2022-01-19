package serverconfig

import (
	"flag"

	"github.com/caarlos0/env"
)

type Config struct {
	FilePath     string `env:"FILE_STORAGE_PATH"`
	ServerAddres string `env:"SERVER_ADDRESS"`
	BaseURL      string `env:"BASE_URL"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	FilePath := flag.String("FILE_STORAGE_PATH", "", "path to storage file")
	if *FilePath != "" {
		cfg.FilePath = *FilePath
	}
	ServerAddres := flag.String("SERVER_ADDRESS", "", "address server to start")
	if *ServerAddres != "" {
		cfg.FilePath = *ServerAddres
	}
	BaseURL := flag.String("BASE_URL", "", "url storage address")
	if *BaseURL != "" {
		cfg.BaseURL = *BaseURL
	}

	return cfg, nil
}
