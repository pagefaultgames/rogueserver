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

package savedata

import (
	"database/sql"
	"errors"

	"github.com/pagefaultgames/rogueserver/defs"
)

type GetSessionStore interface {
	ReadSessionSaveData(uuid []byte, slot int) (defs.SessionSaveData, error)
}

func GetSession[T GetSessionStore](store T, uuid []byte, slot int) (defs.SessionSaveData, error) {
	session, err := store.ReadSessionSaveData(uuid, slot)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrSaveNotExist
		}

		return session, err
	}

	return session, nil
}

type UpdateSessionStore interface {
	StoreSessionSaveData(uuid []byte, data defs.SessionSaveData, slot int) error
	DeleteSessionSaveData(uuid []byte, slot int) error
}

func UpdateSession[T UpdateSessionStore](store T, uuid []byte, slot int, data defs.SessionSaveData) error {
	err := store.StoreSessionSaveData(uuid, data, slot)
	if err != nil {
		return err
	}

	return nil
}

type DeleteSessionStore interface {
	DeleteSessionSaveData(uuid []byte, slot int) error
}

func DeleteSession[T DeleteSessionStore](store T, uuid []byte, slot int) error {
	err := store.DeleteSessionSaveData(uuid, slot)
	if err != nil {
		return err
	}

	return nil
}
