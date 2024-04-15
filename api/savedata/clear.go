package savedata

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/pagefaultgames/pokerogue-server/db"
	"github.com/pagefaultgames/pokerogue-server/defs"
)

type ClearResponse struct {
	Success bool `json:"success"`
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
		err = db.AddOrUpdateAccountDailyRun(uuid, save.Score, waveCompleted)
		if err != nil {
			log.Printf("failed to add or update daily run record: %s", err)
		}
	}

	if sessionCompleted {
		response.Success, err = db.TryAddSeedCompletion(uuid, save.Seed, int(save.GameMode))
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
