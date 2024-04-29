package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/pagefaultgames/rogueserver/api/account"
	"github.com/pagefaultgames/rogueserver/api/daily"
	"github.com/pagefaultgames/rogueserver/db"
)

func Init(mux *http.ServeMux) {
	scheduleStatRefresh()
	daily.Init()

	// account
	mux.HandleFunc("GET /account/info", handleAccountInfo)
	mux.HandleFunc("POST /account/register", handleAccountRegister)
	mux.HandleFunc("POST /account/login", handleAccountLogin)
	mux.HandleFunc("POST /account/changepw", handleAccountChangePW)
	mux.HandleFunc("GET /account/logout", handleAccountLogout)

	// game
	mux.HandleFunc("GET /game/playercount", handleGamePlayerCount)
	mux.HandleFunc("GET /game/titlestats", handleGameTitleStats)
	mux.HandleFunc("GET /game/classicsessioncount", handleGameClassicSessionCount)

	// savedata
	mux.HandleFunc("GET /savedata/get", handleSaveData)
	mux.HandleFunc("POST /savedata/update", handleSaveData)
	mux.HandleFunc("GET /savedata/delete", handleSaveData)
	mux.HandleFunc("POST /savedata/clear", handleSaveData)

	// daily
	mux.HandleFunc("GET /daily/seed", handleDailySeed)
	mux.HandleFunc("GET /daily/rankings", handleDailyRankings)
	mux.HandleFunc("GET /daily/rankingpagecount", handleDailyRankingPageCount)
}

func tokenFromRequest(r *http.Request) ([]byte, error) {
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

func usernameFromRequest(r *http.Request) (string, error) {
	token, err := tokenFromRequest(r)
	if err != nil {
		return "", err
	}

	username, err := db.FetchUsernameFromToken(token)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %s", err)
	}

	return username, nil
}

func uuidFromRequest(r *http.Request) ([]byte, error) {
	token, err := tokenFromRequest(r)
	if err != nil {
		return nil, err
	}

	uuid, err := db.FetchUUIDFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %s", err)
	}

	return uuid, nil
}

func httpError(w http.ResponseWriter, r *http.Request, err error, code int) {
	log.Printf("%s: %s\n", r.URL.Path, err)
	http.Error(w, err.Error(), code)
}