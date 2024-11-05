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
	"time"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

// /savedata/update - update save data
func Update(uuid []byte, slot int, save any) error {

	err := db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	username := db.FetchUsernameFromUUID(uuid)
	switch save := save.(type) {
	case defs.SystemSaveData: // System
		if save.TrainerId == 0 && save.SecretId == 0 {
			return fmt.Errorf("invalid system data")
		}
		ProcessSystemMetrics(save, uuid);
		err = db.UpdateAccountStats(uuid, save.GameStats, save.VoucherCounts)
		if err != nil {
			return fmt.Errorf("failed to update account stats: %s", err)
		}

		return db.StoreSystemSaveData(uuid, save)

	case defs.SessionSaveData: // Session
		if slot < 0 || slot >= defs.SessionSlotCount {
			return fmt.Errorf("slot id %d out of range", slot)
		}
		ProcessSessionMetrics(save, uuid)
		return db.StoreSessionSaveData(uuid, save, slot)

	default:
		return fmt.Errorf("invalid data type")
	}
}

func ProcessSystemMetrics(save defs.SystemSaveData, username string) {

}

func ProcessSessionMetrics(save defs.SessionSaveData, username string) {
	err := Cache.Add(fmt.Sprintf("session-%s-%d", username, save.GameMode), uuid, time.Minute*5)
	if err != nil {
		log.Printf("already cached game mode for %s", username)
		return
	} else {
		log.Printf("increased game mode counter for %s", username)
		switch save.GameMode {
		case 0:
			gameModeCounter.WithLabelValues("classic").Inc()
		case 1:
			gameModeCounter.WithLabelValues("endless").Inc()
		case 2:
			gameModeCounter.WithLabelValues("spliced-endless").Inc()
		case 3:
			gameModeCounter.WithLabelValues("daily").Inc()
		case 4:
			gameModeCounter.WithLabelValues("challenge").Inc()
		}
	}
	
	if save.WaveIndex == 1 {
		for i := 0; i < len(save.Party); i++ {
			formIndex = ""
			if save.Party[i]["formIndex"] != 0 {
				formIndex = "-"+save.Party[i]["formIndex"]
			}
			key = save.Party[i]["Species"] + formIndex
			log.Printf("incremented starter %s", key)
			starterCounter.WithLabelValues(key).Inc()
		}
	}
}
