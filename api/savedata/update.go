/*
	Copyright (C) 2024  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package savedata

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/klauspost/compress/zstd"
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

var zstdEncoder, _ = zstd.NewWriter(nil)

// /savedata/update - update save data
func Update(uuid []byte, slot int, save any) error {
	err := db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	// ideally should have been done at account creation
	err = os.MkdirAll(fmt.Sprintf("userdata/%x", uuid), 0755)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf(fmt.Sprintf("failed to create userdata folder: %s", err))
	}

	var filename string
	switch save := save.(type) {
	case defs.SystemSaveData: // System
		if save.TrainerId == 0 && save.SecretId == 0 {
			return fmt.Errorf("invalid system data")
		}

		if save.GameVersion != "1.0.4" {
			return fmt.Errorf("client version out of date")
		}

		err = db.UpdateAccountStats(uuid, save.GameStats, save.VoucherCounts)
		if err != nil {
			return fmt.Errorf("failed to update account stats: %s", err)
		}

		filename = "system"

		db.DeleteClaimedAccountCompensations(uuid)
	case defs.SessionSaveData: // Session
		if slot < 0 || slot >= defs.SessionSlotCount {
			return fmt.Errorf("slot id %d out of range", slot)
		}

		filename = "session"
		if slot != 0 {
			filename += strconv.Itoa(slot)
		}
	default:
		return fmt.Errorf("invalid data type")
	}

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(save)
	if err != nil {
		return fmt.Errorf("failed to serialize save: %s", err)
	}

	if buf.Len() == 0 {
		return fmt.Errorf("tried to write empty save file")
	}

	err = os.WriteFile(fmt.Sprintf("userdata/%x/%s.pzs", uuid, filename), zstdEncoder.EncodeAll(buf.Bytes(), nil), 0644)
	if err != nil {
		return fmt.Errorf("failed to write save to disk: %s", err)
	}

	return nil
}
