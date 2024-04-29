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
	"log"
	"os"
	"strconv"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

// /savedata/delete - delete save data
func Delete(uuid []byte, datatype, slot int) error {
	err := db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	switch datatype {
	case 0: // System
		err := os.Remove(fmt.Sprintf("userdata/%x/system.pzs", uuid))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete save file: %s", err)
		}
	case 1: // Session
		if slot < 0 || slot >= defs.SessionSlotCount {
			return fmt.Errorf("slot id %d out of range", slot)
		}

		fileName := "session"
		if slot != 0 {
			fileName += strconv.Itoa(slot)
		}

		err = os.Remove(fmt.Sprintf("userdata/%x/%s.pzs", uuid, fileName))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete save file: %s", err)
		}
	default:
		return fmt.Errorf("invalid data type")
	}

	return nil
}
