package api

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/pagefaultgames/pokerogue-server/api/account"
	"github.com/pagefaultgames/pokerogue-server/api/daily"
	"github.com/pagefaultgames/pokerogue-server/db"
)

func Init() {
	scheduleStatRefresh()
	daily.Init()
}

func getTokenFromRequest(r *http.Request) ([]byte, error) {
	if r.Header.Get("Authorization") == "" {
		return nil, fmt.Errorf("missing token")
	}

	token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %s", err)
	}
	
	if len(token) != account.TokenSize {
		return nil, fmt.Errorf("invalid token length: got %d, expected %d", len(token), account.TokenSize)
	}

	return token, nil
}

func getUsernameFromRequest(r *http.Request) (string, error) {
	token, err := getTokenFromRequest(r)
	if err != nil {
		return "", err
	}

	username, err := db.FetchUsernameFromToken(token)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %s", err)
	}

	return username, nil
}

func getUUIDFromRequest(r *http.Request) ([]byte, error) {
	token, err := getTokenFromRequest(r)
	if err != nil {
		return nil, err
	}

	uuid, err := db.FetchUUIDFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %s", err)
	}

	return uuid, nil
}
