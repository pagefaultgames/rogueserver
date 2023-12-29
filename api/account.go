package api

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/Flashfyre/pokerogue-server/db"
	"golang.org/x/crypto/argon2"
)

const (
	argonTime = 1
	argonMemory = 256*1024
	argonThreads = 4
	argonKeyLength = 32
)

var isValidUsername = regexp.MustCompile(`^\w{1,16}$`).MatchString

// /api/account/info - get account info

type AccountInfoResponse struct{
	Username string `json:"string"`
}

func HandleAccountInfo(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode token: %s", err), http.StatusBadRequest)
		return
	}

	username, err := db.GetUsernameFromToken(token)
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

func HandleAccountRegister(w http.ResponseWriter, r *http.Request) {
	var request AccountRegisterRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request body: %s", err), http.StatusBadRequest)
		return
	}

	if !isValidUsername(request.Username) {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}

	if len(request.Password) < 6 {
		http.Error(w, "invalid password", http.StatusBadRequest)
		return
	}

	uuid := make([]byte, 16)

	_, err = rand.Read(uuid)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate uuid: %s", err), http.StatusInternalServerError)
		return
	}

	salt := make([]byte, 16)

	_, err = rand.Read(salt)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate salt: %s", err), http.StatusInternalServerError)
		return
	}

	err = db.AddAccountRecord(uuid, request.Username, argon2.IDKey([]byte(request.Password), salt, argonTime, argonMemory, argonThreads, argonKeyLength), salt)
	if err != nil {
		http.Error(w, "failed to add account record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// /api/account/login - log into account

type AccountLoginRequest GenericAuthRequest
type AccountLoginResponse GenericAuthResponse

func HandleAccountLogin(w http.ResponseWriter, r *http.Request) {
	var request AccountLoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request body: %s", err), http.StatusBadRequest)
		return
	}

	if !isValidUsername(request.Username) {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}

	if len(request.Password) < 6 {
		http.Error(w, "invalid password", http.StatusBadRequest)
		return
	}

	key, salt, err := db.GetAccountKeySaltFromUsername(request.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "account doesn't exist", http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !bytes.Equal(key, argon2.IDKey([]byte(request.Password), salt, argonTime, argonMemory, argonThreads, argonKeyLength)) {
		http.Error(w, "password doesn't match", http.StatusBadRequest)
		return
	}

	token := make([]byte, 16)

	_, err = rand.Read(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate token: %s", err), http.StatusInternalServerError)
		return
	}

	err = db.AddAccountSession(request.Username, token)
	if err != nil {
		http.Error(w, "failed to add account session", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(AccountLoginResponse{Token: base64.StdEncoding.EncodeToString(token)})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

// /api/account/logout - log out of account

func HandleAccountLogout(w http.ResponseWriter, r *http.Request) {
	token, err := base64.StdEncoding.DecodeString(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode token: %s", err), http.StatusBadRequest)
		return
	}

	if len(token) != 32 {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	err = db.RemoveSessionFromToken(token)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "token not found", http.StatusBadRequest)
			return
		}

		http.Error(w, "failed to remove account session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
