package account

import (
	"crypto/rand"
	"fmt"

	"github.com/pagefaultgames/pokerogue-server/db"
)

func ChangePW(uuid []byte, password string) error {
	if len(password) < 6 {
		return fmt.Errorf("invalid password")
	}

	salt := make([]byte, ArgonSaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to generate salt: %s", err))
	}

	err = db.UpdateAccountPassword(uuid, deriveArgon2IDKey([]byte(password), salt), salt)
	if err != nil {
		return fmt.Errorf("failed to add account record: %s", err)
	}

	return nil
}
