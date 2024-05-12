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

package db

import (
	"bytes"
	"encoding/gob"

	"github.com/pagefaultgames/rogueserver/defs"
)

func TryAddSeedCompletion(uuid []byte, seed string, mode int) (bool, error) {
	var count int
	err := handle.QueryRow("SELECT COUNT(*) FROM dailyRunCompletions WHERE uuid = ? AND seed = ?", uuid, seed).Scan(&count)
	if err != nil {
		return false, err
	} else if count > 0 {
		return false, nil
	}

	_, err = handle.Exec("INSERT INTO dailyRunCompletions (uuid, seed, mode, timestamp) VALUES (?, ?, ?, UTC_TIMESTAMP())", uuid, seed, mode)
	if err != nil {
		return false, err
	}

	return true, nil
}

func ReadSeedCompleted(uuid []byte, seed string) (bool, error) {
	var count int
	err := handle.QueryRow("SELECT COUNT(*) FROM dailyRunCompletions WHERE uuid = ? AND seed = ?", uuid, seed).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func ReadSystemSaveData(uuid []byte) (defs.SystemSaveData, error) {
	var system defs.SystemSaveData

	var data []byte
	err := handle.QueryRow("SELECT data FROM systemSaveData WHERE uuid = ?", uuid).Scan(&data)
	if err != nil {
		return system, err
	}

	err = gob.NewDecoder(bytes.NewReader(data)).Decode(&system)
	if err != nil {
		return system, err
	}

	return system, nil
}

func StoreSystemSaveData(uuid []byte, data defs.SystemSaveData) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return err
	}

	_, err = handle.Exec("INSERT INTO systemSaveData (uuid, data, timestamp) VALUES (?, ?, UTC_TIMESTAMP()) ON DUPLICATE KEY UPDATE data = ?, timestamp = UTC_TIMESTAMP()", uuid, buf.Bytes(), buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func DeleteSystemSaveData(uuid []byte) error {
	_, err := handle.Exec("DELETE FROM systemSaveData WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}

func ReadSessionSaveData(uuid []byte, slot int) (defs.SessionSaveData, error) {
	var session defs.SessionSaveData

	var data []byte
	err := handle.QueryRow("SELECT data FROM sessionSaveData WHERE uuid = ? AND slot = ?", uuid, slot).Scan(&data)
	if err != nil {
		return session, err
	}

	err = gob.NewDecoder(bytes.NewReader(data)).Decode(&session)
	if err != nil {
		return session, err
	}

	return session, nil
}

func GetLatestSessionSaveDataSlot(uuid []byte) (int, error) {
	var slot int
	err := handle.QueryRow("SELECT slot FROM sessionSaveData WHERE uuid = ? ORDER BY timestamp DESC, slot ASC LIMIT 1", uuid).Scan(&slot)
	if err != nil {
		return -1, err
	}

	return slot, nil
}

func StoreSessionSaveData(uuid []byte, data defs.SessionSaveData, slot int) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return err
	}

	_, err = handle.Exec("INSERT INTO sessionSaveData (uuid, slot, data, timestamp) VALUES (?, ?, ?, UTC_TIMESTAMP()) ON DUPLICATE KEY UPDATE data = ?, timestamp = UTC_TIMESTAMP()", uuid, slot, buf.Bytes(), buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func DeleteSessionSaveData(uuid []byte, slot int) error {
	_, err := handle.Exec("DELETE FROM sessionSaveData WHERE uuid = ? AND slot = ?", uuid, slot)
	if err != nil {
		return err
	}

	return nil
}

func RetrieveAccountEggs(uuid []byte) ([]defs.EggData, error) {
	var accountEggs []defs.EggData

	rows, err := handle.Query("SELECT uuid, gachaType, hatchWaves, timestamp FROM eggs WHERE owner = ?", uuid)
	if err != nil {
		return accountEggs, err
	}

	// For each row, we parse the raw data into an EggData and add it to the result
	for rows.Next() {
		var egg defs.EggData
		err = rows.Scan(&egg.Id, &egg.GachaType, &egg.HatchWaves, &egg.Timestamp)
		if err != nil {
			return accountEggs, err
		}

		accountEggs = append(accountEggs, egg)
	}

	return accountEggs, nil
}

func UpdateAccountEggs(uuid []byte, eggs []defs.EggData) error {
	for _, egg := range eggs {
		// TODO: find a fix to enforce encoding from body to EggData only if
		// it respects the EggData struct so we can get rid of the test 
		if egg.Id == 0 {
			continue
		}

		_, err = handle.Exec(`INSERT INTO eggs (uuid, owner, gachaType, hatchWaves, timestamp) 
							  VALUES (?, ?, ?, ?, ?) 
							  ON DUPLICATE KEY UPDATE hatchWaves = ?`, 
							  egg.Id, uuid, egg.GachaType, egg.HatchWaves, egg.Timestamp, egg.HatchWaves)
		if err != nil {
			return err
		}
	}

	return nil
}

func RemoveAccountEgg(uuid []byte, eggId int) error {
	_, err := handle.Exec("DELETE FROM eggs WHERE owner = ? AND uuid = ?", uuid, eggId)
	if err != nil {
		return err
	}

	return nil
}
