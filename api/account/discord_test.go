package account

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// respond returns a round-trip func that always answers with status and body.
func respond(status int, body string) func(*http.Request) (*http.Response, error) {
	return func(*http.Request) (*http.Response, error) {
		return jsonResponse(status, body), nil
	}
}

// fail returns a round-trip func that always fails with err, like a network error.
func fail(err error) func(*http.Request) (*http.Response, error) {
	return func(*http.Request) (*http.Response, error) {
		return nil, err
	}
}

// expectId fails the test unless err is nil and gotId matches wantId.
func expectId(t *testing.T, gotId string, err error, wantId string) {
	t.Helper()
	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}
	if gotId != wantId {
		t.Errorf("expected id %q, got: %q", wantId, gotId)
	}
}

// expectErr fails the test unless err is non-nil, and matches wantErrIs when it's set.
func expectErr(t *testing.T, err error, wantErrIs error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if wantErrIs != nil && !errors.Is(err, wantErrIs) {
		t.Errorf("expected %v, got: %v", wantErrIs, err)
	}
}

type mockDiscordTransport struct {
	TokenFunc func(req *http.Request) (*http.Response, error)
	UserFunc  func(req *http.Request) (*http.Response, error)
}

func (m *mockDiscordTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Path {
	case "/token":
		return m.TokenFunc(req)
	case "/users/@me":
		return m.UserFunc(req)
	}
	return nil, fmt.Errorf("unexpected request to %s", req.URL)
}

func defaultMockDiscordTransport() *mockDiscordTransport {
	return &mockDiscordTransport{
		TokenFunc: respond(http.StatusOK, `{"access_token":"access"}`),
		UserFunc:  respond(http.StatusOK, `{"id":"123"}`),
	}
}

func newMockDiscordProvider(transport *mockDiscordTransport) *discordProvider {
	return &discordProvider{
		client:   &http.Client{Transport: transport},
		tokenURL: "https://discord.test/token",
		userURL:  "https://discord.test/users/@me",
	}
}

func TestRetrieveDiscordId(t *testing.T) {
	networkErr := errors.New("network down")

	t.Run("Success", func(t *testing.T) {
		id, err := newMockDiscordProvider(defaultMockDiscordTransport()).RetrieveDiscordId("code")
		expectId(t, id, err, "123")
	})

	// Verifies that RetrieveDiscordId sends the authorization code
	// and grant type in the token exchange, and uses the returned access token
	// as the bearer token when fetching the user.
	t.Run("SendsRequests", func(t *testing.T) {
		var tokenReq, userReq *http.Request
		transport := defaultMockDiscordTransport()
		transport.TokenFunc = func(req *http.Request) (*http.Response, error) {
			tokenReq = req
			return jsonResponse(http.StatusOK, `{"access_token":"access"}`), nil
		}
		transport.UserFunc = func(req *http.Request) (*http.Response, error) {
			userReq = req
			return jsonResponse(http.StatusOK, `{"id":"123"}`), nil
		}

		_, err := newMockDiscordProvider(transport).RetrieveDiscordId("the-code")
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}

		err = tokenReq.ParseForm()
		if err != nil {
			t.Fatalf("failed to parse token request: %v", err)
		}

		if tokenReq.PostForm.Get("code") != "the-code" {
			t.Errorf("expected code the-code, got: %q", tokenReq.PostForm.Get("code"))
		}
		if tokenReq.PostForm.Get("grant_type") != "authorization_code" {
			t.Errorf("expected grant_type authorization_code, got: %q", tokenReq.PostForm.Get("grant_type"))
		}
		if userReq.Header.Get("Authorization") != "Bearer access" {
			t.Errorf("expected Bearer access, got: %q", userReq.Header.Get("Authorization"))
		}
	})
	t.Run("TokenRequestError", func(t *testing.T) {
		transport := defaultMockDiscordTransport()
		transport.TokenFunc = fail(networkErr)
		_, err := newMockDiscordProvider(transport).RetrieveDiscordId("code")
		expectErr(t, err, networkErr)
	})
	t.Run("TokenExchangeFails", func(t *testing.T) {
		transport := defaultMockDiscordTransport()
		transport.TokenFunc = respond(http.StatusBadRequest, `{"error":"invalid_grant"}`)
		_, err := newMockDiscordProvider(transport).RetrieveDiscordId("code")
		expectErr(t, err, nil)
	})
	t.Run("InvalidTokenResponse", func(t *testing.T) {
		transport := defaultMockDiscordTransport()
		transport.TokenFunc = respond(http.StatusOK, `not json`)
		_, err := newMockDiscordProvider(transport).RetrieveDiscordId("code")
		expectErr(t, err, nil)
	})
	t.Run("EmptyAccessToken", func(t *testing.T) {
		transport := defaultMockDiscordTransport()
		transport.TokenFunc = respond(http.StatusOK, `{}`)
		_, err := newMockDiscordProvider(transport).RetrieveDiscordId("code")
		expectErr(t, err, ErrEmptyAccessToken)
	})
	t.Run("UserRequestError", func(t *testing.T) {
		transport := defaultMockDiscordTransport()
		transport.UserFunc = fail(networkErr)
		_, err := newMockDiscordProvider(transport).RetrieveDiscordId("code")
		expectErr(t, err, networkErr)
	})
	t.Run("UserLookupFails", func(t *testing.T) {
		transport := defaultMockDiscordTransport()
		transport.UserFunc = respond(http.StatusUnauthorized, `{"message":"401: Unauthorized","code":0}`)
		_, err := newMockDiscordProvider(transport).RetrieveDiscordId("code")
		expectErr(t, err, nil)
	})
	t.Run("EmptyUserId", func(t *testing.T) {
		transport := defaultMockDiscordTransport()
		transport.UserFunc = respond(http.StatusOK, `{}`)
		_, err := newMockDiscordProvider(transport).RetrieveDiscordId("code")
		expectErr(t, err, ErrEmptyDiscordId)
	})
}

func TestHandleDiscordCallback(t *testing.T) {
	t.Run("Cancelled", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/discord/callback?error=access_denied", nil)
		_, err := newMockDiscordProvider(defaultMockDiscordTransport()).HandleDiscordCallback(req)
		expectErr(t, err, ErrOAuthCancelled)
	})
	t.Run("MissingCode", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/discord/callback", nil)
		_, err := newMockDiscordProvider(defaultMockDiscordTransport()).HandleDiscordCallback(req)
		expectErr(t, err, nil)
		if errors.Is(err, ErrOAuthCancelled) {
			t.Errorf("expected missing code error, got ErrOAuthCancelled")
		}
	})
	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/discord/callback?code=abc", nil)
		id, err := newMockDiscordProvider(defaultMockDiscordTransport()).HandleDiscordCallback(req)
		expectId(t, id, err, "123")
	})
}
