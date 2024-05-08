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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

// account

func handleAccountInfo(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
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

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

func handleAccountChangePW(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to parse request form: %s", err), http.StatusBadRequest)
		return
	}

	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
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
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// game

func handleGamePlayerCount(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.Itoa(playerCount)))
}

func handleGameTitleStats(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(defs.TitleStats{
		PlayerCount: playerCount,
		BattleCount: battleCount,
	})
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

func handleGameClassicSessionCount(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.Itoa(classicSessionCount)))
}

func handleSaveData(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	datatype := -1
	if r.URL.Query().Has("datatype") {
		datatype, err = strconv.Atoi(r.URL.Query().Get("datatype"))
		if err != nil {
			httpError(w, r, err, http.StatusBadRequest)
			return
		}
	}

	var slot int
	if r.URL.Query().Has("slot") {
		slot, err = strconv.Atoi(r.URL.Query().Get("slot"))
		if err != nil {
			httpError(w, r, err, http.StatusBadRequest)
			return
		}
	}

	var save any
	// /savedata/get and /savedata/delete specify datatype, but don't expect data in body
	if r.URL.Path != "/savedata/get" && r.URL.Path != "/savedata/delete" {
		if datatype == 0 {
			var system defs.SystemSaveData
			err = json.NewDecoder(r.Body).Decode(&system)
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
				return
			}

			save = system
			// /savedata/clear doesn't specify datatype, it is assumed to be 1 (session)
		} else if datatype == 1 || r.URL.Path == "/savedata/clear" {
			var session defs.SessionSaveData
			err = json.NewDecoder(r.Body).Decode(&session)
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
				return
			}

			save = session
		}
	}

	var token []byte
	token, err = tokenFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	var active bool
	if r.URL.Path == "/savedata/get" {
		if datatype == 0 {
			err = db.UpdateActiveSession(uuid, token)
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to update active session: %s", err), http.StatusBadRequest)
				return
			}
		}
	} else {
		active, err = db.IsActiveSession(token)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to check active session: %s", err), http.StatusBadRequest)
			return
		}

		// TODO: make this not suck
		if !active && r.URL.Path != "/savedata/clear" {
			httpError(w, r, fmt.Errorf("session out of date"), http.StatusBadRequest)
			return
		}

		var trainerId, secretId int

		if r.URL.Path != "/savedata/update" || datatype == 1 {
			if r.URL.Query().Has("trainerId") && r.URL.Query().Has("secretId") {
				trainerId, err = strconv.Atoi(r.URL.Query().Get("trainerId"))
				if err != nil {
					httpError(w, r, err, http.StatusBadRequest)
					return
				}

				secretId, err = strconv.Atoi(r.URL.Query().Get("secretId"))
				if err != nil {
					httpError(w, r, err, http.StatusBadRequest)
					return
				}
			}
		} else {
			trainerId = save.(defs.SystemSaveData).TrainerId
			secretId = save.(defs.SystemSaveData).SecretId
		}

		storedTrainerId, storedSecretId, err := db.FetchTrainerIds(uuid)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		if storedTrainerId > 0 || storedSecretId > 0 {
			if trainerId != storedTrainerId || secretId != storedSecretId {
				httpError(w, r, fmt.Errorf("session out of date"), http.StatusBadRequest)
				return
			}
		} else {
			db.UpdateTrainerIds(trainerId, secretId, uuid)
		}
	}

	switch r.URL.Path {
	case "/savedata/get":
		save, err = savedata.Get(uuid, datatype, slot)
	case "/savedata/update":
		err = savedata.Update(uuid, slot, save)
	case "/savedata/delete":
		err = savedata.Delete(uuid, datatype, slot)
	case "/savedata/clear":
		if !active {
			// TODO: make this not suck
			save = savedata.ClearResponse{Error: "session out of date"}
			break
		}

		s, ok := save.(defs.SessionSaveData)
		if !ok {
			err = fmt.Errorf("save data is not type SessionSaveData")
			break
		}

		// doesn't return a save, but it works
		var seed string
		seed, err = db.GetDailyRunSeed()
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}
		save, err = savedata.Clear(uuid, slot, seed, s)
	}
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	if save == nil || r.URL.Path == "/savedata/update" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = json.NewEncoder(w).Encode(save)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

// daily

func handleDailySeed(w http.ResponseWriter, r *http.Request) {
	seed, err := db.GetDailyRunSeed()
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}
	bytes, err := base64.StdEncoding.DecodeString(seed)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
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

	err = json.NewEncoder(w).Encode(rankings)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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
