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
func handleGameTitleStats(w http.ResponseWriter, r *http.Request) {
	stats := defs.TitleStats{
		PlayerCount: playerCount,
		BattleCount: battleCount,
	}

	writeJSON(w, r, stats)
}

func handleGameClassicSessionCount(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(strconv.Itoa(classicSessionCount)))
}

func handleGetSessionData(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	var slot int
	if r.URL.Query().Has("slot") {
		slot, err = strconv.Atoi(r.URL.Query().Get("slot"))
		if err != nil {
			httpError(w, r, err, http.StatusBadRequest)
			return
		}
	}

	var clientSessionId string
	if r.URL.Query().Has("clientSessionId") {
		clientSessionId = r.URL.Query().Get("clientSessionId")
	} else {
		httpError(w, r, fmt.Errorf("missing clientSessionId"), http.StatusBadRequest)
	}

	err = db.UpdateActiveSession(uuid, clientSessionId)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to update active session: %s", err), http.StatusBadRequest)
		return
	}

	var save any
	save, err = savedata.Get(uuid, 1, slot)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, r, save)
}

const legacyClientSessionId = "LEGACY_CLIENT"

func legacyHandleGetSaveData(w http.ResponseWriter, r *http.Request) {
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
	if datatype == 0 {
		err = db.UpdateActiveSession(uuid, legacyClientSessionId) // we dont have a client id
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to update active session: %s", err), http.StatusBadRequest)
			return
		}
	}

	save, err = savedata.Get(uuid, datatype, slot)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, r, save)
}

// FIXME UNFINISHED!!!
/*func clearSessionData(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
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
	var session defs.SessionSaveData
	err = json.NewDecoder(r.Body).Decode(&session)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
		return
	}

	save = session

	var active bool
	active, err = db.IsActiveSession(uuid, legacyClientSessionId) //TODO unfinished, read token from query
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to check active session: %s", err), http.StatusBadRequest)
		return
	}

	var trainerId, secretId int
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

	storedTrainerId, storedSecretId, err := db.FetchTrainerIds(uuid)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	if storedTrainerId > 0 || storedSecretId > 0 {
		if trainerId != storedTrainerId || secretId != storedSecretId {
			httpError(w, r, fmt.Errorf("session out of date: stored trainer or secret ID does not match"), http.StatusBadRequest)
			return
		}
	} else {
		err = db.UpdateTrainerIds(trainerId, secretId, uuid)
		if err != nil {
			httpError(w, r, fmt.Errorf("unable to update trainer ID: %s", err), http.StatusInternalServerError)
			return
		}
	}

	if !active {
		save = savedata.ClearResponse{Error: "session out of date: not active"}
	}

	var seed string
	seed, err = db.GetDailyRunSeed()
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	response, err := savedata.Clear(uuid, slot, seed, save.(defs.SessionSaveData))
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, r, response)
}

// FIXME UNFINISHED!!!
func deleteSystemSave(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	datatype := 0
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

	var active bool
	active, err = db.IsActiveSession(uuid, legacyClientSessionId) //TODO unfinished, read token from query
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to check active session: %s", err), http.StatusInternalServerError)
		return
	}

	if !active {
		httpError(w, r, fmt.Errorf("session out of date: not active"), http.StatusBadRequest)
		return
	}

	var trainerId, secretId int

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

	storedTrainerId, storedSecretId, err := db.FetchTrainerIds(uuid)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	if storedTrainerId > 0 || storedSecretId > 0 {
		if trainerId != storedTrainerId || secretId != storedSecretId {
			httpError(w, r, fmt.Errorf("session out of date: stored trainer or secret ID does not match"), http.StatusBadRequest)
			return
		}
	} else {
		if err := db.UpdateTrainerIds(trainerId, secretId, uuid); err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	err = savedata.Delete(uuid, datatype, slot)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}*/

