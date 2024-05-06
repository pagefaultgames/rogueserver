/*
	Copyright (C) 2024  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/markbates/goth/gothic"
	"github.com/pagefaultgames/rogueserver/api/account"
	"github.com/pagefaultgames/rogueserver/api/daily"
	"github.com/pagefaultgames/rogueserver/api/savedata"
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

/*
	The caller of endpoint handler functions are responsible for extracting the necessary data from the request.
	Handler functions are responsible for checking the validity of this data and returning a result or error.
	Handlers should not return serialized JSON, instead return the struct itself.
*/

var (
	user = string("")
)

// account

func handleAccountInfo(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusUnauthorized)
		return
	}

	username, err := db.FetchUsernameFromUUID(uuid)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	response, err := account.Info(username, uuid)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, r, response)
}

func handleAccountRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to parse request form: %s", err), http.StatusBadRequest)
		return
	}

	err = account.Register(r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleAccountLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to parse request form: %s", err), http.StatusBadRequest)
		return
	}

	response, err := account.Login(r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, r, response)
}

func handleAccountChangePW(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to parse request form: %s", err), http.StatusBadRequest)
		return
	}

	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusUnauthorized)
		return
	}

	err = account.ChangePW(uuid, r.Form.Get("password"))
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleAccountLogout(w http.ResponseWriter, r *http.Request) {
	token, err := tokenFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	err = account.Logout(token)
	if err != nil {
		// also possible for InternalServerError but that's unlikely unless the server blew up
		httpError(w, r, err, http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// game
func handleGameTitleStats(w http.ResponseWriter, r *http.Request) {
	stats := defs.TitleStats{
		PlayerCount: playerCount,
		BattleCount: battleCount,
	}

	writeJSON(w, r, stats)
}

func handleGameClassicSessionCount(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.Itoa(classicSessionCount)))
}

func handleSession(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusUnauthorized)
		return
	}

	slot, err := strconv.Atoi(r.URL.Query().Get("slot"))
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	if slot < 0 || slot >= defs.SessionSlotCount {
		httpError(w, r, fmt.Errorf("slot id %d out of range", slot), http.StatusBadRequest)
		return
	}

	if !r.URL.Query().Has("clientSessionId") {
		httpError(w, r, fmt.Errorf("missing clientSessionId"), http.StatusBadRequest)
		return
	}

	err = db.UpdateActiveSession(uuid, r.URL.Query().Get("clientSessionId"))
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to update active session: %s", err), http.StatusBadRequest)
		return
	}

	switch r.PathValue("action") {
	case "get":
		save, err := savedata.GetSession(uuid, slot)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		writeJSON(w, r, save)
	case "update":
		var session defs.SessionSaveData
		err = json.NewDecoder(r.Body).Decode(&session)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
			return
		}

		existingSave, err := savedata.GetSession(uuid, slot)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			httpError(w, r, fmt.Errorf("failed to retrieve session save data: %s", err), http.StatusInternalServerError)
			return
		} else {
			if existingSave.Seed == session.Seed && existingSave.WaveIndex > session.WaveIndex {
				httpError(w, r, fmt.Errorf("session out of date: existing wave index is greater"), http.StatusBadRequest)
				return
			}
		}

		err = savedata.UpdateSession(uuid, slot, session)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to put session data: %s", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "clear":
		var session defs.SessionSaveData
		err = json.NewDecoder(r.Body).Decode(&session)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
			return
		}

		seed, err := db.GetDailyRunSeed()
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		resp, err := savedata.Clear(uuid, slot, seed, session)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		writeJSON(w, r, resp)
	case "newclear":
		resp, err := savedata.NewClear(uuid, slot)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to read new clear: %s", err), http.StatusInternalServerError)
			return
		}

		writeJSON(w, r, resp)
	case "delete":
		err := savedata.DeleteSession(uuid, slot)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		httpError(w, r, fmt.Errorf("unknown action"), http.StatusBadRequest)
		return
	}
}

