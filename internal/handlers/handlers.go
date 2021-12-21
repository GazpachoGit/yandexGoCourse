package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	Ids []string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		s := string(b)
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(201)
		h.Ids = append(h.Ids, s)
		url := r.Host + r.URL.String() + strconv.Itoa(len(h.Ids)-1)
		w.Write([]byte(url))
	case "GET":
		s := r.RequestURI
		if s == "/" {
			http.Error(w, "id is empty", http.StatusBadRequest)
			return
		}
		id := strings.TrimPrefix(s, "/")
		i, err := strconv.Atoi(id)
		if err != nil || i > len(h.Ids)-1 {
			http.Error(w, "con't find id", http.StatusBadRequest)
			return
		}
		res := h.Ids[i]
		w.Header().Set("Location", res)
		w.WriteHeader(307)
	default:
		http.Error(w, "wrong request", http.StatusBadRequest)
		return
	}
}
