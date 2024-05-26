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
		if slot != 0 {
			return nil, fmt.Errorf("invalid slot id for system data")
		}

		system, err := db.ReadSystemSaveData(uuid)
		if err != nil {
			return nil, err
		}

		// TODO this should be a transaction
		compensations, err := db.FetchAndClaimAccountCompensations(uuid)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch compensations: %s", err)
		}

		var needsUpdate bool
		for compensationType, amount := range compensations {
			system.VoucherCounts[strconv.Itoa(compensationType)] += amount
			if amount > 0 {
				needsUpdate = true
			}
		}

		if needsUpdate {
			err = db.StoreSystemSaveData(uuid, system)
			if err != nil {
				return nil, fmt.Errorf("failed to update system save data: %s", err)
			}
			err = db.DeleteClaimedAccountCompensations(uuid)
			if err != nil {
				return nil, fmt.Errorf("failed to delete claimed compensations: %s", err)
			}

			err = db.UpdateAccountStats(uuid, system.GameStats, system.VoucherCounts)
			if err != nil {
				return nil, fmt.Errorf("failed to update account stats: %s", err)
			}
		}

		return system, nil
	case 1: // Session
		if slot < 0 || slot >= defs.SessionSlotCount {
			return nil, fmt.Errorf("slot id %d out of range", slot)
		}

		session, err := db.ReadSessionSaveData(uuid, slot)
		if err != nil {
			return nil, err
		}

		return session, nil
	default:
		return nil, fmt.Errorf("invalid data type")
	}
}
