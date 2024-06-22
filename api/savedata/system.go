package savedata

import (
	"fmt"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

func GetSystem(uuid []byte) (defs.SystemSaveData, error) {
	system, err := db.ReadSystemSaveData(uuid)
	if err != nil {
		return system, err
	}

	return system, nil
}

func UpdateSystem(uuid []byte, data defs.SystemSaveData) error {
	if data.TrainerId == 0 && data.SecretId == 0 {
		return fmt.Errorf("invalid system data")
	}

	if data.GameVersion != "1.0.4" {
		return fmt.Errorf("client version out of date")
	}

	err := db.UpdateAccountStats(uuid, data.GameStats, data.VoucherCounts)
	if err != nil {
		return fmt.Errorf("failed to update account stats: %s", err)
	}

	return db.StoreSystemSaveData(uuid, data)
}

func DeleteSystem(uuid []byte) error {
	err := db.DeleteSystemSaveData(uuid)
	if err != nil {
		return err
	}

	return nil
}
