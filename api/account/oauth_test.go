package account

import (
	"bytes"
	"database/sql"
	"errors"
	"testing"
)

var errUnexpectedCall = errors.New("unexpected store call")

func defaultMockOAuthStore() *mockOAuthStore {
	return &mockOAuthStore{
		FetchByDiscordFunc: func(discordId string) (string, error) { return "", errUnexpectedCall },
		FetchByGoogleFunc:  func(googleId string) (string, error) { return "", errUnexpectedCall },
		FetchBySessionFunc: func(token []byte) (string, error) { return "", errUnexpectedCall },
		AddDiscordFunc:     func(discordId, username string) error { return errUnexpectedCall },
		AddGoogleFunc:      func(googleId, username string) error { return errUnexpectedCall },
		AddSessionFunc:     func(username string, token []byte) error { return nil },
	}
}

type mockOAuthStore struct {
	FetchByDiscordFunc func(discordId string) (string, error)
	FetchByGoogleFunc  func(googleId string) (string, error)
	FetchBySessionFunc func(token []byte) (string, error)
	AddDiscordFunc     func(discordId, username string) error
	AddGoogleFunc      func(googleId, username string) error
	AddSessionFunc     func(username string, token []byte) error
}

func (m *mockOAuthStore) FetchUsernameByDiscordId(discordId string) (string, error) {
	return m.FetchByDiscordFunc(discordId)
}
func (m *mockOAuthStore) FetchUsernameByGoogleId(googleId string) (string, error) {
	return m.FetchByGoogleFunc(googleId)
}
func (m *mockOAuthStore) FetchUsernameBySessionToken(token []byte) (string, error) {
	return m.FetchBySessionFunc(token)
}
func (m *mockOAuthStore) AddDiscordIdByUsername(discordId, username string) error {
	return m.AddDiscordFunc(discordId, username)
}
func (m *mockOAuthStore) AddGoogleIdByUsername(googleId, username string) error {
	return m.AddGoogleFunc(googleId, username)
}
func (m *mockOAuthStore) AddAccountSession(username string, token []byte) error {
	return m.AddSessionFunc(username, token)
}

func TestOAuthLogin(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var sessionFor string
		store := defaultMockOAuthStore()
		store.FetchByDiscordFunc = func(discordId string) (string, error) {
			return "ash", nil
		}
		store.AddSessionFunc = func(username string, token []byte) error {
			sessionFor = username
			return nil
		}

		token, err := OAuthLogin(store, "discord", "123")
		if err != nil {
			t.Errorf("expected success, got error: %v", err)
		}
		if token == "" {
			t.Errorf("expected token to be set on success")
		}
		if sessionFor != "ash" {
			t.Errorf("expected session for ash, got: %q", sessionFor)
		}
	})
	t.Run("GoogleUsesGoogleLookup", func(t *testing.T) {
		store := defaultMockOAuthStore()
		store.FetchByGoogleFunc = func(googleId string) (string, error) {
			return "ash", nil
		}

		_, err := OAuthLogin(store, "google", "g-123")
		if err != nil {
			t.Errorf("expected success, got error: %v", err)
		}
	})
	t.Run("NotLinked", func(t *testing.T) {
		store := defaultMockOAuthStore()
		store.FetchByDiscordFunc = func(discordId string) (string, error) {
			return "", sql.ErrNoRows
		}

		_, err := OAuthLogin(store, "discord", "123")
		expectErr(t, err, ErrOAuthNotLinked)
	})
	t.Run("DBUnexpectedError", func(t *testing.T) {
		dbErr := errors.New("some db error")
		store := defaultMockOAuthStore()
		store.FetchByDiscordFunc = func(discordId string) (string, error) {
			return "", dbErr
		}

		_, err := OAuthLogin(store, "discord", "123")
		expectErr(t, err, dbErr)
		if errors.Is(err, ErrOAuthNotLinked) {
			t.Errorf("expected DB error, got ErrOAuthNotLinked")
		}
	})
	t.Run("SessionFails", func(t *testing.T) {
		store := defaultMockOAuthStore()
		store.FetchByDiscordFunc = func(discordId string) (string, error) {
			return "ash", nil
		}
		store.AddSessionFunc = func(username string, token []byte) error {
			return errors.New("insert failed")
		}

		_, err := OAuthLogin(store, "discord", "123")
		expectErr(t, err, nil)
	})
	t.Run("EmptyExternalId", func(t *testing.T) {
		_, err := OAuthLogin(defaultMockOAuthStore(), "discord", "")
		expectErr(t, err, nil)
	})
	t.Run("UnknownProvider", func(t *testing.T) {
		_, err := OAuthLogin(defaultMockOAuthStore(), "github", "123")
		expectErr(t, err, ErrUnknownProvider)
	})
}

func TestOAuthLink(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var linkedId, linkedUser string
		store := defaultMockOAuthStore()
		store.FetchBySessionFunc = func(token []byte) (string, error) {
			return "ash", nil
		}
		store.AddGoogleFunc = func(googleId, username string) error {
			linkedId = googleId
			linkedUser = username
			return nil
		}

		err := OAuthLink(store, "google", "g-123", "dG9r")
		if err != nil {
			t.Errorf("expected success, got error: %v", err)
		}
		if linkedId != "g-123" || linkedUser != "ash" {
			t.Errorf("expected (g-123, ash), got: (%q, %q)", linkedId, linkedUser)
		}
	})
	// Verify that a space in the state is decoded as '+',
	// since base64 '+' arrives as a space when the state isn't URL-encoded.
	t.Run("SpaceInStateIsPlus", func(t *testing.T) {
		want := []byte{0xfb, 0xff} // base64 "+/8="
		store := defaultMockOAuthStore()
		store.FetchBySessionFunc = func(token []byte) (string, error) {
			if !bytes.Equal(token, want) {
				t.Errorf("expected decoded token %v, got: %v", want, token)
			}
			return "ash", nil
		}
		store.AddDiscordFunc = func(discordId, username string) error {
			return nil
		}

		err := OAuthLink(store, "discord", "123", " /8=")
		if err != nil {
			t.Errorf("expected success, got error: %v", err)
		}
	})
	t.Run("InvalidState", func(t *testing.T) {
		err := OAuthLink(defaultMockOAuthStore(), "discord", "123", "not base64!")
		expectErr(t, err, nil)
	})
	t.Run("SessionNotFound", func(t *testing.T) {
		store := defaultMockOAuthStore()
		store.FetchBySessionFunc = func(token []byte) (string, error) {
			return "", sql.ErrNoRows
		}

		err := OAuthLink(store, "discord", "123", "dG9r")
		expectErr(t, err, sql.ErrNoRows)
	})
	t.Run("AddFails", func(t *testing.T) {
		store := defaultMockOAuthStore()
		store.FetchBySessionFunc = func(token []byte) (string, error) {
			return "ash", nil
		}
		store.AddDiscordFunc = func(discordId, username string) error {
			return errors.New("already linked")
		}

		err := OAuthLink(store, "discord", "123", "dG9r")
		expectErr(t, err, nil)
	})
	t.Run("EmptyExternalId", func(t *testing.T) {
		err := OAuthLink(defaultMockOAuthStore(), "discord", "", "dG9r")
		expectErr(t, err, nil)
	})
	t.Run("UnknownProvider", func(t *testing.T) {
		store := defaultMockOAuthStore()
		store.FetchBySessionFunc = func(token []byte) (string, error) {
			return "ash", nil
		}

		err := OAuthLink(store, "github", "123", "dG9r")
		expectErr(t, err, ErrUnknownProvider)
	})
}
