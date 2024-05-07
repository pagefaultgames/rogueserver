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

func TryAddDailyRunCompletion(uuid []byte, seed string, mode int) (bool, error) {
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

type DbSystemSaveData struct {
	uuid []byte
	data []byte
}

type DbSessionSaveData struct {
	uuid []byte
	data []byte
}

func ReadSystemSaveData(uuid []byte) (defs.SystemSaveData, error) {
	var data DbSystemSaveData
	err := handle.QueryRow("SELECT uuid, data FROM systemSaveData WHERE uuid = ?", uuid).Scan(&data)

	reader := bytes.NewReader(data.data)
	system := defs.SystemSaveData{}
	err = gob.NewDecoder(reader).Decode(&system)
	return system, err
}

func StoreSystemSaveData(uuid []byte, data defs.SystemSaveData) error {

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return err
	}

	_, err = handle.Exec("INSERT INTO systemSaveData (uuid, data, timestamp) VALUES (?, ?, UTC_TIMESTAMP()) ON DUPLICATE KEY UPDATE data = VALUES(data), timestamp = VALUES(timestamp)", uuid, buf.Bytes())

	return err
}

func ReadSessionSaveData(uuid []byte, slot int) (defs.SessionSaveData, error) {

	var data DbSystemSaveData
	err := handle.QueryRow("SELECT uuid, data FROM sessionSaveData WHERE uuid = ?", uuid).Scan(&data)

	reader := bytes.NewReader(data.data)
	save := defs.SessionSaveData{}
	err = gob.NewDecoder(reader).Decode(&save)

	return save, err
}

func StoreSessionSaveData(uuid []byte, data defs.SessionSaveData, slot int) error {

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return err
	}

	_, err = handle.Exec("INSERT INTO sessionSaveData (uuid, data, slot, timestamp) VALUES (?, ?, ?, UTC_TIMESTAMP()) ON DUPLICATE KEY UPDATE data = VALUES(data), timestamp = VALUES(timestamp)", uuid, buf.Bytes(), slot)

	return err
}
