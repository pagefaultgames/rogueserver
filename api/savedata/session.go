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

package savedata

import (
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

func GetSession(uuid []byte, slot int) (defs.SessionSaveData, error) {
	session, err := db.ReadSessionSaveData(uuid, slot)
	if err != nil {
		return session, err
	}

	return session, nil
}

func UpdateSession(uuid []byte, slot int, data defs.SessionSaveData) error {
	err := db.StoreSessionSaveData(uuid, data, slot)
	if err != nil {
		return err
	}

	return nil
}

func DeleteSession(uuid []byte, slot int) error {
	err := db.DeleteSessionSaveData(uuid, slot)
	if err != nil {
		return err
	}

	return nil
}
