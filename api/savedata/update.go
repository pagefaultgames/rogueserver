package savedata

import (
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/klauspost/compress/zstd"
	"github.com/pagefaultgames/pokerogue-server/db"
	"github.com/pagefaultgames/pokerogue-server/defs"
)

// /savedata/update - update save data
func Update(uuid []byte, slot int, save any) error {
	err := db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	hexUUID := hex.EncodeToString(uuid)

	switch save := save.(type) {
	case defs.SystemSaveData: // System
		if save.TrainerId == 0 && save.SecretId == 0 {
			return fmt.Errorf("invalid system data")
		}

		err = db.UpdateAccountStats(uuid, save.GameStats, save.VoucherCounts)
		if err != nil {
			return fmt.Errorf("failed to update account stats: %s", err)
		}

		err = os.MkdirAll("userdata/"+hexUUID, 0755)
		if err != nil && !os.IsExist(err) {
			return fmt.Errorf("failed to create userdata folder: %s", err)
		}

		file, err := os.OpenFile("userdata/"+hexUUID+"/system.pzs", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open save file for writing: %s", err)
		}

		defer file.Close()

		zstdEncoder, err := zstd.NewWriter(file)
		if err != nil {
			return fmt.Errorf("failed to create zstd encoder: %s", err)
		}

		defer zstdEncoder.Close()

		err = gob.NewEncoder(zstdEncoder).Encode(save)
		if err != nil {
			return fmt.Errorf("failed to serialize save: %s", err)
		}

		db.DeleteClaimedAccountCompensations(uuid)
	case defs.SessionSaveData: // Session
		if slot < 0 || slot >= defs.SessionSlotCount {
			return fmt.Errorf("slot id %d out of range", slot)
		}

		fileName := "session"
		if slot != 0 {
			fileName += strconv.Itoa(slot)
		}

		err = os.MkdirAll("userdata/"+hexUUID, 0755)
		if err != nil && !os.IsExist(err) {
			return fmt.Errorf(fmt.Sprintf("failed to create userdata folder: %s", err))
		}

		file, err := os.OpenFile(fmt.Sprintf("userdata/%s/%s.pzs", hexUUID, fileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open save file for writing: %s", err)
		}

		defer file.Close()

		zstdEncoder, err := zstd.NewWriter(file)
		if err != nil {
			return fmt.Errorf("failed to create zstd encoder: %s", err)
		}

		defer zstdEncoder.Close()

		err = gob.NewEncoder(zstdEncoder).Encode(save)
		if err != nil {
			return fmt.Errorf("failed to serialize save: %s", err)
		}
	default:
		return fmt.Errorf("invalid data type")
	}

	return nil
}
