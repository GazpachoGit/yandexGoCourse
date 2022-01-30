package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/GazpachoGit/yandexGoCourse/internal/middlewares"
	"github.com/GazpachoGit/yandexGoCourse/internal/model"
	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
	"github.com/go-chi/chi"
)

type ShortenerHandler struct {
	*chi.Mux
	db            storage.IStorage
	BaseURLstring string
	BaseURL       *url.URL
}

type ShortenerRequestBoby struct {
	URL string `json:"url,omitempty"`
}

type ShortenerResponseBoby struct {
	Result string `json:"result"`
}

func NewShortenerHandler(urlMapInput storage.IStorage, BaseURL string) (*ShortenerHandler, error) {
	h := &ShortenerHandler{
		Mux:           chi.NewMux(),
		db:            urlMapInput,
		BaseURLstring: BaseURL,
	}
	if err := h.initBaseURL(); err != nil {
		return nil, err
	}
	compressor := &middlewares.Compressor{}
	h.Use(compressor.CompressHandler)
	h.Use(middlewares.DecompressHandler)
	h.Use(middlewares.CockieHandler)
	h.Post("/", h.NewShortURL())
	h.Get("/{id}", h.GetShortURL())
	h.Post("/api/shorten", h.NewShortURLByJSON())
	h.Get("/user/urls", h.GetUserURLs())
	h.Get("/ping", h.CheckDBConnection())
	h.Post("/api/shorten/batch", h.SetBatchURLs())
	return h, nil
}

func (h *ShortenerHandler) initBaseURL() error {
	u, err := url.ParseRequestURI(h.BaseURLstring)
	if err != nil {
		return err
	}
	h.BaseURL = u
	return nil
}

func (h *ShortenerHandler) GetUserURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("user").(string)
		if res, err := h.db.GetUserURLs(username); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			if res == nil {
				http.Error(w, "no urls for this user", http.StatusNoContent)
				return
			}
			URLList := make([]model.HandlerURLInfo, 0)
			for _, url := range res {
				URLList = append(URLList, model.HandlerURLInfo{
					Original_url: url.Original_url,
					Short_url:    h.formURL(url.Id),
				})
			}
			respBody, err := json.Marshal(URLList)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(respBody))
		}

	}
}

func (h *ShortenerHandler) SetBatchURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("user").(string)
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		requestBody := make([]*model.HandlerURLInfo, 0)
		json.Unmarshal(b, &requestBody)
		dbUrls, err := h.db.SetBatchURLs(&requestBody, username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		URLList := make([]model.HandlerURLInfo, 0)
		for k, v := range *dbUrls {
			URLList = append(URLList, model.HandlerURLInfo{
				Correlation_id: k,
				Short_url:      h.formURL(v.Id),
			})
		}
		respBody, err := json.Marshal(URLList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(respBody))
	}
}

func (h *ShortenerHandler) NewShortURLByJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("user").(string)
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
		id, err := h.db.Set(requestBody.URL, username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		url := h.formURL(id)
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
		username := r.Context().Value("user").(string)
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
		id, err := h.db.Set(s, username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		url := h.formURL(id)
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
		if res, err := h.db.Get(i); err != nil {
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
func (h *ShortenerHandler) CheckDBConnection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.db.PingDB(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
func (h *ShortenerHandler) formURL(id int) string {
	output := url.URL{
		Scheme: h.BaseURL.Scheme,
		Host:   h.BaseURL.Host,
		Path:   "/" + strconv.Itoa(id),
	}
	return output.String()
}
