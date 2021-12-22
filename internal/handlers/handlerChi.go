package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type HandlerChi struct {
	Ids []string
	*chi.Mux
}

func NewHandlerChi() *HandlerChi {
	h := &HandlerChi{
		Ids: make([]string, 0, 3),
		Mux: chi.NewMux(),
	}
	h.Post("/", h.InitialPostHandler())
	h.Get("/{uId}", h.InitialGetHandler())
	return h
}

func (h *HandlerChi) InitialPostHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		s := string(b)
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(201)
		h.Ids = append(h.Ids, s)
		url := "http://" + r.Host + r.URL.String() + strconv.Itoa(len(h.Ids)-1)
		w.Write([]byte(url))
	}
}
func (h *HandlerChi) InitialGetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := chi.URLParam(r, "uId")
		if s == "" {
			http.Error(w, "id is empty", http.StatusBadRequest)
			return
		}
		i, err := strconv.Atoi(s)
		if err != nil || i > len(h.Ids)-1 {
			http.Error(w, "can't find id", http.StatusBadRequest)
			return
		}
		res := h.Ids[i]
		w.Header().Set("Location", res)
		w.WriteHeader(307)
	}
}
