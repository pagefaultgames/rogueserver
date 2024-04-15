package account

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/pagefaultgames/pokerogue-server/db"
)

type LoginRequest GenericAuthRequest
type LoginResponse GenericAuthResponse

// /account/login - log into account
func Login(request LoginRequest) (LoginResponse, error) {
	var response LoginResponse

	if !isValidUsername(request.Username) {
		return response, fmt.Errorf("invalid username")
	}

	if len(request.Password) < 6 {
		return response, fmt.Errorf("invalid password")
	}

	key, salt, err := db.FetchAccountKeySaltFromUsername(request.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return response, fmt.Errorf("account doesn't exist")
		}

		return response, err
	}

	if !bytes.Equal(key, deriveArgon2IDKey([]byte(request.Password), salt)) {
		return response, fmt.Errorf("password doesn't match")
	}

	token := make([]byte, TokenSize)
	_, err = rand.Read(token)
	if err != nil {
		return response, fmt.Errorf("failed to generate token: %s", err)
	}

	err = db.AddAccountSession(request.Username, token)
	if err != nil {
		return response, fmt.Errorf("failed to add account session")
	}

	response.Token = base64.StdEncoding.EncodeToString(token)

	return response, nil
}
