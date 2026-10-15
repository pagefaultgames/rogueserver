package account

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

type mockGoogleTransport struct {
	TokenFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockGoogleTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/token" {
		return m.TokenFunc(req)
	}
	return nil, fmt.Errorf("unexpected request to %s", req.URL)
}

// googleTokenBody builds a token response whose id_token carries the given claims.
// The signature is never checked, so any key works.
func googleTokenBody(claims jwt.MapClaims) string {
	idToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-key"))
	return fmt.Sprintf(`{"access_token":"access","id_token":%q}`, idToken)
}

func defaultMockGoogleTransport() *mockGoogleTransport {
	return &mockGoogleTransport{
		TokenFunc: respond(http.StatusOK, googleTokenBody(jwt.MapClaims{"sub": "g-123"})),
	}
}

// newTestGoogleProvider returns a provider backed by the default mock transport.
// A nil tokenFunc keeps the default successful response.
func newTestGoogleProvider(tokenFunc func(*http.Request) (*http.Response, error)) *googleProvider {
	transport := defaultMockGoogleTransport()
	if tokenFunc != nil {
		transport.TokenFunc = tokenFunc
	}

	return &googleProvider{
		client:   &http.Client{Transport: transport},
		tokenURL: "https://google.test/token",
	}
}

func TestRetrieveGoogleId(t *testing.T) {
	networkErr := errors.New("network down")

	t.Run("Success", func(t *testing.T) {
		id, err := newTestGoogleProvider(nil).RetrieveGoogleId("code")
		expectId(t, id, err, "g-123")
	})
	// Verifies that RetrieveGoogleId sends the authorization code
	// and grant type in the token exchange.
	t.Run("SendsRequests", func(t *testing.T) {
		var tokenReq *http.Request
		provider := newTestGoogleProvider(func(req *http.Request) (*http.Response, error) {
			tokenReq = req
			return jsonResponse(http.StatusOK, googleTokenBody(jwt.MapClaims{"sub": "g-123"})), nil
		})

		_, err := provider.RetrieveGoogleId("the-code")
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
	})
	t.Run("TokenRequestError", func(t *testing.T) {
		_, err := newTestGoogleProvider(fail(networkErr)).RetrieveGoogleId("code")
		expectErr(t, err, networkErr)
	})
	t.Run("TokenExchangeFails", func(t *testing.T) {
		_, err := newTestGoogleProvider(respond(http.StatusBadRequest, `{"error":"invalid_grant"}`)).RetrieveGoogleId("code")
		expectErr(t, err, nil)
	})
	t.Run("InvalidTokenResponse", func(t *testing.T) {
		_, err := newTestGoogleProvider(respond(http.StatusOK, `not json`)).RetrieveGoogleId("code")
		expectErr(t, err, nil)
	})
	t.Run("MissingIdToken", func(t *testing.T) {
		_, err := newTestGoogleProvider(respond(http.StatusOK, `{"access_token":"access"}`)).RetrieveGoogleId("code")
		expectErr(t, err, nil)
	})
	t.Run("InvalidIdToken", func(t *testing.T) {
		_, err := newTestGoogleProvider(respond(http.StatusOK, `{"id_token":"not-a-jwt"}`)).RetrieveGoogleId("code")
		expectErr(t, err, nil)
	})
}

func TestHandleGoogleCallback(t *testing.T) {
	t.Run("Cancelled", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?error=access_denied", nil)
		_, err := newTestGoogleProvider(nil).HandleGoogleCallback(req)
		expectErr(t, err, ErrOAuthCancelled)
	})
	t.Run("MissingCode", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback", nil)
		_, err := newTestGoogleProvider(nil).HandleGoogleCallback(req)
		expectErr(t, err, nil)
		if errors.Is(err, ErrOAuthCancelled) {
			t.Errorf("expected missing code error, got ErrOAuthCancelled")
		}
	})
	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=abc", nil)
		id, err := newTestGoogleProvider(nil).HandleGoogleCallback(req)
		expectId(t, id, err, "g-123")
	})
}
