package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/pagefaultgames/pokerogue-server/api/account"
	"github.com/pagefaultgames/pokerogue-server/api/daily"
	"github.com/pagefaultgames/pokerogue-server/api/savedata"
	"github.com/pagefaultgames/pokerogue-server/defs"
)

/*
	The caller of endpoint handler functions are responsible for extracting the necessary data from the request.
	Handler functions are responsible for checking the validity of this data and returning a result or error.
	Handlers should not return serialized JSON, instead return the struct itself.
*/

func handleAccountInfo(w http.ResponseWriter, r *http.Request) {
	username, err := getUsernameFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

	uuid, err := getUUIDFromRequest(r) // lazy
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
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
}

func handleAccountLogout(w http.ResponseWriter, r *http.Request) {
	token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to decode token: %s", err), http.StatusBadRequest)
		return
	}

	err = account.Logout(token)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

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
}

func handleGameClassicSessionCount(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.Itoa(classicSessionCount)))
}

func handleSaveData(w http.ResponseWriter, r *http.Request) {
	uuid, err := getUUIDFromRequest(r)
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
	if r.URL.Path != "/api/savedata/get" && r.URL.Path != "/api/savedata/delete" {
		if datatype == 0 {
			var system defs.SystemSaveData
			err = json.NewDecoder(r.Body).Decode(&system)
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
				return
			}

			save = system
		// /savedata/clear doesn't specify datatype, it is assumed to be 1 (session)
		} else if datatype == 1 || r.URL.Path == "/api/savedata/clear" {
			var session defs.SessionSaveData
			err = json.NewDecoder(r.Body).Decode(&session)
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
				return
			}

			save = session
		}
	}

	switch r.URL.Path {
	case "/api/savedata/get":
		save, err = savedata.Get(uuid, datatype, slot)
	case "/api/savedata/update":
		err = savedata.Update(uuid, slot, save)
	case "/api/savedata/delete":
		err = savedata.Delete(uuid, datatype, slot)
	case "/api/savedata/clear":
		s, ok := save.(defs.SessionSaveData)
		if !ok {
			httpError(w, r, fmt.Errorf("save data is not type SessionSaveData"), http.StatusBadRequest)
			return
		}

		// doesn't return a save, but it works
		save, err = savedata.Clear(uuid, slot, daily.Seed(), s)
	}
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	if save == nil || r.URL.Path == "/api/savedata/update" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = json.NewEncoder(w).Encode(save)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}
}

func handleDailySeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(daily.Seed()))
}

func handleDailyRankings(w http.ResponseWriter, r *http.Request) {
	uuid, err := getUUIDFromRequest(r)
	if err != nil {
		httpError(w, r, err, http.StatusBadRequest)
		return
	}

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

	rankings, err := daily.Rankings(uuid, category, page)
	if err != nil {
		httpError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(rankings)
	if err != nil {
		httpError(w, r, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
		return
	}
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

func httpError(w http.ResponseWriter, r *http.Request, err error, code int) {
	log.Printf("%s: %s\n", r.URL.Path, err)
	http.Error(w, err.Error(), code)
}
