package account

import (
	"database/sql"
	"fmt"

	"github.com/pagefaultgames/rogueserver/db"
)

// /account/logout - log out of account
func Logout(token []byte) error {
	err := db.RemoveSessionFromToken(token)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("token not found")
		}

		return fmt.Errorf("failed to remove account session")
	}

	return nil
}
