package serverConfig

import (
	"flag"

	"github.com/caarlos0/env"
)

type Config struct {
	FilePath     string `env:"FILE_STORAGE_PATH"`
	ServerAddres string `env:"SERVER_ADDRESS"`
	BaseUrl      string `env:"BASE_URL"`
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
	BaseUrl := flag.String("BASE_URL", "", "url storage address")
	if *BaseUrl != "" {
		cfg.FilePath = *BaseUrl
	}

	return cfg, nil
}
