package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
	"github.com/go-chi/chi"
)

type ShortenerHandler struct {
	*chi.Mux
}

func NewShortenerHandler(urlMap storage.GetSet) *ShortenerHandler {
	h := &ShortenerHandler{
		Mux: chi.NewMux(),
	}
	h.Post("/", h.NewShortURL(urlMap))
	h.Get("/{id}", h.GetShortURL(urlMap))
	return h
}

func (h *ShortenerHandler) NewShortURL(urlMap storage.GetSet) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		s := string(b)
		if s == "" {
			http.Error(w, "url is empty", http.StatusBadRequest)
		}
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		id := urlMap.Set(s)
		url := "http://" + r.Host + r.URL.String() + strconv.Itoa(id)
		w.Write([]byte(url))
	}
}
func (h *ShortenerHandler) GetShortURL(urlMap storage.GetSet) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := chi.URLParam(r, "id")
		if s == "" {
			http.Error(w, "id is empty", http.StatusBadRequest)
			return
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, "can't find id", http.StatusNotFound)
			return
		}
		if res, err := urlMap.Get(i); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)

		} else {
			w.Header().Set("Location", res)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}
