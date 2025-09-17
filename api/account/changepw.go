/*
	Copyright (C) 2024 - 2025  Pagefault Games

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
)

// Interface for database operations needed for changing password.
type ChangePWStore interface {
	RemoveSessionsFromUUID(uuid []byte) error
	UpdateAccountPassword(uuid []byte, newKey []byte, newSalt []byte) error
}

func ChangePW[T ChangePWStore](store T, uuid []byte, password string) error {
	if len(password) < 6 {
		return fmt.Errorf("invalid password")
	}

	salt := make([]byte, ArgonSaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return fmt.Errorf("failed to generate salt: %s", err)
	}

	err = store.RemoveSessionsFromUUID(uuid)
	if err != nil {
		return fmt.Errorf("failed to remove sessions: %s", err)
	}

	err = store.UpdateAccountPassword(uuid, deriveArgon2IDKey([]byte(password), salt), salt)
	if err != nil {
		return fmt.Errorf("failed to add account record: %s", err)
	}

	return nil
}
