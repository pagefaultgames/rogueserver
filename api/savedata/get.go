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
	"strconv"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

// /savedata/get - get save data
func Get(uuid []byte, datatype, slot int) (any, error) {
	switch datatype {
	case 0: // System
		system, err := readSystemSaveData(uuid)
		if err != nil {
			return nil, err
		}

		compensations, err := db.FetchAndClaimAccountCompensations(uuid)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch compensations: %s", err)
		}

		for compensationType, amount := range compensations {
			system.VoucherCounts[strconv.Itoa(compensationType)] += amount
		}

		return system, nil
	case 1: // Session
		if slot < 0 || slot >= defs.SessionSlotCount {
			return nil, fmt.Errorf("slot id %d out of range", slot)
		}

		session, err := readSessionSaveData(uuid, slot)
		if err != nil {
			return nil, err
		}

		return session, nil
	default:
		return nil, fmt.Errorf("invalid data type")
	}
}
