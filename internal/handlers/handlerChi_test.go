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

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body []byte) (*http.Response, string) {

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

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	r := NewHandlerChi()
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, body := testRequest(t, ts, http.MethodPost, "/", []byte("http://ya.ru")) //nolint:bodyclose // linters bug
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, ts.URL+"/0", body)

	resp, body = testRequest(t, ts, http.MethodPost, "/", []byte("http://google.ru")) //nolint:bodyclose // linters bug
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, ts.URL+"/1", body)

	resp, _ = testRequest(t, ts, http.MethodGet, "/0", nil) //nolint:bodyclose // linters bug
	assert.Equal(t, 307, resp.StatusCode)
	assert.Equal(t, "http://ya.ru", resp.Header.Get("Location")) //nolint:bodyclose // linters bug

	resp, _ = testRequest(t, ts, http.MethodGet, "/1", nil)
	assert.Equal(t, 307, resp.StatusCode)
	assert.Equal(t, "http://google.ru", resp.Header.Get("Location"))

}
