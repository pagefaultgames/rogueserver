package api

import (
	"encoding/gob"
	"net/http"
)

type Server struct {
	Debug bool
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Debug {
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})

	switch r.URL.Path {
	case "/account/info":
		s.HandleAccountInfo(w, r)
	case "/account/register":
		s.HandleAccountRegister(w, r)
	case "/account/login":
		s.HandleAccountLogin(w, r)
	case "/account/logout":
		s.HandleAccountLogout(w, r)

	case "/game/playercount":
		s.HandlePlayerCountGet(w, r)

	case "/savedata/get":
		s.HandleSavedataGet(w, r)
	case "/savedata/update":
		s.HandleSavedataUpdate(w, r)
	case "/savedata/delete":
		s.HandleSavedataDelete(w, r)
	case "/savedata/clear":
		s.HandleSavedataClear(w, r)

	case "/daily/seed":
		s.HandleSeed(w, r)
	case "/daily/rankings":
		s.HandleRankings(w, r)
	case "/daily/rankingpagecount":
		s.HandleRankingPageCount(w, r)
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
