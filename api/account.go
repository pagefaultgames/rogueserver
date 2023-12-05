package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Flashfyre/pokerogue-server/db"
)

// /api/account/info - get account info

type AccountInfoResponse struct{
	Username string `json:"string"`
}

func HandleAccountInfo(w http.ResponseWriter, r *http.Request) {
	token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode token: %s", err), http.StatusBadRequest)
		return
	}

	username, err := db.GetAccountInfoFromToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := json.Marshal(AccountInfoResponse{Username: username})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

// /api/account/register - register account

type AccountRegisterRequest GenericAuthRequest
type AccountRegisterResponse GenericAuthResponse

func HandleAccountRegister(w http.ResponseWriter, r *http.Request) {
	var request AccountRegisterRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request body: %s", err), http.StatusBadRequest)
		return
	}

	
}

// /api/account/login - log into account

type AccountLoginRequest GenericAuthRequest
type AccountLoginResponse GenericAuthResponse

func HandleAccountLogin(w http.ResponseWriter, r *http.Request) {

}

// /api/account/logout - log out of account

func HandleAccountLogout(w http.ResponseWriter, r *http.Request) {

}
