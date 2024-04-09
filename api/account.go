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

	"github.com/Flashfyre/pokerogue-server/db"
	"golang.org/x/crypto/argon2"
)

const (
	UUIDSize      = 16
	ArgonTime     = 1
	ArgonMemory   = 256 * 1024
	ArgonThreads  = 4
	ArgonKeySize  = 32
	ArgonSaltSize = 16
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
func handleAccountRegister(username, password string) error {
	if !isValidUsername(username) {
		return fmt.Errorf("invalid username")
	}

	if len(password) < 6 {
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

	err = db.AddAccountRecord(uuid, username, argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeySize), salt)
	if err != nil {
		return fmt.Errorf("failed to add account record: %s", err)
	}

	return nil
}

type AccountLoginRequest GenericAuthRequest
type AccountLoginResponse GenericAuthResponse

// /account/login - log into account
func handleAccountLogin(username, password string) (AccountLoginResponse, error) {
	if !isValidUsername(username) {
		return AccountLoginResponse{}, fmt.Errorf("invalid username")
	}

	if len(password) < 6 {
		return AccountLoginResponse{}, fmt.Errorf("invalid password")
	}

	key, salt, err := db.FetchAccountKeySaltFromUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return AccountLoginResponse{}, fmt.Errorf("account doesn't exist")
		}

		return AccountLoginResponse{}, err
	}

	if !bytes.Equal(key, argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeySize)) {
		return AccountLoginResponse{}, fmt.Errorf("password doesn't match")
	}

	token := make([]byte, 32)
	_, err = rand.Read(token)
	if err != nil {
		return AccountLoginResponse{}, fmt.Errorf("failed to generate token: %s", err)
	}

	err = db.AddAccountSession(username, token)
	if err != nil {
		return AccountLoginResponse{}, fmt.Errorf("failed to add account session")
	}

	return AccountLoginResponse{Token: base64.StdEncoding.EncodeToString(token)}, nil
}

// /account/logout - log out of account
func handleAccountLogout(token []byte) error {
	if len(token) != 32 {
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
