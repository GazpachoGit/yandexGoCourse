//go test github.com/GazpachoGit/yandexGoCourse/internal/handlers -run TestRouter -count 1
package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	serverConfig "github.com/GazpachoGit/yandexGoCourse/internal/config"
	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type respData struct {
	body       string
	StatusCode int
	Header     http.Header
}
type want struct {
	code     int
	response string
	header   string
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body []byte) *respData {

	var req *http.Request
	var err error
	if body == nil {
		req, err = http.NewRequest(method, ts.URL+path, nil)
	} else {
		req, err = http.NewRequest(method, ts.URL+path, bytes.NewBuffer(body))
	}
	require.NoError(t, err)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	resp, err := client.Do(req)

	require.NoError(t, err)

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)

	require.NoError(t, err)

	return &respData{
		body:       string(respBody),
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
	}
}

func TestRouter(t *testing.T) {
	cfg, _ := serverConfig.GetConfig()
	urlMap, _ := storage.NewURLMap(cfg.FilePath)

	r := NewShortenerHandler(urlMap, cfg.BaseURL)
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name   string
		method string
		path   string
		body   []byte
		want   *want
	}{
		{
			name:   "post 1",
			method: http.MethodPost,
			path:   "/",
			body:   []byte("http://ya.ru"),
			want: &want{
				code:     http.StatusCreated,
				response: cfg.BaseURL + strconv.Itoa(urlMap.GetCount()+1),
			},
		},
		{
			name:   "post 2",
			method: http.MethodPost,
			path:   "/",
			body:   []byte("https://google.com"),
			want: &want{
				code:     http.StatusCreated,
				response: cfg.BaseURL + strconv.Itoa(urlMap.GetCount()+2),
			},
		},
		{
			name:   "get 1",
			method: http.MethodGet,
			path:   "/1",
			body:   nil,
			want: &want{
				code:   http.StatusTemporaryRedirect,
				header: "http://ya.ru",
			},
		},
		{
			name:   "get 2",
			method: http.MethodGet,
			path:   "/2",
			body:   nil,
			want: &want{
				code:   http.StatusTemporaryRedirect,
				header: "https://google.com",
			},
		},
		{
			name:   "get with incorrect id",
			method: http.MethodGet,
			path:   "/a",
			body:   nil,
			want: &want{
				code:     http.StatusBadRequest,
				header:   "",
				response: "incorrect id\n",
			},
		},
		{
			name:   "get when id is not found",
			method: http.MethodGet,
			path:   "/123",
			body:   nil,
			want: &want{
				code:     http.StatusNotFound,
				header:   "",
				response: storage.ErrNotFound + "\n",
			},
		},
		{
			name:   "post json",
			method: http.MethodPost,
			path:   "/api/shorten",
			body:   []byte("{\"url\":\"http://yandex.ru\"}"),
			want: &want{
				code:     http.StatusCreated,
				response: "{\"result\":\"" + cfg.BaseURL + strconv.Itoa(urlMap.GetCount()+3) + "\"}",
			},
		},
		{
			name:   "post json without url",
			method: http.MethodPost,
			path:   "/api/shorten",
			body:   []byte("{}"),
			want: &want{
				code:     http.StatusBadRequest,
				response: "url is empty\n",
			},
		},
	}

	for _, tt := range tests {
		resp := testRequest(t, ts, tt.method, tt.path, tt.body)
		assert.Equal(t, tt.want.code, resp.StatusCode, tt.name)
		if tt.method == http.MethodPost {
			assert.Equal(t, tt.want.response, resp.body, tt.name)
		} else {
			assert.Equal(t, tt.want.header, resp.Header.Get("Location"), tt.name)
			if tt.want.response != "" {
				assert.Equal(t, tt.want.response, resp.body, tt.name)
			}
		}
	}
}
