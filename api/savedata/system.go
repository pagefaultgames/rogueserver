package savedata

import (
	"fmt"
	"strconv"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

func System(uuid []byte) (defs.SystemSaveData, error) {
	system, err := db.ReadSystemSaveData(uuid)
	if err != nil {
		return system, err
	}

	// TODO this should be a transaction
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