package api

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/Flashfyre/pokerogue-server/db"
)

func getUsernameFromRequest(r *http.Request) (string, error) {
	if r.Header.Get("Authorization") == "" {
		return "", fmt.Errorf("missing token")
	}

	token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %s", err)
	}

	if len(token) != 32 {
		return "", fmt.Errorf("invalid token length: got %d, expected 32", len(token))
	}

	username, err := db.FetchUsernameFromToken(token)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %s", err)
	}

	return username, nil
}

func getUUIDFromRequest(r *http.Request) ([]byte, error) {
	if r.Header.Get("Authorization") == "" {
		return nil, fmt.Errorf("missing token")
	}

	token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %s", err)
	}

	if len(token) != 32 {
		return nil, fmt.Errorf("invalid token length: got %d, expected 32", len(token))
	}

	uuid, err := db.FetchUUIDFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %s", err)
	}

	return uuid, nil
}
