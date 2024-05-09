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
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var handle *sql.DB

func Init(username, password, protocol, address, database string) error {
	var err error

	handle, err = sql.Open("mysql", username+":"+password+"@"+protocol+"("+address+")/"+database)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %s", err)
	}

	handle.SetMaxIdleConns(256)
	handle.SetMaxOpenConns(256)
	handle.SetConnMaxIdleTime(time.Second * 30)
	handle.SetConnMaxLifetime(time.Minute)

	tx, err := handle.Begin()
	if err != nil {
		panic(err)
	}
	tx.Exec("CREATE TABLE IF NOT EXISTS accounts (uuid BINARY(16) NOT NULL, username VARCHAR(16) NOT NULL, hash BINARY(32) NOT NULL, salt BINARY(16) NOT NULL, registered TIMESTAMP NOT NULL, lastLoggedIn TIMESTAMP null DEFAULT null, lastActivity TIMESTAMP null DEFAULT null, banned INT(1) NOT NULL DEFAULT 0, trainerId INT(5) unsigned DEFAULT 0, secretId INT(5) unsigned DEFAULT 0, PRIMARY KEY (uuid), UNIQUE KEY (username))")
	tx.Exec("CREATE TABLE IF NOT EXISTS accountStats (uuid BINARY(16) NOT NULL, playTime INT(11) NOT NULL DEFAULT 0, battles INT(11) NOT NULL DEFAULT 0, classicSessionsPlayed INT(11) NOT NULL DEFAULT 0, sessionsWon INT(11) NOT NULL DEFAULT 0, highestEndlessWave INT(11) NOT NULL DEFAULT 0, highestLevel INT(11) NOT NULL DEFAULT 0, pokemonSeen INT(11) NOT NULL DEFAULT 0, pokemonDefeated INT(11) NOT NULL DEFAULT 0, pokemonCaught INT(11) NOT NULL DEFAULT 0, pokemonHatched INT(11) NOT NULL DEFAULT 0, eggsPulled INT(11) NOT NULL DEFAULT 0, regularVouchers INT(11) NOT NULL DEFAULT 0, plusVouchers INT(11) NOT NULL DEFAULT 0, premiumVouchers INT(11) NOT NULL DEFAULT 0, goldenVouchers INT(11) NOT NULL DEFAULT 0, PRIMARY KEY (uuid), KEY uuid (uuid), FOREIGN KEY (uuid) REFERENCES accounts(uuid))")
	tx.Exec("CREATE TABLE IF NOT EXISTS accountCompensations (id INT(11) NOT NULL auto_increment, uuid BINARY(16) NOT NULL, voucherType INT(11) NOT NULL, count INT(11) NOT NULL, claimed BIT(1) NOT NULL DEFAULT b'0', PRIMARY KEY (ID), KEY uuid (uuid), FOREIGN KEY (uuid) REFERENCES accounts(uuid))")
	tx.Exec("CREATE TABLE IF NOT EXISTS dailyRuns (date DATE NOT NULL, seed CHAR(24) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, PRIMARY KEY (date))")
	tx.Exec("CREATE TABLE IF NOT EXISTS accountDailyRuns (uuid BINARY(16) NOT NULL, date DATE NOT NULL, timestamp TIMESTAMP NOT NULL, score INT(11) NOT NULL DEFAULT 0, wave INT(11) NOT NULL, PRIMARY KEY (uuid, date), KEY uuid (uuid), KEY date (date), FOREIGN KEY (uuid) REFERENCES accounts(uuid), FOREIGN KEY (date) REFERENCES dailyRuns(date))")
	tx.Exec("CREATE TABLE IF NOT EXISTS dailyRunCompletions (uuid BINARY(16) NOT NULL, seed CHAR(24) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, mode INT(11) NOT NULL DEFAULT 0, timestamp TIMESTAMP NOT NULL, score INT(11) NOT NULL DEFAULT 0, PRIMARY KEY (uuid, seed), KEY uuid (uuid), FOREIGN KEY (uuid) REFERENCES accounts(uuid))")
	tx.Exec("CREATE TABLE IF NOT EXISTS sessions (token BINARY(32) NOT NULL, uuid BINARY(16) NOT NULL, expire TIMESTAMP NULL DEFAULT NULL, active INT(1) NOT NULL DEFAULT 0, PRIMARY KEY (token), KEY uuid (uuid), FOREIGN KEY (uuid) REFERENCES accounts(uuid))")
	tx.Exec("CREATE TABLE IF NOT EXISTS systemSaveData (uuid BINARY(16) PRIMARY KEY, data LONGBLOB, timestamp TIMESTAMP)")
	tx.Exec("CREATE TABLE IF NOT EXISTS sessionSaveData (uuid BINARY(16), slot TINYINT, data LONGBLOB, timestamp TIMESTAMP, PRIMARY KEY (uuid, slot))")
	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	// TODO temp code
	// create the directory so that it doesn't fail on systems that do not yet have this directory
	os.MkdirAll("userdata", os.ModePerm)
	entries, err := os.ReadDir("userdata")
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		uuidString := entry.Name()
		uuid, err := hex.DecodeString(uuidString)
		if err != nil {
			log.Printf("failed to decode uuid: %s", err)
			continue
		}

		// store new system data
		systemData, err := LegacyReadSystemSaveData(uuid)
		if err != nil {
			log.Printf("failed to read system save data for %v: %s", uuidString, err)
			continue
		}

		err = StoreSystemSaveData(uuid, systemData)
		if err != nil {
			log.Fatalf("failed to store system save data for %v: %s\n", uuidString, err)
			continue
		}

		// delete old system data
		err = os.Remove("userdata/" + uuidString + "/system.pzs")
		if err != nil {
			log.Fatalf("failed to remove legacy system save data for %v: %s", uuidString, err)
		}

		for i := 0; i < 5; i++ {
			sessionData, err := LegacyReadSessionSaveData(uuid, i)
			if err != nil {
				log.Printf("failed to read session save data %v for %v: %s", i, uuidString, err)
				continue
			}

			// store new session data
			err = StoreSessionSaveData(uuid, sessionData, i)
			if err != nil {
				log.Fatalf("failed to store session save data for %v: %s\n", uuidString, err)
			}

			// delete old session data
			filename := "session"
			if i != 0 {
				filename += fmt.Sprintf("%d", i)
			}
			err = os.Remove(fmt.Sprintf("userdata/%s/%s.pzs", uuidString, filename))
			if err != nil {
				log.Fatalf("failed to remove legacy session save data %v for %v: %s", i, uuidString, err)
			}
		}
	}

	return nil
}
