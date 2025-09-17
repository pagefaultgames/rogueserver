//go:build devsetup
// +build devsetup

package db

import (
	"database/sql"
	"fmt"
	"os"
)

// MaybeSetupDb is called by db.go and runs setupDb only in devsetup builds.
func MaybeSetupDb(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	err = setupDb(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func setupDb(tx *sql.Tx) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
		       uuid BINARY(16) NOT NULL PRIMARY KEY,
		       username VARCHAR(16) UNIQUE NOT NULL,
		       hash BINARY(32) NOT NULL,
		       salt BINARY(16) NOT NULL,
		       registered TIMESTAMP NOT NULL,
		       lastLoggedIn TIMESTAMP DEFAULT NULL,
		       lastActivity TIMESTAMP DEFAULT NULL,
		       banned TINYINT(1) NOT NULL DEFAULT 0,
		       trainerId SMALLINT(5) UNSIGNED DEFAULT 0,
		       secretId SMALLINT(5) UNSIGNED DEFAULT 0,
		       discordId VARCHAR(32) UNIQUE DEFAULT NULL,
		       googleId VARCHAR(32) UNIQUE DEFAULT NULL
	       )`,
		`CREATE INDEX IF NOT EXISTS accountsByActivity ON accounts (lastActivity)`,

		`CREATE TABLE IF NOT EXISTS sessions (
		       token BINARY(32) NOT NULL PRIMARY KEY,
		       uuid BINARY(16) NOT NULL,
		       expire TIMESTAMP DEFAULT NULL,
		       CONSTRAINT sessions_ibfk_1 FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE
	       )`,
		`CREATE INDEX IF NOT EXISTS sessionsByUuid ON sessions (uuid)`,

		`CREATE TABLE IF NOT EXISTS accountStats (
		       uuid BINARY(16) NOT NULL PRIMARY KEY,
		       playTime INT(11) NOT NULL DEFAULT 0,
		       battles INT(11) NOT NULL DEFAULT 0,
		       classicSessionsPlayed INT(11) NOT NULL DEFAULT 0,
		       sessionsWon INT(11) NOT NULL DEFAULT 0,
		       highestEndlessWave INT(11) NOT NULL DEFAULT 0,
		       highestLevel INT(11) NOT NULL DEFAULT 0,
		       pokemonSeen INT(11) NOT NULL DEFAULT 0,
		       pokemonDefeated INT(11) NOT NULL DEFAULT 0,
		       pokemonCaught INT(11) NOT NULL DEFAULT 0,
		       pokemonHatched INT(11) NOT NULL DEFAULT 0,
		       eggsPulled INT(11) NOT NULL DEFAULT 0,
		       regularVouchers INT(11) NOT NULL DEFAULT 0,
		       plusVouchers INT(11) NOT NULL DEFAULT 0,
		       premiumVouchers INT(11) NOT NULL DEFAULT 0,
		       goldenVouchers INT(11) NOT NULL DEFAULT 0,
		       CONSTRAINT accountStats_ibfk_1 FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE
	       )`,

		`CREATE TABLE IF NOT EXISTS dailyRuns (
		       date DATE NOT NULL PRIMARY KEY,
		       seed CHAR(24) CHARACTER SET ascii COLLATE ascii_bin NOT NULL
	       )`,
		`CREATE INDEX IF NOT EXISTS dailyRunsByDateAndSeed ON dailyRuns (date, seed)`,

		`CREATE TABLE IF NOT EXISTS dailyRunCompletions (
		       uuid BINARY(16) NOT NULL,
		       seed CHAR(24) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
		       mode INT(11) NOT NULL DEFAULT 0,
		       score INT(11) NOT NULL DEFAULT 0,
		       timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		       PRIMARY KEY (uuid, seed),
		       CONSTRAINT dailyRunCompletions_ibfk_1 FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE
	       )`,
		`CREATE INDEX IF NOT EXISTS dailyRunCompletionsByUuidAndSeed ON dailyRunCompletions (uuid, seed)`,

		`CREATE TABLE IF NOT EXISTS accountDailyRuns (
		       uuid BINARY(16) NOT NULL,
		       date DATE NOT NULL,
		       score INT(11) NOT NULL DEFAULT 0,
		       wave INT(11) NOT NULL DEFAULT 0,
		       timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		       PRIMARY KEY (uuid, date),
		       CONSTRAINT accountDailyRuns_ibfk_1 FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE,
		       CONSTRAINT accountDailyRuns_ibfk_2 FOREIGN KEY (date) REFERENCES dailyRuns (date) ON DELETE NO ACTION ON UPDATE NO ACTION
	       )`,
		`CREATE INDEX IF NOT EXISTS accountDailyRunsByDate ON accountDailyRuns (date)`,

		`CREATE TABLE IF NOT EXISTS sessionSaveData (
		       uuid BINARY(16),
		       slot TINYINT,
		       data LONGBLOB,
		       timestamp TIMESTAMP,
		       PRIMARY KEY (uuid, slot),
		       FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE
	       )`,

		`CREATE TABLE IF NOT EXISTS activeClientSessions (
		       uuid BINARY(16) NOT NULL PRIMARY KEY,
		       clientSessionId VARCHAR(32) NOT NULL,
		       FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE
	       )`,
	}

	// Conditionally add systemSaveData table if AWS_ENDPOINT_URL_S3 is not set
	if os.Getenv("AWS_ENDPOINT_URL_S3") == "" {
		queries = append(queries, `CREATE TABLE IF NOT EXISTS systemSaveData (
		       uuid BINARY(16) PRIMARY KEY,
		       data LONGBLOB,
		       timestamp TIMESTAMP,
		       FOREIGN KEY (uuid) REFERENCES accounts (uuid) ON DELETE CASCADE ON UPDATE CASCADE
	       )`)
	}

	for _, q := range queries {
		_, err := tx.Exec(q)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w, query: %s", err, q)
		}
	}

	return nil
}
