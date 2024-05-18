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
	"database/sql"
	"errors"
	"fmt"
	"slices"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pagefaultgames/rogueserver/defs"
)

func AddAccountRecord(uuid []byte, username string, key, salt []byte) error {
	_, err := handle.Exec("INSERT INTO accounts (uuid, username, hash, salt, registered) VALUES (?, ?, ?, ?, UTC_TIMESTAMP())", uuid, username, key, salt)
	if err != nil {
		return err
	}

	return nil
}

func AddAccountSession(username string, token []byte) error {
	_, err := handle.Exec("INSERT INTO sessions (uuid, token, expire) SELECT a.uuid, ?, DATE_ADD(UTC_TIMESTAMP(), INTERVAL 1 WEEK) FROM accounts a WHERE a.username = ?", token, username)
	if err != nil {
		return err
	}

	_, err = handle.Exec("UPDATE accounts SET lastLoggedIn = UTC_TIMESTAMP() WHERE username = ?", username)
	if err != nil {
		return err
	}

	return nil
}

func UpdateAccountPassword(uuid, key, salt []byte) error {
	_, err := handle.Exec("UPDATE accounts SET (hash, salt) VALUES (?, ?) WHERE uuid = ?", key, salt, uuid)
	if err != nil {
		return err
	}

	return nil
}

func UpdateAccountLastActivity(uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET lastActivity = UTC_TIMESTAMP() WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}

func UpdateAccountStats(uuid []byte, stats defs.GameStats, voucherCounts map[string]int) error {
	var columns = []string{"playTime", "battles", "classicSessionsPlayed", "sessionsWon", "highestEndlessWave", "highestLevel", "pokemonSeen", "pokemonDefeated", "pokemonCaught", "pokemonHatched", "eggsPulled", "regularVouchers", "plusVouchers", "premiumVouchers", "goldenVouchers"}

	var statCols []string
	var statValues []interface{}

	m, ok := stats.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, got %T", stats)
	}

	for k, v := range m {
		value, ok := v.(float64)
		if !ok {
			return fmt.Errorf("expected float64, got %T", v)
		}

		if slices.Contains(columns, k) {
			statCols = append(statCols, k)
			statValues = append(statValues, value)
		}
	}

	for k, v := range voucherCounts {
		var column string
		switch k {
		case "0":
			column = "regularVouchers"
		case "1":
			column = "plusVouchers"
		case "2":
			column = "premiumVouchers"
		case "3":
			column = "goldenVouchers"
		default:
			continue
		}
		statCols = append(statCols, column)
		statValues = append(statValues, v)
	}

	var statArgs []interface{}
	statArgs = append(statArgs, uuid)
	for range 2 {
		statArgs = append(statArgs, statValues...)
	}

	query := "INSERT INTO accountStats (uuid"

	for _, col := range statCols {
		query += ", " + col
	}

	query += ") VALUES (?"

	for range len(statCols) {
		query += ", ?"
	}

	query += ") ON DUPLICATE KEY UPDATE "

	for i, col := range statCols {
		if i > 0 {
			query += ", "
		}

		query += col + " = ?"
	}

	_, err := handle.Exec(query, statArgs...)
	if err != nil {
		return err
	}

	return nil
}

func FetchAndClaimAccountCompensations(uuid []byte) (map[int]int, error) {
	var compensations = make(map[int]int)

	results, err := handle.Query("SELECT voucherType, count FROM accountCompensations WHERE uuid = ?", uuid)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	for results.Next() {
		var voucherType int
		var count int
		err := results.Scan(&voucherType, &count)
		if err != nil {
			return compensations, err
		}
		compensations[voucherType] = count
	}

	_, err = handle.Exec("UPDATE accountCompensations SET claimed = 1 WHERE uuid = ?", uuid)
	if err != nil {
		return compensations, err
	}

	return compensations, nil
}

func DeleteClaimedAccountCompensations(uuid []byte) error {
	_, err := handle.Exec("DELETE FROM accountCompensations WHERE uuid = ? AND claimed = 1", uuid)
	if err != nil {
		return err
	}

	return nil
}

func FetchAccountKeySaltFromUsername(username string) ([]byte, []byte, error) {
	var key, salt []byte
	err := handle.QueryRow("SELECT hash, salt FROM accounts WHERE username = ?", username).Scan(&key, &salt)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}

func FetchTrainerIds(uuid []byte) (trainerId, secretId int, err error) {
	err = handle.QueryRow("SELECT trainerId, secretId FROM accounts WHERE uuid = ?", uuid).Scan(&trainerId, &secretId)
	if err != nil {
		return 0, 0, err
	}

	return trainerId, secretId, nil
}

func UpdateTrainerIds(trainerId, secretId int, uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET trainerId = ?, secretId = ? WHERE uuid = ?", trainerId, secretId, uuid)
	if err != nil {
		return err
	}

	return nil
}

func IsActiveSession(uuid []byte, clientSessionId string) (bool, error) {
	var storedId string
	err := handle.QueryRow("SELECT clientSessionId FROM activeClientSessions WHERE uuid = ?", uuid).Scan(&storedId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = UpdateActiveSession(uuid, clientSessionId)
			if err != nil {
				return false, err
			}
			return true, nil
		}
		return false, err
	}

	return storedId == "" || storedId == clientSessionId, nil
}

func UpdateActiveSession(uuid []byte, clientSessionId string) error {
	_, err := handle.Exec("REPLACE INTO activeClientSessions VALUES (?, ?)", uuid, clientSessionId)
	if err != nil {
		return err
	}

	return nil
}

func FetchUUIDFromToken(token []byte) ([]byte, error) {
	var uuid []byte
	err := handle.QueryRow("SELECT uuid FROM sessions WHERE token = ?", token).Scan(&uuid)
	if err != nil {
		return nil, err
	}

	return uuid, nil
}

func RemoveSessionFromToken(token []byte) error {
	_, err := handle.Exec("DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		return err
	}

	return nil
}

func FetchUsernameFromUUID(uuid []byte) (string, error) {
	var username string
	err := handle.QueryRow("SELECT username FROM accounts WHERE uuid = ?", uuid).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func isFriendWith(sourceUsername string, friendUsername string) (bool, error) {
	var result int
	err := handle.QueryRow("SELECT COUNT(*) FROM friends WHERE user = ? AND friend = ?", sourceUsername, friendUsername).Scan(&result)
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func AddFriend(uuid []byte, friendUsername string) (bool, error) {
	username, err := FetchUsernameFromUUID(uuid);
	if err != nil {
		return false, err
	}

	var doesUserExist int
	err = handle.QueryRow("SELECT COUNT(*) FROM accounts WHERE username = ?", friendUsername).Scan(&doesUserExist)
	if err != nil {
		return false, err
	}

	if doesUserExist == 0 {
		return false, fmt.Errorf("User does not exist")
	}

	alreadyFriends, _ := isFriendWith(username, friendUsername)
	if alreadyFriends {
		return false, fmt.Errorf("Already friend with this user")
	}

	_, err = handle.Exec("INSERT INTO friends (user, friend, since) VALUES (?, ?, UTC_TIMESTAMP())", username, friendUsername)
	if err != nil {
		return false, err
	}

	return true, nil
}