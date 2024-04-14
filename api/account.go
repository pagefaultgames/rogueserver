package api

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/pagefaultgames/pokerogue-server/db"
)

const (
	UUIDSize  = 16
	TokenSize = 32
)

var isValidUsername = regexp.MustCompile(`^\w{1,16}$`).MatchString

type AccountInfoResponse struct {
	Username        string `json:"username"`
	LastSessionSlot int    `json:"lastSessionSlot"`
}

// /account/info - get account info
func handleAccountInfo(username string, uuid []byte) (AccountInfoResponse, error) {
	var latestSave time.Time
	latestSaveID := -1
	for id := range sessionSlotCount {
		fileName := "session"
		if id != 0 {
			fileName += strconv.Itoa(id)
		}

		stat, err := os.Stat(fmt.Sprintf("userdata/%x/%s.pzs", uuid, fileName))
		if err != nil {
			continue
		}

		if stat.ModTime().After(latestSave) {
			latestSave = stat.ModTime()
			latestSaveID = id
		}
	}

	return AccountInfoResponse{Username: username, LastSessionSlot: latestSaveID}, nil
}

type AccountRegisterRequest GenericAuthRequest

// /account/register - register account
func handleAccountRegister(request AccountRegisterRequest) error {
	if !isValidUsername(request.Username) {
		return fmt.Errorf("invalid username")
	}

	if len(request.Password) < 6 {
		return fmt.Errorf("invalid password")
	}

	uuid := make([]byte, UUIDSize)
	_, err := rand.Read(uuid)
	if err != nil {
		return fmt.Errorf("failed to generate uuid: %s", err)
	}

	salt := make([]byte, ArgonSaltSize)
	_, err = rand.Read(salt)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to generate salt: %s", err))
	}

	err = db.AddAccountRecord(uuid, request.Username, deriveArgon2IDKey([]byte(request.Password), salt), salt)
	if err != nil {
		return fmt.Errorf("failed to add account record: %s", err)
	}

	return nil
}

type AccountLoginRequest GenericAuthRequest
type AccountLoginResponse GenericAuthResponse

// /account/login - log into account
func handleAccountLogin(request AccountLoginRequest) (AccountLoginResponse, error) {
	if !isValidUsername(request.Username) {
		return AccountLoginResponse{}, fmt.Errorf("invalid username")
	}

	if len(request.Password) < 6 {
		return AccountLoginResponse{}, fmt.Errorf("invalid password")
	}

	key, salt, err := db.FetchAccountKeySaltFromUsername(request.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return AccountLoginResponse{}, fmt.Errorf("account doesn't exist")
		}

		return AccountLoginResponse{}, err
	}

	if !bytes.Equal(key, deriveArgon2IDKey([]byte(request.Password), salt)) {
		return AccountLoginResponse{}, fmt.Errorf("password doesn't match")
	}

	token := make([]byte, TokenSize)
	_, err = rand.Read(token)
	if err != nil {
		return AccountLoginResponse{}, fmt.Errorf("failed to generate token: %s", err)
	}

	err = db.AddAccountSession(request.Username, token)
	if err != nil {
		return AccountLoginResponse{}, fmt.Errorf("failed to add account session")
	}

	return AccountLoginResponse{Token: base64.StdEncoding.EncodeToString(token)}, nil
}

// /account/logout - log out of account
func handleAccountLogout(token []byte) error {
	if len(token) != TokenSize {
		return fmt.Errorf("invalid token")
	}

	err := db.RemoveSessionFromToken(token)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("token not found")
		}

		return fmt.Errorf("failed to remove account session")
	}

	return nil
}
