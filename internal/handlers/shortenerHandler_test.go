//go test github.com/GazpachoGit/yandexGoCourse/internal/handlers -run TestRouter -count 1
package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	serverConfig "github.com/GazpachoGit/yandexGoCourse/internal/config"
	myerrors "github.com/GazpachoGit/yandexGoCourse/internal/errors"
	"github.com/GazpachoGit/yandexGoCourse/internal/storage"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sqlxmock "github.com/zhashkevych/go-sqlxmock"
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
	mockdb, mock, err := sqlxmock.New()

	defer mockdb.Close()
	db := sqlx.NewDb(mockdb, "sqlmock")

	rowsInsert1 := sqlxmock.NewRows([]string{"id,", "conf"}).AddRow(1, false)
	rowsInsert2 := sqlxmock.NewRows([]string{"id,", "conf"}).AddRow(2, false)

	insertSQL := regexp.QuoteMeta(`with stmt AS (INSERT INTO public.urls_torn(original_url, user_id)
	VALUES ($1, $2) 
	ON CONFLICT(original_url) do nothing
	RETURNING id, false as conf)

	select id, conf from stmt 
	where id is not null
	UNION ALL
	select id, true from public.urls_torn
	where original_url = $1 and not exists (select 1 from stmt)`)

	mock.ExpectPrepare(insertSQL).ExpectQuery().WithArgs("http://ya.ru", "1").WillReturnRows(rowsInsert1)
	mock.ExpectPrepare(insertSQL).ExpectQuery().WithArgs("https://google.com", "1").WillReturnRows(rowsInsert2)

	rowsSelect1 := sqlxmock.NewRows([]string{"original_url"}).AddRow("http://ya.ru")
	rowsSelect2 := sqlxmock.NewRows([]string{"original_url"}).AddRow("https://google.com")

	selectOneSQL := regexp.QuoteMeta("SELECT original_url FROM public.urls_torn WHERE id = $1 LIMIT 1")
	mock.ExpectPrepare(selectOneSQL).ExpectQuery().WithArgs(1).WillReturnRows(rowsSelect1)
	mock.ExpectPrepare(selectOneSQL).ExpectQuery().WithArgs(2).WillReturnRows(rowsSelect2)

	rowsUserSelect1 := sqlxmock.NewRows([]string{"id", "original_url"}).AddRow(1, "http://ya.ru").AddRow(2, "https://google.com")
	selectUserURLsSQL := regexp.QuoteMeta("SELECT id, original_url FROM public.urls_torn WHERE user_id = $1")
	mock.ExpectPrepare(selectUserURLsSQL).ExpectQuery().WithArgs(1).WillReturnRows(rowsUserSelect1)

	pDb, err := storage.ConfigDb(db)
	require.NoError(t, err)
	defer pDb.Close()

	r, _ := NewShortenerHandler(pDb, cfg.BaseURL)
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
				response: cfg.BaseURL + "1",
			},
		},
		{
			name:   "post 2",
			method: http.MethodPost,
			path:   "/",
			body:   []byte("https://google.com"),
			want: &want{
				code:     http.StatusCreated,
				response: cfg.BaseURL + "2",
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
				response: myerrors.NewNotFoundError().Error() + "\n",
			},
		},
		{
			name:   "post json",
			method: http.MethodPost,
			path:   "/api/shorten",
			body:   []byte("{\"url\":\"http://yandex.ru\"}"),
			want: &want{
				code:     http.StatusCreated,
				response: "{\"result\":\"" + cfg.BaseURL + "3" + "\"}",
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
