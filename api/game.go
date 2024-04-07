package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Flashfyre/pokerogue-server/db"
	"github.com/Flashfyre/pokerogue-server/defs"
	"github.com/go-co-op/gocron"
)

var (
	statScheduler       = gocron.NewScheduler(time.UTC)
	playerCount         int
	battleCount         int
	classicSessionCount int
)

func ScheduleStatRefresh() {
	statScheduler.Every(10).Second().Do(updateStats)
	statScheduler.StartAsync()
}

func updateStats() {
	var err error
	playerCount, err = db.FetchPlayerCount()
	if err != nil {
		log.Print(err)
	}
	battleCount, err = db.FetchBattleCount()
	if err != nil {
		log.Print(err)
	}
	classicSessionCount, err = db.FetchClassicSessionCount()
	if err != nil {
		log.Print(err)
	}
}

// /game/playercount - get player count
func (s *Server) handlePlayerCountGet(w http.ResponseWriter, r *http.Request) {
	response, err := json.Marshal(playerCount)
	if err != nil {
		httpError(w, r, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

// /game/titlestats - get title stats
func (s *Server) handleTitleStatsGet(w http.ResponseWriter, r *http.Request) {
	titleStats := &defs.TitleStats{
		PlayerCount: playerCount,
		BattleCount: battleCount,
	}
	response, err := json.Marshal(titleStats)
	if err != nil {
		httpError(w, r, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

// /game/classicsessioncount - get classic session count
func (s *Server) handleClassicSessionCountGet(w http.ResponseWriter, r *http.Request) {
	response, err := json.Marshal(classicSessionCount)
	if err != nil {
		httpError(w, r, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}
