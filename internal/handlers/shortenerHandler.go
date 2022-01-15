package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
	"github.com/go-chi/chi"
)

type ShortenerHandler struct {
	*chi.Mux
	urlMap storage.GetSet
}

type ShortenerRequestBoby struct {
	URL string `json:"url,omitempty"`
}

type ShortenerResponseBoby struct {
	Result string `json:"result"`
}

func NewShortenerHandler(urlMapInput storage.GetSet) *ShortenerHandler {
	h := &ShortenerHandler{
		Mux:    chi.NewMux(),
		urlMap: urlMapInput,
	}
	h.Post("/", h.NewShortURL())
	h.Get("/{id}", h.GetShortURL())
	h.Post("/api/shorten", h.NewShortURLByJson())
	return h
}
func (h *ShortenerHandler) NewShortURLByJson() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		requestBody := &ShortenerRequestBoby{}
		json.Unmarshal(b, requestBody)
		if requestBody.URL == "" {
			http.Error(w, "url is empty", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		id := h.urlMap.Set(requestBody.URL)

		url := h.formUrl(r, id)
		responseBody := &ShortenerResponseBoby{Result: url}
		requestBodyJson, err := json.Marshal(responseBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(requestBodyJson))
	}
}

func (h *ShortenerHandler) NewShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s := string(b)
		if s == "" {
			http.Error(w, "url is empty", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		id := h.urlMap.Set(s)

		url := h.formUrl(r, id)
		w.Write([]byte(url))
	}
}
func (h *ShortenerHandler) formUrl(r *http.Request, id int) string {
	url := url.URL{
		Scheme: "http",
		Host:   r.Host,
		Path:   "/" + strconv.Itoa(id),
	}
	return url.String()
}
func (h *ShortenerHandler) GetShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := chi.URLParam(r, "id")
		if s == "" {
			http.Error(w, "id is empty", http.StatusBadRequest)
			return
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, "incorrect id", http.StatusBadRequest)
			return
		}
		if res, err := h.urlMap.Get(i); err != nil {
			if err.Error() == storage.ErrNotFound {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		} else {
			w.Header().Set("Location", res)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
}
