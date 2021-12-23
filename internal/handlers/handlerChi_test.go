//go test github.com/GazpachoGit/yandexGoCourse/internal/handlers -run TestRequest -count 1
package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type respData struct {
	body       string
	StatusCode int
	Header     http.Header
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
	transport := http.Transport{}
	resp, err := transport.RoundTrip(req)

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
	r := NewHandlerChi()
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp := testRequest(t, ts, http.MethodPost, "/", []byte("http://ya.ru"))
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, ts.URL+"/0", resp.body)

	resp = testRequest(t, ts, http.MethodPost, "/", []byte("http://google.ru"))
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, ts.URL+"/1", resp.body)

	resp = testRequest(t, ts, http.MethodGet, "/0", nil)
	assert.Equal(t, 307, resp.StatusCode)
	assert.Equal(t, "http://ya.ru", resp.Header.Get("Location"))

	resp = testRequest(t, ts, http.MethodGet, "/1", nil)
	assert.Equal(t, 307, resp.StatusCode)
	assert.Equal(t, "http://google.ru", resp.Header.Get("Location"))

}
