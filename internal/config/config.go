package serverconfig

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	FilePath     string `env:"FILE_STORAGE_PATH"`
	ServerAddres string `env:"SERVER_ADDRESS"`
	BaseURL      string `env:"BASE_URL"`
}

func GetConfig() (*Config, error) {

	// os.Setenv("FILE_STORAGE_PATH", "../../internal/storage/storage.txt")
	// os.Setenv("SERVER_ADDRESS", ":8080")
	// os.Setenv("BASE_URL", "http://localhost:8080/")

	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	fmt.Println("env filePath: " + cfg.FilePath)
	fmt.Println("env ServerAddres: " + cfg.ServerAddres)
	fmt.Println("env BaseURL: " + cfg.BaseURL)

	if cfg.FilePath == "" {
		cfg.FilePath = "storage.txt"
	}
	if cfg.ServerAddres == "" {
		cfg.ServerAddres = ":8080"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080/"
	}

	FilePath := flag.String("f", "", "path to storage file")
	ServerAddres := flag.String("a", "", "address server to start")
	BaseURL := flag.String("b", "", "url storage address")
	flag.Parse()
	if *FilePath != "" {
		cfg.FilePath = *FilePath
	}
	if *ServerAddres != "" {
		cfg.ServerAddres = *ServerAddres
	}
	if *BaseURL != "" {
		cfg.BaseURL = *BaseURL
	}

	return cfg, nil
}
