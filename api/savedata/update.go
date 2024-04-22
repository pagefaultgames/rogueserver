package savedata

import (
	"bytes"
	"encoding/gob"
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

	// ideally should have been done at account creation
	err = os.MkdirAll(fmt.Sprintf("userdata/%x", uuid), 0755)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf(fmt.Sprintf("failed to create userdata folder: %s", err))
	}

	var filename string
	var buf bytes.Buffer

	switch save := save.(type) {
	case defs.SystemSaveData: // System
		if save.TrainerId == 0 && save.SecretId == 0 {
			return fmt.Errorf("invalid system data")
		}

		if save.GameVersion != "1.0.2" {
			return fmt.Errorf("client version out of date")
		}

		err = db.UpdateAccountStats(uuid, save.GameStats, save.VoucherCounts)
		if err != nil {
			return fmt.Errorf("failed to update account stats: %s", err)
		}
		
		filename = "system"

		zstdEncoder, err := zstd.NewWriter(&buf)
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

		filename = "session"
		if slot != 0 {
			filename += strconv.Itoa(slot)
		}

		zstdEncoder, err := zstd.NewWriter(&buf)
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

	err = os.WriteFile(fmt.Sprintf("userdata/%x/%s.pzs", uuid, filename), buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write save to disk: %s", err)
	}

	return nil
}
