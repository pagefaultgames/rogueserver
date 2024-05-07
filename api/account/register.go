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
	"crypto/rand"
	"fmt"
	"github.com/pagefaultgames/rogueserver/db"
)

// /account/register - register account
func Register(username, password string) error {
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

	err = db.AddAccountRecord(uuid, username, deriveArgon2IDKey([]byte(password), salt), salt)
	if err != nil {
		return fmt.Errorf("failed to add account record: %s", err)
	}

	return nil
}
