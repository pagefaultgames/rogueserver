package account

import (
	"fmt"

	"github.com/pagefaultgames/rogueserver/db"
)

// Interface for database operations needed for changing username.
type ChangeUsernameStore interface {
	FetchDiscordIdByUUID(uuid []byte) (string, error)
	FetchUsernameFromUUID(uuid []byte) (string, error)
	HasChangedUsernameRecently(uuid []byte) (bool, error)
	IsUsernameReserved(uuid []byte, username string) (bool, error)
	UpdateAccountUsername(uuid []byte, oldUsername, newUsername string) error
}

func ChangeUsername[T ChangeUsernameStore](store T, uuid []byte, newUsername string) error {
	if !isValidUsername(newUsername) {
		return fmt.Errorf("invalid username")
	}

	discordId, err := store.FetchDiscordIdByUUID(uuid)
	if err != nil {
		return fmt.Errorf("failed to check discord link: %s", err)
	}
	if discordId == "" {
		return db.ErrNoDiscord
	}

	oldUsername, err := store.FetchUsernameFromUUID(uuid)
	if err != nil {
		return fmt.Errorf("failed to fetch current username: %s", err)
	}
	if oldUsername == newUsername {
		return fmt.Errorf("new username is the same as current username")
	}

	recentlyChanged, err := store.HasChangedUsernameRecently(uuid)
	if err != nil {
		return fmt.Errorf("failed to check username change cooldown: %s", err)
	}
	if recentlyChanged {
		return db.ErrUsernameCooldown
	}

	reserved, err := store.IsUsernameReserved(uuid, newUsername)
	if err != nil {
		return fmt.Errorf("failed to check username availability: %s", err)
	}
	if reserved {
		return db.ErrUsernameReserved
	}

	err = store.UpdateAccountUsername(uuid, oldUsername, newUsername)
	if err != nil {
		return fmt.Errorf("failed to change username: %s", err)
	}

	return nil
}