type CombinedSaveData struct {
	System          defs.SystemSaveData  `json:"system"`
	Session         defs.SessionSaveData `json:"session"`
	SessionSlotId   int                  `json:"sessionSlotId"`
	ClientSessionId string               `json:"clientSessionId"`
}

// TODO wrap this in a transaction
func handleUpdateAll(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusUnauthorized)
		return
	}

	var data CombinedSaveData
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
		return
	}

	if data.ClientSessionId == "" {
		httpError(w, r, fmt.Errorf("missing clientSessionId"), http.StatusBadRequest)
		return
	}

	var active bool
	active, err = db.IsActiveSession(uuid, data.ClientSessionId)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to check active session: %s", err), http.StatusBadRequest)
		return
	}

	if !active {
		httpError(w, r, fmt.Errorf("session out of date: not active"), http.StatusBadRequest)
		return
	}

	storedTrainerId, storedSecretId, err := db.FetchTrainerIds(uuid)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	if storedTrainerId > 0 || storedSecretId > 0 {
		if data.System.TrainerId != storedTrainerId || data.System.SecretId != storedSecretId {
			httpError(w, r, fmt.Errorf("session out of date: stored trainer or secret ID does not match"), http.StatusBadRequest)
			return
		}
	} else {
		err = db.UpdateTrainerIds(data.System.TrainerId, data.System.SecretId, uuid)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	existingPlaytime, err := db.RetrievePlaytime(uuid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		httpError(w, r, fmt.Errorf("failed to retrieve playtime: %s", err), http.StatusInternalServerError)
		return
	} else {
		playtime, ok := data.System.GameStats.(map[string]interface{})["playTime"].(float64)
		if !ok {
			httpError(w, r, fmt.Errorf("no playtime found"), http.StatusBadRequest)
			return
		}

		if float64(existingPlaytime) > playtime {
			httpError(w, r, fmt.Errorf("session out of date: existing playtime is greater"), http.StatusBadRequest)
			return
		}
	}

	existingSave, err := savedata.GetSession(uuid, data.SessionSlotId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		httpError(w, r, fmt.Errorf("failed to retrieve session save data: %s", err), http.StatusInternalServerError)
		return
	} else {
		if existingSave.Seed == data.Session.Seed && existingSave.WaveIndex > data.Session.WaveIndex {
			httpError(w, r, fmt.Errorf("session out of date: existing wave index is greater"), http.StatusBadRequest)
			return
		}
	}

	err = savedata.Update(uuid, data.SessionSlotId, data.Session)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = savedata.Update(uuid, 0, data.System)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type SystemVerifyResponse struct {
	Valid      bool                `json:"valid"`
	SystemData defs.SystemSaveData `json:"systemData"`
}

func handleSystem(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusUnauthorized)
		return
	}

	var active bool
	if !r.URL.Query().Has("clientSessionId") {
		httpError(w, r, fmt.Errorf("missing clientSessionId"), http.StatusBadRequest)
		return
	}

	active, err = db.IsActiveSession(uuid, r.URL.Query().Get("clientSessionId"))
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to check active session: %s", err), http.StatusBadRequest)
		return
	}

	switch r.PathValue("action") {
	case "get":
		if !active {
			err = db.UpdateActiveSession(uuid, r.URL.Query().Get("clientSessionId"))
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to update active session: %s", err), http.StatusBadRequest)
				return
			}
		}

		save, err := savedata.GetSystem(uuid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				httpError(w, r, err, http.StatusInternalServerError)
			}

			return
		}

		writeJSON(w, r, save)
	case "update":
		if !active {
			httpError(w, r, fmt.Errorf("session out of date: not active"), http.StatusBadRequest)
			return
		}

		var system defs.SystemSaveData
		err = json.NewDecoder(r.Body).Decode(&system)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
			return
		}

		existingPlaytime, err := db.RetrievePlaytime(uuid)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			httpError(w, r, fmt.Errorf("failed to retrieve playtime: %s", err), http.StatusInternalServerError)
			return
		} else {
			playtime, ok := system.GameStats.(map[string]interface{})["playTime"].(float64)
			if !ok {
				httpError(w, r, fmt.Errorf("no playtime found"), http.StatusBadRequest)
				return
			}

			if float64(existingPlaytime) > playtime {
				httpError(w, r, fmt.Errorf("session out of date: existing playtime is greater"), http.StatusBadRequest)
				return
			}
		}

		err = savedata.UpdateSystem(uuid, system)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to put system data: %s", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	case "verify":
		response := SystemVerifyResponse{
			Valid: active,
		}

		// not valid, send server state
		if !active {
			err = db.UpdateActiveSession(uuid, r.URL.Query().Get("clientSessionId"))
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to update active session: %s", err), http.StatusBadRequest)
				return
			}

			var storedSaveData defs.SystemSaveData
			storedSaveData, err = db.ReadSystemSaveData(uuid)
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to read session save data: %s", err), http.StatusInternalServerError)
				return
			}

			response.SystemData = storedSaveData
		}

		writeJSON(w, r, response)
	case "delete":
		err := savedata.DeleteSystem(uuid)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		httpError(w, r, fmt.Errorf("unknown action"), http.StatusBadRequest)
		return
	}
}

