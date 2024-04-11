package api

import (
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Flashfyre/pokerogue-server/db"
	"github.com/Flashfyre/pokerogue-server/defs"
	"github.com/klauspost/compress/zstd"
)

const sessionSlotCount = 3

// /savedata/get - get save data
func handleSavedataGet(uuid []byte, datatype, slot int) (any, error) {
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

		for k, v := range compensations {
			typeKey := strconv.Itoa(k)
			system.VoucherCounts[typeKey] += v
		}

		return system, nil
	case 1: // Session
		if slot < 0 || slot >= sessionSlotCount {
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

// /savedata/update - update save data
func handleSavedataUpdate(uuid []byte, slot int, save any) error {
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

		err = db.UpdateAccountStats(uuid, save.GameStats)
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
		if slot < 0 || slot >= sessionSlotCount {
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

// /savedata/delete - delete save data
func handleSavedataDelete(uuid []byte, datatype, slot int) error {
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
		if slot < 0 || slot >= sessionSlotCount {
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

type SavedataClearResponse struct {
	Success bool `json:"success"`
}

// /savedata/clear - mark session save data as cleared and delete
func handleSavedataClear(uuid []byte, slot int, save defs.SessionSaveData) (SavedataClearResponse, error) {
	err := db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	if slot < 0 || slot >= sessionSlotCount {
		return SavedataClearResponse{}, fmt.Errorf("slot id %d out of range", slot)
	}

	sessionCompleted := validateSessionCompleted(save)
	newCompletion := false

	if save.GameMode == 3 && save.Seed == dailyRunSeed {
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
		newCompletion, err = db.TryAddSeedCompletion(uuid, save.Seed, int(save.GameMode))
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
		return SavedataClearResponse{}, fmt.Errorf("failed to delete save file: %s", err)
	}

	return SavedataClearResponse{Success: newCompletion}, nil
}
