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
	"fmt"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

// /savedata/newclear - return whether a session is a new clear for its seed
func NewClear(uuid []byte, slot int) (bool, error) {
	if slot < 0 || slot >= defs.SessionSlotCount {
		return false, fmt.Errorf("slot id %d out of range", slot)
	}

	session, err := db.ReadSessionSaveData(uuid, slot)
	if err != nil {
		return false, err
	}

	completed, err := db.ReadSeedCompleted(uuid, session.Seed)
	if err != nil {
		return false, fmt.Errorf("failed to read seed completed: %s", err)
	}

	return !completed, nil
}
