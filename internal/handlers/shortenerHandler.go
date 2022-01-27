package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/GazpachoGit/yandexGoCourse/internal/middlewares"
	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
	"github.com/go-chi/chi"
)

type ShortenerHandler struct {
	*chi.Mux
	URLMap  storage.GetSet
	BaseURL string
}

type ShortenerRequestBoby struct {
	URL string `json:"url,omitempty"`
}

type ShortenerResponseBoby struct {
	Result string `json:"result"`
}

func NewShortenerHandler(urlMapInput storage.GetSet, BaseURL string) *ShortenerHandler {
	h := &ShortenerHandler{
		Mux:     chi.NewMux(),
		URLMap:  urlMapInput,
		BaseURL: BaseURL,
	}
	compressor := &middlewares.Compressor{}
	h.Use(compressor.CompressHandler)
	h.Use(middlewares.DecompressHandler)
	h.Use(middlewares.CockieHandler)
	h.Post("/", h.NewShortURL())
	h.Get("/{id}", h.GetShortURL())
	h.Post("/api/shorten", h.NewShortURLByJSON())
	return h
}
func (h *ShortenerHandler) NewShortURLByJSON() http.HandlerFunc {
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
		id, err := h.URLMap.Set(requestBody.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		url, err := h.formURL(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		responseBody := &ShortenerResponseBoby{Result: url}
		requestBodyJSON, err := json.Marshal(responseBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(requestBodyJSON))
	}
}

func (h *ShortenerHandler) NewShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("user")
		fmt.Println(username)
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
		id, err := h.URLMap.Set(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		url, err := h.formURL(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(url))
	}
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
		if res, err := h.URLMap.Get(i); err != nil {
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
func (h *ShortenerHandler) formURL(id int) (string, error) {
	u, err := url.ParseRequestURI(h.BaseURL)
	if err != nil {
		return "", err
	}
	output := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "/" + strconv.Itoa(id),
	}
	return output.String(), nil
}
