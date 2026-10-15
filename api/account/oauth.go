package account

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OAuthLoginStore defines the database operations needed for OAuth login.
type OAuthLoginStore interface {
	GenerateTokenForUsernameStore
	FetchUsernameByDiscordId(discordId string) (string, error)
	FetchUsernameByGoogleId(googleId string) (string, error)
}

var (
	ErrOAuthCancelled  = errors.New("user cancelled authorization")
	ErrOAuthNotLinked  = errors.New("oauth account not linked")
	ErrUnknownProvider = errors.New("unknown oauth provider")
)

// oauthHTTPClient is used for requests to OAuth providers. It has a timeout
// so that requests can't stay pending forever.
var oauthHTTPClient = &http.Client{Timeout: 10 * time.Second}

// codeFromCallback extracts the authorization code from a provider callback.
func codeFromCallback(r *http.Request) (string, error) {
	q := r.URL.Query()
	if q.Get("error") != "" { // e.g. access_denied when the user clicks Cancel
		return "", ErrOAuthCancelled
	}
	code := q.Get("code")
	if code == "" {
		return "", errors.New("code is empty")
	}
	return code, nil
}

// OAuthLogin finds the account linked to externalId and creates a session for it.
// It returns [ErrOAuthNotLinked] if no account has that external ID.
func OAuthLogin[T OAuthLoginStore](store T, provider, externalId string) (string, error) {
	if externalId == "" {
		return "", errors.New("empty external id")
	}

	var username string
	var err error
	switch provider {
	case "discord":
		username, err = store.FetchUsernameByDiscordId(externalId)
	case "google":
		username, err = store.FetchUsernameByGoogleId(externalId)
	default:
		return "", ErrUnknownProvider
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrOAuthNotLinked
		}
		return "", err
	}

	return GenerateTokenForUsername(store, username)
}

// OAuthLinkStore defines the database operations needed for linking OAuth accounts.
type OAuthLinkStore interface {
	FetchUsernameBySessionToken(token []byte) (string, error)
	AddDiscordIdByUsername(discordId, username string) error
	AddGoogleIdByUsername(googleId, username string) error
}

// OAuthLink attaches externalId to the account whose session token is in state.
func OAuthLink[T OAuthLinkStore](store T, provider, externalId, state string) error {
	if externalId == "" {
		return errors.New("empty external id")
	}

	// '+' in base64 arrives as ' ' when the state isn't URL-encoded
	sessionToken, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(state, " ", "+"))
	if err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	username, err := store.FetchUsernameBySessionToken(sessionToken)
	if err != nil {
		return fmt.Errorf("session lookup failed: %w", err)
	}

	switch provider {
	case "discord":
		return store.AddDiscordIdByUsername(externalId, username)
	case "google":
		return store.AddGoogleIdByUsername(externalId, username)
	default:
		return ErrUnknownProvider
	}
}
