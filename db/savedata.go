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
	"errors"
	"time"

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
	systemData, err := ReadSystemSaveData(uuid)
	if err == nil && systemData.Timestamp > data.Timestamp {
		return errors.New("attempted to save an older system save")
	}

	if data.Timestamp > int(time.Now().UTC().Add(time.Hour*48).UnixMilli()) {
		return errors.New("attempted to save a future system save")
	}
	if data.Timestamp < int(time.Now().UTC().Add(-time.Hour*48).UnixMilli()) {
		return errors.New("attempted to save a past system save")
	}

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(data)
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
	session, err := ReadSessionSaveData(uuid, slot)
	if err == nil && session.Seed == data.Seed && session.WaveIndex > data.WaveIndex {
		return errors.New("attempted to save an older session")
	}

	if int64(data.Timestamp) > int64(time.Now().Add(time.Hour*48).UnixMilli()) {
		return errors.New("attempted to save a future session save")
	}
	if int64(data.Timestamp) < int64(time.Now().Add(-time.Hour*48).UnixMilli()) {
		return errors.New("attempted to save a past session save")
	}

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(data)
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
