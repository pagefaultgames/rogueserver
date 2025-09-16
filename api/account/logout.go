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
	"database/sql"
	"errors"
	"fmt"
)

// /account/logout - log out of account

// Interface for database operations needed for logout.
type LogoutStore interface {
	RemoveSessionFromToken(token []byte) error
}

func Logout[T LogoutStore](store T, token []byte) error {
	err := store.RemoveSessionFromToken(token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("token not found")
		}

		return fmt.Errorf("failed to remove account session")
	}

	return nil
}
