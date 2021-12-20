package main

import (
	"net/http"

	"github.com/GazpachoGit/yandexGoCourse/internal/handlers"
)

func main() {
	handler := &handlers.Handler{Ids: make([]string, 0, 3)}
	http.Handle("/", handler)
	http.ListenAndServe(":8080", nil)
}
