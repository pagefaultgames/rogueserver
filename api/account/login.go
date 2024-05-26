/*
	Copyright (C) 2024  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package account

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/pagefaultgames/rogueserver/db"
)

type LoginResponse GenericAuthResponse

// /account/login - log into account
func Login(username, password string) (LoginResponse, error) {
	var response LoginResponse

	if !isValidUsername(username) {
		return response, fmt.Errorf("invalid username")
	}

	if len(password) < 6 {
		return response, fmt.Errorf("invalid password")
	}

	key, salt, err := db.FetchAccountKeySaltFromUsername(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, fmt.Errorf("account doesn't exist")
		}

		return response, err
	}

	if !bytes.Equal(key, deriveArgon2IDKey([]byte(password), salt)) {
		return response, fmt.Errorf("password doesn't match")
	}

	token := make([]byte, TokenSize)
	_, err = rand.Read(token)
	if err != nil {
		return response, fmt.Errorf("failed to generate token: %s", err)
	}

	err = db.AddAccountSession(username, token)
	if err != nil {
		return response, fmt.Errorf("failed to add account session")
	}

	response.Token = base64.StdEncoding.EncodeToString(token)

	return response, nil
}
