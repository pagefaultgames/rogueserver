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

func AddDiscordIdByUsername(discordId string, username string) error {
	_, err := handle.Exec("UPDATE accounts SET discordId = ? WHERE username = ?", discordId, username)
	if err != nil {
		return err
	}

	return nil
}

func AddGoogleIdByUsername(googleId string, username string) error {
	_, err := handle.Exec("UPDATE accounts SET googleId = ? WHERE username = ?", googleId, username)
	if err != nil {
		return err
	}

	return nil
}

func AddGoogleIdByUUID(googleId string, uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET googleId = ? WHERE uuid = ?", googleId, uuid)
	if err != nil {
		return err
	}

	return nil
}

func AddDiscordIdByUUID(discordId string, uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET discordId = ? WHERE uuid = ?", discordId, uuid)
	if err != nil {
		return err
	}

	return nil
}


func FetchUsernameByDiscordId(discordId string) (string, error) {
	var username string
	err := handle.QueryRow("SELECT username FROM accounts WHERE discordId = ?", discordId).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func FetchUsernameByGoogleId(googleId string) (string, error) {
	var username string
	err := handle.QueryRow("SELECT username FROM accounts WHERE googleId = ?", googleId).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func FetchDiscordIdByUsername(username string) (string, error) {
	var discordId sql.NullString
	err := handle.QueryRow("SELECT discordId FROM accounts WHERE username = ?", username).Scan(&discordId)
	if err != nil {
		return "", err
	}

	if !discordId.Valid {
		return "", nil
	}

	return discordId.String, nil
}

func FetchGoogleIdByUsername(username string) (string, error) {
	var googleId sql.NullString
	err := handle.QueryRow("SELECT googleId FROM accounts WHERE username = ?", username).Scan(&googleId)
	if err != nil {
		return "", err
	}

	if !googleId.Valid {
		return "", nil
	}

	return googleId.String, nil
}

func FetchDiscordIdByUUID(uuid []byte) (string, error) {
	var discordId sql.NullString
	err := handle.QueryRow("SELECT discordId FROM accounts WHERE uuid = ?", uuid).Scan(&discordId)
	if err != nil {
		return "", err
	}

	if !discordId.Valid {
		return "", nil
	}

	return discordId.String, nil
}

func FetchGoogleIdByUUID(uuid []byte) (string, error) {
	var googleId sql.NullString
	err := handle.QueryRow("SELECT googleId FROM accounts WHERE uuid = ?", uuid).Scan(&googleId)
	if err != nil {
		return "", err
	}

	if !googleId.Valid {
		return "", nil
	}

	return googleId.String, nil
}

func FetchUsernameBySessionToken(token []byte) (string, error) {
	var username string
	err := handle.QueryRow("SELECT a.username FROM accounts a JOIN sessions s ON a.uuid = s.uuid WHERE s.token = ?", token).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func CheckUsernameExists(username string) (string, error) {
	var dbUsername sql.NullString
	err := handle.QueryRow("SELECT username FROM accounts WHERE username = ?", username).Scan(&dbUsername)
	if err != nil {
		return "", err
	}
	if !dbUsername.Valid {
		return "", nil
	}

	return dbUsername.String, nil
}

func FetchLastLoggedInDateByUsername(username string) (string, error) {
	var lastLoggedIn sql.NullString
	err := handle.QueryRow("SELECT lastLoggedIn FROM accounts WHERE username = ?", username).Scan(&lastLoggedIn)
	if err != nil {
		return "", err
	}
	if !lastLoggedIn.Valid {
		return "", nil
	}

	return lastLoggedIn.String, nil
}

type AdminSearchResponse struct {
	Username        string `json:"username"`
	DiscordId       string `json:"discordId"`
	GoogleId        string `json:"googleId"`
	LastLoggedIn	string `json:"lastLoggedIn"`
	Registered		string `json:"registered"`
}

func FetchAdminDetailsByUsername(dbUsername string) (AdminSearchResponse, error) {
	var resultUsername, resultDiscordId, resultGoogleId, resultLastLoggedIn, resultRegistered sql.NullString
	var adminResponse AdminSearchResponse

	err := handle.QueryRow("SELECT username, discordId, googleId, lastLoggedIn, registered from accounts WHERE username = ?", dbUsername).Scan(&resultUsername, &resultDiscordId, &resultGoogleId, &resultLastLoggedIn, &resultRegistered)
	if err != nil {
		return adminResponse, err
	}

	adminResponse = AdminSearchResponse{
		Username:        resultUsername.String,
		DiscordId:       resultDiscordId.String,
		GoogleId:        resultGoogleId.String,
		LastLoggedIn:    resultLastLoggedIn.String,
		Registered:		 resultRegistered.String,
	}

	return adminResponse, nil
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

func SetAccountBanned(uuid []byte, banned bool) error {
	_, err := handle.Exec("UPDATE accounts SET banned = ? WHERE uuid = ?", banned, uuid)
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

func IsActiveSession(uuid []byte, sessionId string) (bool, error) {
	var id string
	err := handle.QueryRow("SELECT clientSessionId FROM activeClientSessions WHERE uuid = ?", uuid).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = UpdateActiveSession(uuid, sessionId)
			if err != nil {
				return false, err
			}

			return true, nil
		}

		return false, err
	}

	return id == "" || id == sessionId, nil
}

func UpdateActiveSession(uuid []byte, clientSessionId string) error {
	_, err := handle.Exec("INSERT INTO activeClientSessions (uuid, clientSessionId) VALUES (?, ?) ON DUPLICATE KEY UPDATE clientSessionId = ?", uuid, clientSessionId, clientSessionId)
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

func FetchUUIDFromUsername(username string) ([]byte, error) {
	var uuid []byte
	err := handle.QueryRow("SELECT uuid FROM accounts WHERE username = ?", username).Scan(&uuid)
	if err != nil {
		return nil, err
	}

	return uuid, nil
}

func RemoveDiscordIdByUUID(uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET discordId = NULL WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}

func RemoveGoogleIdByUUID(uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET googleId = NULL WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}

func RemoveGoogleIdByUsername(username string) error {
	_, err := handle.Exec("UPDATE accounts SET googleId = NULL WHERE username = ?", username)
	if err != nil {
		return err
	}

	return nil
}

func RemoveDiscordIdByUsername(username string) error {
	_, err := handle.Exec("UPDATE accounts SET discordId = NULL WHERE username = ?", username)
	if err != nil {
		return err
	}

	return nil
}

func RemoveDiscordIdByDiscordId(discordId string) error {
	_, err := handle.Exec("UPDATE accounts SET discordId = NULL WHERE discordId = ?", discordId)
	if err != nil {
		return err
	}

	return nil
}

func RemoveGoogleIdByDiscordId(discordId string) error {
	_, err := handle.Exec("UPDATE accounts SET googleId = NULL WHERE discordId = ?", discordId)
	if err != nil {
		return err
	}

	return nil
}