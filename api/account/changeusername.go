package account

import (
	"errors"
	"fmt"

	"github.com/pagefaultgames/rogueserver/db"
)

// Interface for database operations needed for changing username.
type ChangeUsernameStore interface {
	UpdateAccountUsername(uuid []byte, newUsername string) error
}

func ChangeUsername[T ChangeUsernameStore](store T, uuid []byte, newUsername string) error {
	if !isValidUsername(newUsername) {
		return fmt.Errorf("invalid username")
	}

	err := store.UpdateAccountUsername(uuid, newUsername)
	if err != nil {
		if errors.Is(err, db.ErrNoDiscord) {
			return err
		}
		return fmt.Errorf("failed to change username: %s", err)
	}

	return nil
}
