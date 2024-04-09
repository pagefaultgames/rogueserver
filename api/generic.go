package api

import (
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Flashfyre/pokerogue-server/defs"
)

type Server struct {
	Debug bool
}

/*
	The caller of endpoint handler functions are responsible for extracting the necessary data from the request.
	Handler functions are responsible for checking the validity of this data and returning a result or error.
	Handlers should not return serialized JSON, instead return the struct itself.
*/

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})

	if s.Debug {
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	switch r.URL.Path {
	// /account
	case "/account/info":
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

		info, err := handleAccountInfo(username, uuid)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(info)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to marshal response json: %s", err), http.StatusInternalServerError)
			return
		}

		w.Write(response)
	case "/account/register":
		var request AccountRegisterRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
			return
		}

		err = handleAccountRegister(request.Username, request.Password)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	case "/account/login":
		var request AccountLoginRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to decode request body: %s", err), http.StatusBadRequest)
			return
		}

		token, err := handleAccountLogin(request.Username, request.Password)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(token)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to marshal response json: %s", err), http.StatusInternalServerError)
			return
		}

		w.Write(response)
	case "/account/logout":
		token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to decode token: %s", err), http.StatusBadRequest)
			return
		}

		err = handleAccountLogout(token)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

		// /game
	case "/game/playercount":
		w.Write([]byte(strconv.Itoa(playerCount)))
	case "/game/titlestats":
		response, err := json.Marshal(&defs.TitleStats{
			PlayerCount: playerCount,
			BattleCount: battleCount,
		})
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to marshal response json: %s", err), http.StatusInternalServerError)
			return
		}

		w.Write(response)
	case "/game/classicsessioncount":
		w.Write([]byte(strconv.Itoa(classicSessionCount)))

		// /savedata
	case "/savedata/get", "/savedata/update", "/savedata/delete", "/savedata/clear":
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
		// /savedata/delete specifies datatype, but doesn't expect data in body
		if r.URL.Path != "/savedata/delete" {
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

		switch r.URL.Path {
		case "/savedata/get":
			save, err = handleSavedataGet(uuid, datatype, slot)
		case "/savedata/update":
			err = handleSavedataUpdate(uuid, slot, save)
		case "/savedata/delete":
			err = handleSavedataDelete(uuid, datatype, slot)
		case "/savedata/clear":
			// doesn't return a save, but it works
			save, err = handleSavedataClear(uuid, slot, save.(defs.SessionSaveData))
		}
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		if save == nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		response, err := json.Marshal(save)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to marshal response json: %s", err), http.StatusInternalServerError)
			return
		}

		w.Write(response)

		// /daily
	case "/daily/seed":
		w.Write([]byte(dailyRunSeed))
	case "/daily/rankings":
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

		rankings, err := handleRankings(uuid, category, page)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(rankings)
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to marshal response json: %s", err), http.StatusInternalServerError)
			return
		}

		w.Write(response)
	case "/daily/rankingpagecount":
		var category int
		if r.URL.Query().Has("category") {
			var err error
			category, err = strconv.Atoi(r.URL.Query().Get("category"))
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to convert category: %s", err), http.StatusBadRequest)
				return
			}
		}

		count, err := handleRankingPageCount(category)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
		}

		w.Write([]byte(strconv.Itoa(count)))
	}
}

// auth

type GenericAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GenericAuthResponse struct {
	Token string `json:"token"`
}
