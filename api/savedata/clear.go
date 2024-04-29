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
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

type ClearResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// /savedata/clear - mark session save data as cleared and delete
func Clear(uuid []byte, slot int, seed string, save defs.SessionSaveData) (ClearResponse, error) {
	var response ClearResponse
	err := db.UpdateAccountLastActivity(uuid)
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
			db.UpdateAccountBanned(uuid, true)
		}

		err = db.AddOrUpdateAccountDailyRun(uuid, save.Score, waveCompleted)
		if err != nil {
			log.Printf("failed to add or update daily run record: %s", err)
		}
	}

	if sessionCompleted {
		response.Success, err = db.TryAddDailyRunCompletion(uuid, save.Seed, int(save.GameMode))
		if err != nil {
			log.Printf("failed to mark seed as completed: %s", err)
		}
	}

	fileName := "session"
	if slot != 0 {
		fileName += strconv.Itoa(slot)
	}

	err = os.Remove(fmt.Sprintf("userdata/%s/%s.pzs", hex.EncodeToString(uuid), fileName))
	if err != nil && !os.IsNotExist(err) {
		return response, fmt.Errorf("failed to delete save file: %s", err)
	}

	return response, nil
}
