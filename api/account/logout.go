package account

import (
	"database/sql"
	"fmt"

	"github.com/pagefaultgames/pokerogue-server/db"
)

// /account/logout - log out of account
func Logout(token []byte) error {
	if len(token) != TokenSize {
		return fmt.Errorf("invalid token")
	}

	err := db.RemoveSessionFromToken(token)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("token not found")
		}

		return fmt.Errorf("failed to remove account session")
	}

	return nil
}
