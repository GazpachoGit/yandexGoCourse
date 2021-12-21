package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitalHandler(t *testing.T) {
	type want struct {
		code           int
		response       string
		locationHeader string
	}
	type test struct {
		name   string
		want   want
		body   []byte
		method string
		id     string
	}
	postTests := []test{
		{
			name:   "test POST 1",
			body:   []byte(`https://ya.ru`),
			method: http.MethodPost,
			want: want{
				code:     201,
				response: `example.com/0`,
			},
		},
		{
			name:   "test POST 2",
			body:   []byte(`https://google.com`),
			method: http.MethodPost,
			want: want{
				code:     201,
				response: `example.com/1`,
			},
		},
	}
	getTests := []test{
		{
			name:   "test GET 1",
			method: http.MethodGet,
			id:     "0",
			want: want{
				code:           307,
				locationHeader: `https://ya.ru`,
			},
		},
		{
			name:   "test GET 2",
			method: http.MethodGet,
			id:     "1",
			want: want{
				code:           307,
				locationHeader: `https://google.com`,
			},
		},
		{
			name:   "test GET 3",
			method: http.MethodGet,
			id:     "123",
			want: want{
				code:           400,
				locationHeader: "",
			},
		},
		{
			name:   "test GET 3",
			method: http.MethodGet,
			id:     "",
			want: want{
				code:           400,
				locationHeader: "",
			},
		},
	}
	handler := Handler{Ids: make([]string, 0, 3)}
	for _, tt := range postTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, `/`, bytes.NewBuffer(tt.body))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler.ServeHTTP)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tt.want.response {
				t.Errorf("Expected body %s, got %s", tt.want.response, w.Body.String())
			}
		})
	}
	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, `/`+tt.id, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler.ServeHTTP)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			if res.Header.Get("Location") != tt.want.locationHeader {
				t.Errorf("Expected Location header %v, got %v", tt.want.locationHeader, res.Header.Get("Location"))
			}
		})
	}
}
