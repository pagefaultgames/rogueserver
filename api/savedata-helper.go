package api

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/klauspost/compress/zstd"
)

func GetSystemSaveData(uuid []byte) (SystemSaveData, error) {
	var system SystemSaveData

	save, err := os.ReadFile("userdata/" + hex.EncodeToString(uuid) + "/system.pzs")
	if err != nil {
		return system, fmt.Errorf("failed to read save file: %s", err)
	}

	zstdReader, err := zstd.NewReader(nil)
	if err != nil {
		return system, fmt.Errorf("failed to create zstd reader: %s", err)
	}

	decompressed, err := zstdReader.DecodeAll(save, nil)
	if err != nil {
		return system, fmt.Errorf("failed to decompress save file: %s", err)
	}

	gobDecoderBuf := bytes.NewBuffer(decompressed)

	err = gob.NewDecoder(gobDecoderBuf).Decode(&system)
	if err != nil {
		return system, fmt.Errorf("failed to deserialize save: %s", err)
	}

	return system, nil
}

func GetSessionSaveData(uuid []byte, slotId int) (SessionSaveData, error) {
	var session SessionSaveData

	fileName := "session"
	if slotId != 0 {
		fileName += strconv.Itoa(slotId)
	}

	save, err := os.ReadFile(fmt.Sprintf("userdata/%s/%s.pzs", hex.EncodeToString(uuid), fileName))
	if err != nil {
		return session, fmt.Errorf("failed to read save file: %s", err)
	}

	zstdReader, err := zstd.NewReader(nil)
	if err != nil {
		return session, fmt.Errorf("failed to create zstd reader: %s", err)
	}

	decompressed, err := zstdReader.DecodeAll(save, nil)
	if err != nil {
		return session, fmt.Errorf("failed to decompress save file: %s", err)
	}

	gobDecoderBuf := bytes.NewBuffer(decompressed)

	err = gob.NewDecoder(gobDecoderBuf).Decode(&session)
	if err != nil {
		return session, fmt.Errorf("failed to deserialize save: %s", err)
	}

	return session, nil
}

func ValidateSessionCompleted(session SessionSaveData) bool {
	switch session.GameMode {
	case 0:
		return session.WaveIndex == 200
	case 3:
		return session.WaveIndex == 50
	}
	return false
}
