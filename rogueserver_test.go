package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"
	"testing"

	"github.com/pagefaultgames/rogueserver/api/account"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pagefaultgames/rogueserver/api"
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

const (
	dbImage    = "mariadb:11"
	dbName     = "pokeroguedb"
	dbUsername = "pokerogue"
	dbPassword = "pokerogue"
)

func TestRogueServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	if runtime.GOOS == "windows" {
		t.Skip("testcontainers-go does not support windows")
	}

	setupDB(t)
	apiURL := runServer(t)

	username := "user1234"
	oldPassword := "password1234"
	newPassword := "password12345"
	t.Run("/account", func(t *testing.T) {
		t.Run("/account/register", func(t *testing.T) {
			t.Run("invalid username", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/register", nil, url.Values{"username": {""}, "password": {"password"}})
				resp.assertStringResponse(t, http.StatusBadRequest, "invalid username\n")
			})

			t.Run("invalid password", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/register", nil, url.Values{"username": {"username"}, "password": {"1"}})
				resp.assertStringResponse(t, http.StatusBadRequest, "invalid password\n")
			})
			t.Run("register", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/register", nil, url.Values{"username": {username}, "password": {oldPassword}})
				resp.assertStringResponse(t, http.StatusCreated, "")
			})
			t.Run("fail when register called for already registered user", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/register", nil, url.Values{"username": {username}, "password": {oldPassword}})
				resp.assertStringResponse(t, http.StatusConflict, "username \"user1234\" already taken\n")
			})
		})

		var token string
		t.Run("/account/login", func(t *testing.T) {
			t.Run("invalid username", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/login", nil, url.Values{"username": {""}, "password": {"password"}})
				resp.assertStringResponse(t, http.StatusBadRequest, "invalid username\n")
			})
			t.Run("invalid password", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/login", nil, url.Values{"username": {"username"}, "password": {"1"}})
				resp.assertStringResponse(t, http.StatusBadRequest, "invalid password\n")
			})

			t.Run("user do not exist", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/login", nil, url.Values{"username": {"notexist"}, "password": {"123456"}})
				resp.assertStringResponse(t, http.StatusNotFound, "account doesn't exist\n")
			})

			t.Run("bad password", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/login", nil, url.Values{"username": {username}, "password": {"badpassword"}})
				resp.assertStringResponse(t, http.StatusUnauthorized, "password doesn't match\n")
			})

			t.Run("login successfully", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/login", nil, url.Values{"username": {username}, "password": {oldPassword}})
				resp.assertStatusCode(t, http.StatusOK)
				var accountResp account.LoginResponse
				resp.unmarshalJSON(t, &accountResp)
				assert.NotEmpty(t, accountResp.Token)
				token = accountResp.Token
			})
		})

		t.Run("/account/info", func(t *testing.T) {
			t.Run("missing token", func(t *testing.T) {
				resp := getHTTP(t, apiURL, "/account/info", nil)
				resp.assertStringResponse(t, http.StatusBadRequest, "missing token\n")
			})

			t.Run("bad token", func(t *testing.T) {
				resp := getHTTP(t, apiURL, "/account/info", authHeader("foo"))
				resp.assertStringResponse(t, http.StatusBadRequest, "failed to decode token\n")
			})

			t.Run("success", func(t *testing.T) {
				resp := getHTTP(t, apiURL, "/account/info", authHeader(token))
				resp.assertJSONResponse(t, http.StatusOK, `{"username":"user1234","lastSessionSlot":-1}`)
			})
		})

		t.Run("/account/logout", func(t *testing.T) {
			t.Run("missing token", func(t *testing.T) {
				resp := getHTTP(t, apiURL, "/account/logout", nil)
				resp.assertStringResponse(t, http.StatusBadRequest, "missing token\n")
			})

			t.Run("bad token", func(t *testing.T) {
				resp := getHTTP(t, apiURL, "/account/logout", authHeader("foo"))
				resp.assertStringResponse(t, http.StatusBadRequest, "failed to decode token\n")
			})
			t.Run("success", func(t *testing.T) {
				resp := getHTTP(t, apiURL, "/account/logout", authHeader(token))
				resp.assertStatusCode(t, http.StatusNoContent)
			})

			t.Run("do nothing on second logout", func(t *testing.T) {
				resp := getHTTP(t, apiURL, "/account/logout", authHeader(token))
				resp.assertStatusCode(t, http.StatusNoContent)
			})
		})

		t.Run("/account/changepw", func(t *testing.T) {
			t.Run("missing token", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/changepw", nil, url.Values{"password": {newPassword}})
				resp.assertStringResponse(t, http.StatusBadRequest, "missing token\n")
			})

			t.Run("bad token", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/changepw", authHeader("foo"), url.Values{"password": {newPassword}})
				resp.assertStringResponse(t, http.StatusBadRequest, "failed to decode token\n")
			})

			t.Run("fail on unlogged token", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/changepw", authHeader(token), url.Values{"password": {newPassword}})
				resp.assertStringResponse(t, http.StatusUnauthorized, "bad token\n")
			})

			t.Run("login successfully once again", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/login", nil, url.Values{"username": {username}, "password": {oldPassword}})
				resp.assertStatusCode(t, http.StatusOK)
				var accountResp account.LoginResponse
				resp.unmarshalJSON(t, &accountResp)
				assert.NotEmpty(t, accountResp.Token)
				token = accountResp.Token
			})

			t.Run("bad password", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/changepw", authHeader(token), url.Values{"password": {"123"}})
				resp.assertStringResponse(t, http.StatusBadRequest, "invalid password\n")
			})

			t.Run("success", func(t *testing.T) {
				resp := postHttpForm(t, apiURL, "/account/changepw", authHeader(token), url.Values{"password": {newPassword}})
				resp.assertStatusCode(t, http.StatusNoContent)
			})
		})
	})
}

