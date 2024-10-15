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
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var handle *sql.DB

func Init(username, password, protocol, address, database string) error {
	var err error

	handle, err = sql.Open("mysql", username+":"+password+"@"+protocol+"("+address+":3306)/"+database)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %s", err)
	}

	conns := 64

	handle.SetMaxOpenConns(conns)
	handle.SetMaxIdleConns(conns)

	tx, err := handle.Begin()
	if err != nil {
		log.Fatal(err)
	}

	err = setupDb(tx)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func setupDb(tx *sql.Tx) error {
	queries := []string{
		// MIGRATION 000

		`CREATE TABLE IF NOT EXISTS accounts (uuid BINARY(16) NOT NULL PRIMARY KEY, username VARCHAR(16) UNIQUE NOT NULL, hash BINARY(32) NOT NULL, salt BINARY(16) NOT NULL, registered TIMESTAMP NOT NULL, lastLoggedIn TIMESTAMP DEFAULT NULL, lastActivity TIMESTAMP DEFAULT NULL, banned TINYINT(1) NOT NULL DEFAULT 0, trainerId SMALLINT(5) UNSIGNED DEFAULT 0, secretId SMALLINT(5) UNSIGNED DEFAULT 0)`,
		`CREATE INDEX IF NOT EXISTS accountsByActivity ON accounts (lastActivity)`,

		`CREATE TABLE IF NOT EXISTS sessions (token BINARY(32) NOT NULL PRIMARY KEY, uuid BINARY(16) NOT NULL, active TINYINT(1) NOT NULL DEFAULT 0, expire TIMESTAMP DEFAULT NULL, CONSTRAINT sessions_ibfk_1 FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE)`,
		`CREATE INDEX IF NOT EXISTS sessionsByUuid ON sessions (uuid)`,

		`CREATE TABLE IF NOT EXISTS accountStats (uuid BINARY(16) NOT NULL PRIMARY KEY, playTime INT(11) NOT NULL DEFAULT 0, battles INT(11) NOT NULL DEFAULT 0, classicSessionsPlayed INT(11) NOT NULL DEFAULT 0, sessionsWon INT(11) NOT NULL DEFAULT 0, highestEndlessWave INT(11) NOT NULL DEFAULT 0, highestLevel INT(11) NOT NULL DEFAULT 0, pokemonSeen INT(11) NOT NULL DEFAULT 0, pokemonDefeated INT(11) NOT NULL DEFAULT 0, pokemonCaught INT(11) NOT NULL DEFAULT 0, pokemonHatched INT(11) NOT NULL DEFAULT 0, eggsPulled INT(11) NOT NULL DEFAULT 0, regularVouchers INT(11) NOT NULL DEFAULT 0, plusVouchers INT(11) NOT NULL DEFAULT 0, premiumVouchers INT(11) NOT NULL DEFAULT 0, goldenVouchers INT(11) NOT NULL DEFAULT 0, CONSTRAINT accountStats_ibfk_1 FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE)`,

		`CREATE TABLE IF NOT EXISTS accountCompensations (id INT(11) NOT NULL AUTO_INCREMENT PRIMARY KEY, uuid BINARY(16) NOT NULL, voucherType INT(11) NOT NULL, count INT(11) NOT NULL DEFAULT 1, claimed BIT(1) NOT NULL DEFAULT b'0', CONSTRAINT accountCompensations_ibfk_1 FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE)`,
		`CREATE INDEX IF NOT EXISTS accountCompensationsByUuid ON accountCompensations (uuid)`,

		`CREATE TABLE IF NOT EXISTS dailyRuns (date DATE NOT NULL PRIMARY KEY, seed CHAR(24) CHARACTER SET ascii COLLATE ascii_bin NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS dailyRunsByDateAndSeed ON dailyRuns (date, seed)`,

		`CREATE TABLE IF NOT EXISTS dailyRunCompletions (uuid BINARY(16) NOT NULL, seed CHAR(24) CHARACTER SET ascii COLLATE ascii_bin NOT NULL, mode INT(11) NOT NULL DEFAULT 0, score INT(11) NOT NULL DEFAULT 0, timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (uuid, seed), CONSTRAINT dailyRunCompletions_ibfk_1 FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE)`,
		`CREATE INDEX IF NOT EXISTS dailyRunCompletionsByUuidAndSeed ON dailyRunCompletions (uuid, seed)`,

		`CREATE TABLE IF NOT EXISTS accountDailyRuns (uuid BINARY(16) NOT NULL, date DATE NOT NULL, score INT(11) NOT NULL DEFAULT 0, wave INT(11) NOT NULL DEFAULT 0, timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (uuid, date), CONSTRAINT accountDailyRuns_ibfk_1 FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE, CONSTRAINT accountDailyRuns_ibfk_2 FOREIGN KEY (date) REFERENCES dailyRuns (date) ON DELETE NO ACTION ON UPDATE NO ACTION)`,
		`CREATE INDEX IF NOT EXISTS accountDailyRunsByDate ON accountDailyRuns (date)`,

		`CREATE TABLE IF NOT EXISTS systemSaveData (uuid BINARY(16) PRIMARY KEY, data LONGBLOB, timestamp TIMESTAMP, FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE)`,

		`CREATE TABLE IF NOT EXISTS sessionSaveData (uuid BINARY(16), slot TINYINT, data LONGBLOB, timestamp TIMESTAMP, PRIMARY KEY (uuid, slot), FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE)`,

		// ----------------------------------
		// MIGRATION 001

		`ALTER TABLE sessions DROP COLUMN IF EXISTS active`,
		`CREATE TABLE IF NOT EXISTS activeClientSessions (uuid BINARY(16) NOT NULL PRIMARY KEY, clientSessionId VARCHAR(32) NOT NULL, FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE)`,

		// ----------------------------------
		// MIGRATION 002

		`DROP TABLE accountCompensations`,

		// ----------------------------------
		// MIGRATION 003

		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS discordId VARCHAR(32) UNIQUE DEFAULT NULL`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS googleId VARCHAR(32) UNIQUE DEFAULT NULL`,
	}

	for _, q := range queries {
		_, err := tx.Exec(q)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w, query: %s", err, q)
		}
	}
	return nil
}
