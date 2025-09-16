/*
	Copyright (C) 2024 - 2025  Pagefault Games

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
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pagefaultgames/rogueserver/defs"
)

var ErrSaveNotExist = errors.New("save does not exist")

type GetSystemStore interface {
	GetSystemSaveFromS3(uuid []byte) (defs.SystemSaveData, error)
	ReadSystemSaveData(uuid []byte) (defs.SystemSaveData, error)
}

func GetSystem[T GetSystemStore](store T, uuid []byte) (defs.SystemSaveData, error) {
	var system defs.SystemSaveData
	var err error

	if os.Getenv("S3_SYSTEM_BUCKET_NAME") != "" { // use S3
		system, err = store.GetSystemSaveFromS3(uuid)
		var nokey *types.NoSuchKey
		if errors.As(err, &nokey) {
			err = ErrSaveNotExist
		}
	} else { // use database
		system, err = store.ReadSystemSaveData(uuid)
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrSaveNotExist
		}
	}
	if err != nil {
		return system, err
	}

	return system, nil
}

// Interface for database operations needed for updating system data.
type UpdateSystemStore interface {
	UpdateAccountStats(uuid []byte, stats defs.GameStats, voucherCounts map[string]int) error
	StoreSystemSaveDataS3(uuid []byte, data defs.SystemSaveData) error
	StoreSystemSaveData(uuid []byte, data defs.SystemSaveData) error
}

func UpdateSystem[T UpdateSystemStore](store T, uuid []byte, data defs.SystemSaveData) error {
	if data.TrainerId == 0 && data.SecretId == 0 {
		return fmt.Errorf("invalid system data")
	}

	err := store.UpdateAccountStats(uuid, data.GameStats, data.VoucherCounts)
	if err != nil {
		return fmt.Errorf("failed to update account stats: %s", err)
	}

	if os.Getenv("S3_SYSTEM_BUCKET_NAME") != "" { // use S3
		err = store.StoreSystemSaveDataS3(uuid, data)
	} else {
		err = store.StoreSystemSaveData(uuid, data)
	}
	if err != nil {
		return err
	}

	return nil
}

type DeleteSystemStore interface {
	DeleteSystemSaveData(uuid []byte) error
}

func DeleteSystem[T DeleteSystemStore](store T, uuid []byte) error {
	err := store.DeleteSystemSaveData(uuid)
	if err != nil {
		return err
	}

	return nil
}
