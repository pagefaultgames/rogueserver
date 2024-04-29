// Copyright (C) 2024 Pagefault Games - All Rights Reserved
// https://github.com/pagefaultgames

package account

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/pagefaultgames/rogueserver/db"
)

// /account/register - register account
func Register(username, password string) error {
	if !isValidUsername(username) {
		return fmt.Errorf("invalid username")
	}

	if len(password) < 6 {
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

	err = db.AddAccountRecord(uuid, username, deriveArgon2IDKey([]byte(password), salt), salt)
	if err != nil {
		return fmt.Errorf("failed to add account record: %s", err)
	}

	err = os.MkdirAll(fmt.Sprintf("userdata/%x", uuid), 0755)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf(fmt.Sprintf("failed to create userdata folder: %s", err))
	}

	return nil
}
