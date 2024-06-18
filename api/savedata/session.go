package savedata

import (
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

func GetSession(uuid []byte, slot int) (defs.SessionSaveData, error) {
	session, err := db.ReadSessionSaveData(uuid, slot)
	if err != nil {
		return session, err
	}

	return session, nil
}

func PutSession(uuid []byte, slot int, data defs.SessionSaveData) error {
	err := db.StoreSessionSaveData(uuid, data, slot)
	if err != nil {
		return err
	}

	return nil
}

func DeleteSession(uuid []byte, slot int) error {
	err := db.DeleteSessionSaveData(uuid, slot)
	if err != nil {
		return err
	}

	return nil
}
