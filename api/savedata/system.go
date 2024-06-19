package savedata

import (
	"fmt"
	"strconv"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

func GetSystem(uuid []byte) (defs.SystemSaveData, error) {
	system, err := db.ReadSystemSaveData(uuid)
	if err != nil {
		return system, err
	}

	// TODO: this should be a transaction
	compensations, err := db.FetchAndClaimAccountCompensations(uuid)
	if err != nil {
		return system, fmt.Errorf("failed to fetch compensations: %s", err)
	}

	var needsUpdate bool
	for compensationType, amount := range compensations {
		system.VoucherCounts[strconv.Itoa(compensationType)] += amount
		if amount > 0 {
			needsUpdate = true
		}
	}

	if needsUpdate {
		err = db.StoreSystemSaveData(uuid, system)
		if err != nil {
			return system, fmt.Errorf("failed to update system save data: %s", err)
		}
		err = db.DeleteClaimedAccountCompensations(uuid)
		if err != nil {
			return system, fmt.Errorf("failed to delete claimed compensations: %s", err)
		}

		err = db.UpdateAccountStats(uuid, system.GameStats, system.VoucherCounts)
		if err != nil {
			return system, fmt.Errorf("failed to update account stats: %s", err)
		}
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

	err = db.DeleteClaimedAccountCompensations(uuid)
	if err != nil {
		return fmt.Errorf("failed to delete claimed compensations: %s", err)
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
