package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Flashfyre/pokerogue-server/db"
)

// /game/playercount - get player count

func (s *Server) HandlePlayerCountGet(w http.ResponseWriter, r *http.Request) {
	playerCount, err := db.FetchPlayerCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(playerCount)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}
