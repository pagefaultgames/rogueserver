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
	"fmt"
	"os"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

func GetSystem(uuid []byte) (defs.SystemSaveData, error) {
	var system defs.SystemSaveData
	var err error

	if os.Getenv("S3_SYSTEM_BUCKET_NAME") != "" { // use S3
		system, err = db.GetSystemSaveFromS3(uuid)
	} else { // use database
		system, err = db.ReadSystemSaveData(uuid)
	}
	if err != nil {
		return system, err
	}

	return system, nil
}

func UpdateSystem(uuid []byte, data defs.SystemSaveData) error {
	if data.TrainerId == 0 && data.SecretId == 0 {
		return fmt.Errorf("invalid system data")
	}

	err := db.UpdateAccountStats(uuid, data.GameStats, data.VoucherCounts)
	if err != nil {
		return fmt.Errorf("failed to update account stats: %s", err)
	}

	if os.Getenv("S3_SYSTEM_BUCKET_NAME") != "" { // use S3
		err = db.StoreSystemSaveDataS3(uuid, data)
	} else {
		err = db.StoreSystemSaveData(uuid, data)
	}
	if err != nil {
		return err
	}

	return nil
}

func DeleteSystem(uuid []byte) error {
	err := db.DeleteSystemSaveData(uuid)
	if err != nil {
		return err
	}

	return nil
}
