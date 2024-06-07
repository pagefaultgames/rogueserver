package savedata

import (
	"fmt"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

func Session(uuid []byte, slot int) (defs.SessionSaveData, error) {
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
