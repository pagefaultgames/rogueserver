package savedata

import (
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/klauspost/compress/zstd"
	"github.com/pagefaultgames/rogueserver/defs"
)

func readSystemSaveData(uuid []byte) (defs.SystemSaveData, error) {
	var system defs.SystemSaveData

	file, err := os.Open("userdata/" + hex.EncodeToString(uuid) + "/system.pzs")
	if err != nil {
		return system, fmt.Errorf("failed to open save file for reading: %s", err)
	}

	defer file.Close()

	zstdDecoder, err := zstd.NewReader(file)
	if err != nil {
		return system, fmt.Errorf("failed to create zstd decoder: %s", err)
	}

	defer zstdDecoder.Close()

	err = gob.NewDecoder(zstdDecoder).Decode(&system)
	if err != nil {
		return system, fmt.Errorf("failed to deserialize save: %s", err)
	}

	return system, nil
}

func readSessionSaveData(uuid []byte, slotID int) (defs.SessionSaveData, error) {
	var session defs.SessionSaveData

	fileName := "session"
	if slotID != 0 {
		fileName += strconv.Itoa(slotID)
	}

	file, err := os.Open(fmt.Sprintf("userdata/%s/%s.pzs", hex.EncodeToString(uuid), fileName))
	if err != nil {
		return session, fmt.Errorf("failed to open save file for reading: %s", err)
	}

	defer file.Close()

	zstdDecoder, err := zstd.NewReader(file)
	if err != nil {
		return session, fmt.Errorf("failed to create zstd decoder: %s", err)
	}

	defer zstdDecoder.Close()

	err = gob.NewDecoder(zstdDecoder).Decode(&session)
	if err != nil {
		return session, fmt.Errorf("failed to deserialize save: %s", err)
	}

	return session, nil
}

func validateSessionCompleted(session defs.SessionSaveData) bool {
	switch session.GameMode {
	case 0:
		return session.BattleType == 2 && session.WaveIndex == 200
	case 3:
		return session.BattleType == 2 && session.WaveIndex == 50
	}

	return false
}
