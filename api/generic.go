package api

import "net/http"

type Server struct {
	Debug bool
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Debug {
		w.Header().Add("Access-Control-Allow-Origin", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	switch r.URL.Path {
	case "/account/info":
		s.HandleAccountInfo(w, r)
	case "/account/register":
		s.HandleAccountRegister(w, r)
	case "/account/login":
		s.HandleAccountLogin(w, r)
	case "/account/logout":
		s.HandleAccountLogout(w, r)

	case "/savedata/get":
		s.HandleSavedataGet(w, r)
	case "/savedata/update":
		s.HandleSavedataUpdate(w, r)
	case "/savedata/delete":
		s.HandleSavedataDelete(w, r)
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
