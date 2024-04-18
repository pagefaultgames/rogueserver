package savedata

import (
	"fmt"
	"strconv"

	"github.com/pagefaultgames/pokerogue-server/db"
	"github.com/pagefaultgames/pokerogue-server/defs"
)

// /savedata/get - get save data
func Get(uuid []byte, datatype, slot int) (any, error) {
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

		for compensationType, amount := range compensations {
			system.VoucherCounts[strconv.Itoa(compensationType)] += amount
		}

		return system, nil
	case 1: // Session
		if slot < 0 || slot >= defs.SessionSlotCount {
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
