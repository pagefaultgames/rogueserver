package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Flashfyre/pokerogue-server/db"
	"github.com/go-co-op/gocron"
)

var (
	playerCountScheduler = gocron.NewScheduler(time.UTC)
	playerCount          = 0
)

func SchedulePlayerCountRefresh() {
	playerCountScheduler.Every(10).Second().Do(UpdatePlayerCount)
	playerCountScheduler.StartAsync()
}

func UpdatePlayerCount() {
	var err error
	playerCount, err = db.FetchPlayerCount()
	if err != nil {
		log.Print(err.Error())
	}
}

// /game/playercount - get player count

func (s *Server) HandlePlayerCountGet(w http.ResponseWriter, r *http.Request) {
	response, err := json.Marshal(playerCount)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}
