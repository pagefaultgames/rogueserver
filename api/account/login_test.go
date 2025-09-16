package account

import (
	"database/sql"
	"errors"
	"testing"
)

func defaultMockStore() *mockDBAccountStore {
	return &mockDBAccountStore{
		FetchFunc: func(username string) ([]byte, []byte, error) {
			return []byte("key"), []byte("salt"), nil
		},
		AddSessionFunc: func(username string, token []byte) error { return nil },
	}
}

type mockDBAccountStore struct {
	FetchFunc      func(username string) ([]byte, []byte, error)
	AddSessionFunc func(username string, token []byte) error
}

func (m *mockDBAccountStore) FetchAccountKeySaltFromUsername(username string) ([]byte, []byte, error) {
	return m.FetchFunc(username)
}
func (m *mockDBAccountStore) AddAccountSession(username string, token []byte) error {
	return m.AddSessionFunc(username, token)
}

func TestLogin(t *testing.T) {
	t.Run("UsernameMinLength", func(t *testing.T) {
		store := defaultMockStore()
		_, err := Login(store, "a", "password123")
		if err == nil {
			t.Errorf("expected error due to password mismatch or DB, got nil")
		}
	})
	t.Run("UsernameMaxLength", func(t *testing.T) {
		uname := "abcdefghijklmnop"
		store := defaultMockStore()
		_, err := Login(store, uname, "password123")
		if err == nil {
			t.Errorf("expected error due to password mismatch or DB, got nil")
		}
	})
	t.Run("UsernameTooLong", func(t *testing.T) {
		uname := "abcdefghijklmnopq"
		store := defaultMockStore()
		_, err := Login(store, uname, "password123")
		if err == nil || err.Error() != "invalid username" {
			t.Errorf("expected invalid username error for too long username, got: %v", err)
		}
	})
	t.Run("UsernameWithInvalidChars", func(t *testing.T) {
		store := defaultMockStore()
		_, err := Login(store, "user!@#", "password123")
		if err == nil || err.Error() != "invalid username" {
			t.Errorf("expected invalid username error for special chars, got: %v", err)
		}
	})
	t.Run("EmptyUsername", func(t *testing.T) {
		store := defaultMockStore()
		_, err := Login(store, "", "password123")
		if err == nil || err.Error() != "invalid username" {
			t.Errorf("expected invalid username error for empty username, got: %v", err)
		}
	})
	t.Run("EmptyPassword", func(t *testing.T) {
		store := defaultMockStore()
		_, err := Login(store, "validuser", "")
		if err == nil || err.Error() != "invalid password" {
			t.Errorf("expected invalid password error for empty password, got: %v", err)
		}
	})
	t.Run("MinPasswordLength", func(t *testing.T) {
		store := defaultMockStore()
		_, err := Login(store, "validuser", "123456")
		if err == nil {
			t.Errorf("expected error due to password mismatch or DB, got nil")
		}
	})
	t.Run("PasswordWithSpecialChars", func(t *testing.T) {
		store := defaultMockStore()
		_, err := Login(store, "validuser", "p@$$w0rd!")
		if err == nil {
			t.Errorf("expected error due to password mismatch or DB, got nil")
		}
	})
	t.Run("DBUnexpectedError", func(t *testing.T) {
		store := defaultMockStore()
		store.FetchFunc = func(username string) ([]byte, []byte, error) {
			return nil, nil, errors.New("some db error")
		}
		_, err := Login(store, "validuser", "password123")
		if err == nil || err.Error() != "some db error" {
			t.Errorf("expected DB error to propagate, got: %v", err)
		}
	})
	t.Run("InvalidUsername", func(t *testing.T) {
		store := defaultMockStore()
		_, err := Login(store, "!invaliduser", "password123")
		if err == nil || err.Error() != "invalid username" {
			t.Errorf("expected invalid username error, got: %v", err)
		}
	})
	t.Run("ShortPassword", func(t *testing.T) {
		store := defaultMockStore()
		_, err := Login(store, "validuser", "123")
		if err == nil || err.Error() != "invalid password" {
			t.Errorf("expected invalid password error, got: %v", err)
		}
	})
	t.Run("AccountDoesNotExist", func(t *testing.T) {
		store := defaultMockStore()
		store.FetchFunc = func(username string) ([]byte, []byte, error) {
			return nil, nil, sql.ErrNoRows
		}
		_, err := Login(store, "nonexistent", "password123")
		if err == nil || err.Error() != "account doesn't exist" {
			t.Errorf("expected account doesn't exist error, got: %v", err)
		}
	})
	t.Run("PasswordMismatch", func(t *testing.T) {
		correctSalt := []byte("somesalt")
		correctKey := []byte("correctkey")
		store := defaultMockStore()
		store.FetchFunc = func(username string) ([]byte, []byte, error) {
			return correctKey, correctSalt, nil
		}
		_, err := Login(store, "validuser", "wrongpassword")
		if err == nil || err.Error() != "password doesn't match" {
			t.Errorf("expected password doesn't match error, got: %v", err)
		}
	})
	t.Run("Success", func(t *testing.T) {
		correctSalt := []byte("somesalt")
		password := "goodpassword"
		correctKey := deriveArgon2IDKey([]byte(password), correctSalt)
		store := defaultMockStore()
		store.FetchFunc = func(username string) ([]byte, []byte, error) {
			return correctKey, correctSalt, nil
		}
		store.AddSessionFunc = func(username string, token []byte) error {
			return nil
		}
		resp, err := Login(store, "validuser", password)
		if err != nil {
			t.Errorf("expected success, got error: %v", err)
		}
		if resp.Token == "" {
			t.Errorf("expected token to be set on success")
		}
	})
}
