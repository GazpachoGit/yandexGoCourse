package main

import (
	"log"
	"net/http"

	serverConfig "github.com/GazpachoGit/yandexGoCourse/internal/config"
	"github.com/GazpachoGit/yandexGoCourse/internal/handlers"
	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
)

type Config struct {
	FilePath string `env:"FILE_STORAGE_PATH"`
}

func main() {
	cfg, err := serverConfig.GetConfig()
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
	server := &http.Server{
		Addr:    cfg.ServerAddres,
		Handler: r,
	}
	server.ListenAndServe()

	//http.ListenAndServe(":8080", r)
}
