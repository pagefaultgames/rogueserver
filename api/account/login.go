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
	if !isValidUsername(request.Username) {
		return LoginResponse{}, fmt.Errorf("invalid username")
	}

	if len(request.Password) < 6 {
		return LoginResponse{}, fmt.Errorf("invalid password")
	}

	key, salt, err := db.FetchAccountKeySaltFromUsername(request.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return LoginResponse{}, fmt.Errorf("account doesn't exist")
		}

		return LoginResponse{}, err
	}

	if !bytes.Equal(key, deriveArgon2IDKey([]byte(request.Password), salt)) {
		return LoginResponse{}, fmt.Errorf("password doesn't match")
	}

	token := make([]byte, TokenSize)
	_, err = rand.Read(token)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to generate token: %s", err)
	}

	err = db.AddAccountSession(request.Username, token)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to add account session")
	}

	return LoginResponse{Token: base64.StdEncoding.EncodeToString(token)}, nil
}
