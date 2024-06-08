package savedata

import (
	"fmt"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

func GetSession(uuid []byte, slot int) (defs.SessionSaveData, error) {
	var session defs.SessionSaveData

	if slot < 0 || slot >= defs.SessionSlotCount {
		return session, fmt.Errorf("slot id %d out of range", slot)
	}

	var err error
	session, err = db.ReadSessionSaveData(uuid, slot)
	if err != nil {
		return session, err
	}

	return session, nil
}

func PutSession(uuid []byte, slot int, data defs.SessionSaveData) error {
	if slot < 0 || slot >= defs.SessionSlotCount {
		return fmt.Errorf("slot id %d out of range", slot)
	}

	err := db.StoreSessionSaveData(uuid, data, slot)
	if err != nil {
		return err
	}

	return nil
}

func DeleteSession(uuid []byte, slot int) error {
	if slot < 0 || slot >= defs.SessionSlotCount {
		return fmt.Errorf("slot id %d out of range", slot)
	}

	err := db.DeleteSessionSaveData(uuid, slot)
	if err != nil {
		return err
	}

	return nil
}
