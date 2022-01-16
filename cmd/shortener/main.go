package main

import (
	"log"
	"net/http"

	"github.com/GazpachoGit/yandexGoCourse/internal/handlers"
	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	FilePath string `env:"FILE_STORAGE_PATH"`
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
		return
	}

	urlMap, err := storage.NewUrlMap(cfg.FilePath)
	defer urlMap.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	r := handlers.NewShortenerHandler(urlMap)
	http.ListenAndServe(":8080", r)
}
