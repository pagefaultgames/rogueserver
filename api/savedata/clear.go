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
	"fmt"
	"log"

	"github.com/pagefaultgames/rogueserver/defs"
)

type ClearResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// Interface for database operations needed for `Clear`
// Helps with testing and reduces coupling.
type ClearStore interface {
	UpdateAccountLastActivity(uuid []byte) error
	TryAddSeedCompletion(uuid []byte, seed string, mode int) (bool, error)
	DeleteSessionSaveData(uuid []byte, slot int) error
	AddOrUpdateAccountDailyRun(uuid []byte, score int, waveCompleted int) error
	SetAccountBanned(uuid []byte, banned bool) error
}

// /savedata/clear - mark session save data as cleared and delete
func Clear[T ClearStore](store T, uuid []byte, slot int, seed string, save defs.SessionSaveData) (ClearResponse, error) {
	var response ClearResponse
	err := store.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	if slot < 0 || slot >= defs.SessionSlotCount {
		return response, fmt.Errorf("slot id %d out of range", slot)
	}

	sessionCompleted := validateSessionCompleted(save)

	if save.GameMode == 3 && save.Seed == seed {
		waveCompleted := save.WaveIndex
		if !sessionCompleted {
			waveCompleted--
		}

		if save.Score >= 20000 {
			store.SetAccountBanned(uuid, true)
		}

		err = store.AddOrUpdateAccountDailyRun(uuid, save.Score, waveCompleted)
		if err != nil {
			log.Printf("failed to add or update daily run record: %s", err)
		}
	}

	if sessionCompleted {
		response.Success, err = store.TryAddSeedCompletion(uuid, save.Seed, int(save.GameMode))
		if err != nil {
			log.Printf("failed to mark seed as completed: %s", err)
		}
	}

	err = store.DeleteSessionSaveData(uuid, slot)
	if err != nil {
		log.Printf("failed to delete session save data: %s", err)
	}

	return response, nil
}
