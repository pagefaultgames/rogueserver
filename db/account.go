package db

import (
	"database/sql"
	"fmt"
)

func GetAccountInfoFromToken(token []byte) (string, error) {
	var username string
	err := handle.QueryRow("SELECT username FROM accounts WHERE uuid IN (SELECT uuid FROM sessions WHERE token = ? AND expire > UTC_TIMESTAMP())").Scan(&username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("invalid token")
		}

		return "", fmt.Errorf("query failed: %s", err)
	}

	return username, nil
}
