package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/pagefaultgames/pokerogue-server/api/account"
	"github.com/pagefaultgames/pokerogue-server/api/daily"
	"github.com/pagefaultgames/pokerogue-server/api/savedata"
	"github.com/pagefaultgames/pokerogue-server/db"
	"github.com/pagefaultgames/pokerogue-server/defs"
)

type Server struct {
	Debug bool
	Exit  *sync.RWMutex
}

/*
	The caller of endpoint handler functions are responsible for extracting the necessary data from the request.
	Handler functions are responsible for checking the validity of this data and returning a result or error.
	Handlers should not return serialized JSON, instead return the struct itself.
*/

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// kind of misusing the RWMutex but it doesn't matter
	s.Exit.RLock()
	defer s.Exit.RUnlock()

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
	case "/account/register":
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
	case "/account/login":
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
	case "/account/logout":
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

	// /game
	case "/game/playercount":
		w.Write([]byte(strconv.Itoa(playerCount)))
	case "/game/titlestats":
		err := json.NewEncoder(w).Encode(defs.TitleStats{
			PlayerCount: playerCount,
			BattleCount: battleCount,
		})
		if err != nil {
			httpError(w, r, fmt.Errorf("failed to encode response json: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
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
		token, err = getTokenFromRequest(r)
		if err != nil {
			httpError(w, r, err, http.StatusBadRequest)
			return
		}

		var active bool
		if r.URL.Path == "/savedata/get" {
			err = db.UpdateActiveSession(uuid, token)
			if err != nil {
				httpError(w, r, fmt.Errorf("failed to update active session: %s", err), http.StatusBadRequest)
				return
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
			save, err = savedata.Clear(uuid, slot, daily.Seed(), s)
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

	// /daily
	case "/daily/seed":
		w.Write([]byte(daily.Seed()))
	case "/daily/rankings":
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

		count, err := daily.RankingPageCount(category)
		if err != nil {
			httpError(w, r, err, http.StatusInternalServerError)
		}

		w.Write([]byte(strconv.Itoa(count)))
	default:
		httpError(w, r, fmt.Errorf("unknown endpoint"), http.StatusNotFound)
		return
	}
}

func httpError(w http.ResponseWriter, r *http.Request, err error, code int) {
	log.Printf("%s: %s\n", r.URL.Path, err)
	http.Error(w, err.Error(), code)
}
