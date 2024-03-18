package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
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

func UpdateAccountLastActivity(uuid []byte) error {
	_, err := handle.Exec("UPDATE accounts SET lastActivity = UTC_TIMESTAMP() WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}

func FetchUsernameFromToken(token []byte) (string, error) {
	var username string
	err := handle.QueryRow("SELECT a.username FROM accounts a JOIN sessions s ON s.uuid = a.uuid WHERE s.token = ? AND s.expire > UTC_TIMESTAMP()", token).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

func FetchAccountKeySaltFromUsername(username string) ([]byte, []byte, error) {
	var key, salt []byte
	err := handle.QueryRow("SELECT hash, salt FROM accounts WHERE username = ?", username).Scan(&key, &salt)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}

func FetchUuidFromToken(token []byte) ([]byte, error) {
	var uuid []byte
	err := handle.QueryRow("SELECT uuid FROM sessions WHERE token = ? AND expire > UTC_TIMESTAMP()", token).Scan(&uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}

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
