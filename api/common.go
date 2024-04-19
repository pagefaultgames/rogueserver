package api

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/pagefaultgames/pokerogue-server/api/account"
	"github.com/pagefaultgames/pokerogue-server/api/daily"
	"github.com/pagefaultgames/pokerogue-server/db"
)

func Init(mux *http.ServeMux) {
	scheduleStatRefresh()
	daily.Init()

	// account
	mux.HandleFunc("GET /api/account/info", handleAccountInfo)
	mux.HandleFunc("POST /api/account/register", handleAccountRegister)
	mux.HandleFunc("POST /api/account/login", handleAccountLogin)
	mux.HandleFunc("GET /api/account/logout", handleAccountLogout)

	// game
	mux.HandleFunc("GET /api/game/playercount", handleGamePlayerCount)
	mux.HandleFunc("GET /api/game/titlestats", handleGameTitleStats)
	mux.HandleFunc("GET /api/game/classicsessioncount", handleGameClassicSessionCount)

	// savedata
	mux.HandleFunc("GET /api/savedata/get", handleSaveData)
	mux.HandleFunc("POST /api/savedata/update", handleSaveData)
	mux.HandleFunc("GET /api/savedata/delete", handleSaveData)
	mux.HandleFunc("POST /api/savedata/clear", handleSaveData)

	// daily
	mux.HandleFunc("GET /api/daily/seed", handleDailySeed)
	mux.HandleFunc("GET /api/daily/rankings", handleDailyRankings)
	mux.HandleFunc("GET /api/daily/rankingpagecount", handleDailyRankingPageCount)
}

func getUsernameFromRequest(r *http.Request) (string, error) {
	if r.Header.Get("Authorization") == "" {
		return "", fmt.Errorf("missing token")
	}

	token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %s", err)
	}

	if len(token) != account.TokenSize {
		return "", fmt.Errorf("invalid token length: got %d, expected %d", len(token), account.TokenSize)
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

	if len(token) != account.TokenSize {
		return nil, fmt.Errorf("invalid token length: got %d, expected %d", len(token), account.TokenSize)
	}

	uuid, err := db.FetchUUIDFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %s", err)
	}

	return uuid, nil
}
