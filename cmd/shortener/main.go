package main

import (
	"net/http"

	"github.com/GazpachoGit/yandexGoCourse/internal/handlers"
	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
)

func main() {
	var urlMap storage.UrlGetSet = &storage.UrlMap{}
	r := handlers.NewShortenerHandler(urlMap)
	http.ListenAndServe(":8080", r)
}
