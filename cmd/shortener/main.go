package main

import (
	"log"
	"net/http"

	serverConfig "github.com/GazpachoGit/yandexGoCourse/internal/config"
	"github.com/GazpachoGit/yandexGoCourse/internal/handlers"
	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
)

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
	r := handlers.NewShortenerHandler(urlMap, cfg.BaseUrl)
	server := &http.Server{
		Addr:    cfg.ServerAddres,
		Handler: r,
	}
	server.ListenAndServe()

	//http.ListenAndServe(":8080", r)
}