// daily
func handleDailySeed(w http.ResponseWriter, r *http.Request) {
	seed, err := db.GetDailyRunSeed()
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(seed))
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to write seed: %s", err), http.StatusInternalServerError)
	}
}

func handleDailyRankings(w http.ResponseWriter, r *http.Request) {
	var err error

	var category int
	if r.URL.Query().Has("category") {
		category, err = strconv.Atoi(r.URL.Query().Get("category"))
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to convert category: %s", err), http.StatusBadRequest)
			return
		}
	}

	page := 1
	if r.URL.Query().Has("page") {
		page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to convert page: %s", err), http.StatusBadRequest)
			return
		}
	}

	rankings, err := daily.Rankings(category, page)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, r, rankings)
}

func handleDailyRankingPageCount(w http.ResponseWriter, r *http.Request) {
	var category int
	if r.URL.Query().Has("category") {
		var err error
		category, err = strconv.Atoi(r.URL.Query().Get("category"))
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to convert category: %s", err), http.StatusBadRequest)
			return
		}
	}

	count, err := daily.RankingPageCount(category)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
	}

	w.Write([]byte(strconv.Itoa(count)))
}

// redirect link after authorizing application link
func handleProviderCallback(w http.ResponseWriter, r *http.Request) {
	gothic.GetProviderName = func(r *http.Request) (string, error) { return r.PathValue("provider"), nil }

	// called again with code after authorization
	code := r.URL.Query().Get("code")
	if code != "" {
		userId, err := db.FetchDiscordIdByUsername(user)
		if err != nil {

		}
		defer http.Redirect(w, r, "http://localhost:8000", http.StatusSeeOther)
	}

	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Println("callback err", w, r)
		return
	} else {
		err := db.AddDiscordAuthByUsername(gothUser.UserID, user)
		if err != nil {
			log.Println("error adding Discord Auth to database")
			return
		}
	}
	log.Println("user", gothUser.UserID)
}

func handleProviderLink(w http.ResponseWriter, r *http.Request) {
	gothic.GetProviderName = func(r *http.Request) (string, error) { return r.PathValue("provider"), nil }
	username := r.URL.Query().Get("username")
	// username recorded prior to authorization
	if username != "" {
		user = username
	}
	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		log.Print("gothUser:", gothUser.Name)
	} else {
		gothic.BeginAuthHandler(w, r)
	}

}

func handleProviderLogout(w http.ResponseWriter, r *http.Request) {
	gothic.GetProviderName = func(r *http.Request) (string, error) { return r.PathValue("provider"), nil }
	gothic.Logout(w, r)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
