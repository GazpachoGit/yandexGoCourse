package serverconfig

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	FilePath     string `env:"FILE_STORAGE_PATH"`
	ServerAddres string `env:"SERVER_ADDRESS"`
	BaseURL      string `env:"BASE_URL"`
}

func GetConfig() (*Config, error) {

	cfg := &Config{}

	//flags
	FilePath := flag.String("f", "", "path to storage file")
	ServerAddres := flag.String("a", "", "address server to start")
	BaseURL := flag.String("b", "", "url storage address")
	flag.Parse()

	//env
	envCfg := &Config{}
	err := env.Parse(envCfg)
	if err != nil {
		return nil, err
	}
	if envCfg.FilePath != "" {
		cfg.FilePath = envCfg.FilePath
	} else {
		cfg.FilePath = *FilePath
	}
	if envCfg.ServerAddres != "" {
		cfg.ServerAddres = envCfg.ServerAddres
	} else {
		cfg.ServerAddres = *ServerAddres
	}
	if envCfg.BaseURL != "" {
		cfg.BaseURL = envCfg.BaseURL
	} else {
		cfg.BaseURL = *BaseURL
	}

	log.Println("env filePath: " + cfg.FilePath)
	log.Println("env ServerAddres: " + cfg.ServerAddres)
	log.Println("env BaseURL: " + cfg.BaseURL)

	//default
	if cfg.FilePath == "" {
		cfg.FilePath = "storage.txt"
	}
	if cfg.ServerAddres == "" {
		cfg.ServerAddres = ":8080"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080/"
	}

	return cfg, nil
}
