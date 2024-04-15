package account

import (
	"crypto/rand"
	"fmt"

	"github.com/pagefaultgames/pokerogue-server/db"
)

const (
	UUIDSize  = 16
	TokenSize = 32
)

type RegisterRequest GenericAuthRequest

// /account/register - register account
func Register(request RegisterRequest) error {
	if !isValidUsername(request.Username) {
		return fmt.Errorf("invalid username")
	}

	if len(request.Password) < 6 {
		return fmt.Errorf("invalid password")
	}

	uuid := make([]byte, UUIDSize)
	_, err := rand.Read(uuid)
	if err != nil {
		return fmt.Errorf("failed to generate uuid: %s", err)
	}

	salt := make([]byte, ArgonSaltSize)
	_, err = rand.Read(salt)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to generate salt: %s", err))
	}

	err = db.AddAccountRecord(uuid, request.Username, deriveArgon2IDKey([]byte(request.Password), salt), salt)
	if err != nil {
		return fmt.Errorf("failed to add account record: %s", err)
	}

	return nil
}
