package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var handle *sql.DB

func Init(username, password, protocol, address, database string) error {
	db, err := sql.Open("mysql", username+":"+password+"@"+protocol+"("+address+")/"+database)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %s", err)
	}

	handle = db

	return nil
}
