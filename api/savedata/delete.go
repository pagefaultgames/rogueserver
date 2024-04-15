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

// /savedata/delete - delete save data
func Delete(uuid []byte, datatype, slot int) error {
	err := db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	hexUUID := hex.EncodeToString(uuid)

	switch datatype {
	case 0: // System
		err := os.Remove("userdata/" + hexUUID + "/system.pzs")
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

		err = os.Remove(fmt.Sprintf("userdata/%s/%s.pzs", hexUUID, fileName))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete save file: %s", err)
		}
	default:
		return fmt.Errorf("invalid data type")
	}

	return nil
}