func legacyHandleSaveData(w http.ResponseWriter, r *http.Request) {
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

	var clientSessionId string
	if r.URL.Query().Has("clientSessionId") {
		clientSessionId = r.URL.Query().Get("clientSessionId")
	}
	if clientSessionId == "" {
		clientSessionId = legacyClientSessionId
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

	var active bool
	if r.URL.Path == "/savedata/get" {
		if datatype == 0 {
			err = db.UpdateActiveSession(uuid, clientSessionId)
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to update active session: %s", err), http.StatusBadRequest)
				return
			}
		}
	} else {
		active, err = db.IsActiveSession(uuid, clientSessionId)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to check active session: %s", err), http.StatusBadRequest)
			return
		}

		// TODO: make this not suck
		if !active && r.URL.Path != "/savedata/clear" {
			httpError(w, r, fmt.Errorf("session out of date: not active"), http.StatusBadRequest)
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
				httpError(w, r, fmt.Errorf("session out of date: stored trainer or secret ID does not match"), http.StatusBadRequest)
				return
			}
		} else {
			if err := db.UpdateTrainerIds(trainerId, secretId, uuid); err != nil {
				httpError(w, r, err, http.StatusInternalServerError)
				return
			}
		}
	}

	switch r.URL.Path {
	case "/savedata/get":
		save, err = savedata.Get(uuid, datatype, slot)
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	case "/savedata/update":
		err = savedata.Update(uuid, slot, save)
	case "/savedata/delete":
		err = savedata.Delete(uuid, datatype, slot)
	case "/savedata/clear":
		if !active {
			// TODO: make this not suck
			save = savedata.ClearResponse{Error: "session out of date: not active"}
			break
		}

		var seed string
		seed, err = db.GetDailyRunSeed()
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		// doesn't return a save, but it works
		save, err = savedata.Clear(uuid, slot, seed, save.(defs.SessionSaveData))
	}
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	if save == nil || r.URL.Path == "/savedata/update" {
		w.WriteHeader(http.StatusOK)
		return
	}

	writeJSON(w, r, save)
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
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	var data CombinedSaveData
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
		return
	}
	if data.ClientSessionId == "" {
		data.ClientSessionId = legacyClientSessionId
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

	trainerId := data.System.TrainerId
	secretId := data.System.SecretId

	storedTrainerId, storedSecretId, err := db.FetchTrainerIds(uuid)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	if storedTrainerId > 0 || storedSecretId > 0 {
		if trainerId != storedTrainerId || secretId != storedSecretId {
			httpError(w, r, fmt.Errorf("session out of date: stored trainer or secret ID does not match"), http.StatusBadRequest)
			return
		}
	} else {
		if err = db.UpdateTrainerIds(trainerId, secretId, uuid); err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
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
	Valid      bool                 `json:"valid"`
	SystemData *defs.SystemSaveData `json:"systemData"`
}

type SystemVerifyRequest struct {
	ClientSessionId string `json:"clientSessionId"`
}

func handleSystemVerify(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	var input SystemVerifyRequest
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
		return
	}

	var active bool
	active, err = db.IsActiveSession(uuid, input.ClientSessionId)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to check active session: %s", err), http.StatusBadRequest)
		return
	}

	response := SystemVerifyResponse{
		Valid: active,
	}

	// not valid, send server state
	if !active {
		err = db.UpdateActiveSession(uuid, input.ClientSessionId)
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

		response.SystemData = &storedSaveData
	}

	err = db.UpdateAccountLastActivity(uuid)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to update account last activity: %s", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, r, response)
}

func handleGetSystemData(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	var clientSessionId string
	if r.URL.Query().Has("clientSessionId") {
		clientSessionId = r.URL.Query().Get("clientSessionId")
	} else {
		httpError(w, r, fmt.Errorf("missing clientSessionId"), http.StatusBadRequest)
	}

	err = db.UpdateActiveSession(uuid, clientSessionId)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to update active session: %s", err), http.StatusBadRequest)
		return
	}

	var save any //TODO this is always system save data
	save, err = savedata.Get(uuid, 0, 0)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			httpError(w, r, err, http.StatusInternalServerError)
		}

		return
	}
	//TODO apply vouchers

	writeJSON(w, r, save)
}

func legacyHandleNewClear(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuidFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	var slot int
	if r.URL.Query().Has("slot") {
		slot, err = strconv.Atoi(r.URL.Query().Get("slot"))
		if err != nil {
			httpError(w, r, err, http.StatusBadRequest)
			return
		}
	}

	newClear, err := savedata.NewClear(uuid, slot)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to read new clear: %s", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, r, newClear)
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

	_, _ = w.Write([]byte(strconv.Itoa(count)))
}