func authHeader(token string) map[string]string {
	return map[string]string{"Authorization": token}
}

type response struct {
	code    int
	headers http.Header
	body    string
}

func (r response) assertStatusCode(t *testing.T, expectedStatus int) {
	t.Helper()
	assert.Equal(t, expectedStatus, r.code)
}

func (r response) assertStringResponse(t *testing.T, expectedStatus int, expectedBody string) {
	t.Helper()
	assert.Equal(t, expectedStatus, r.code)
	assert.Equal(t, expectedBody, r.body)
}

func (r response) assertJSONResponse(t *testing.T, expectedStatus int, expectedBody string) {
	t.Helper()
	assert.Equal(t, expectedStatus, r.code)
	assert.JSONEq(t, expectedBody, r.body)
}

func (r response) unmarshalJSON(t *testing.T, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal([]byte(r.body), &v))
}

func postHttpForm(t *testing.T, apiURL string, p string, headers map[string]string, form url.Values) response {
	t.Helper()
	u := joinPathToAPIURL(t, apiURL, p)
	req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(form.Encode()))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	return readResponse(t, resp)
}

func getHTTP(t *testing.T, apiURL string, p string, headers map[string]string) response {
	t.Helper()
	u := joinPathToAPIURL(t, apiURL, p)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	require.NoError(t, err)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	return readResponse(t, resp)
}

func joinPathToAPIURL(t *testing.T, apiURL string, p string) string {
	u, err := url.JoinPath(apiURL, p)
	require.NoError(t, err)
	return u
}

func readResponse(t *testing.T, resp *http.Response) response {
	t.Helper()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return response{
		code:    resp.StatusCode,
		headers: resp.Header,
		body:    string(respBody),
	}
}

func runServer(t *testing.T) string {
	mux := http.NewServeMux()
	require.NoError(t, api.Init(mux))

	handler := prodHandler(mux)
	s := httptest.NewServer(handler)
	t.Cleanup(s.Close)
	return s.URL
}

func setupDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	c, err := mariadb.RunContainer(ctx,
		testcontainers.WithImage(dbImage),
		mariadb.WithDatabase(dbName),
		mariadb.WithUsername(dbUsername),
		mariadb.WithPassword(dbPassword),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, c.Terminate(ctx))
	})
	port, err := c.MappedPort(ctx, "3306/tcp")
	require.NoError(t, err)

	host, err := c.Host(ctx)
	require.NoError(t, err)

	err = db.Init(dbUsername, dbPassword, "tcp", net.JoinHostPort(host, port.Port()), dbName)
	require.NoError(t, err)

	t.Logf("connection string to db for debugging (valid until tests are running): %s", c.MustConnectionString(ctx))
}
