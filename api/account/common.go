package account

import (
	"regexp"

	"golang.org/x/crypto/argon2"
)

type GenericAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GenericAuthResponse struct {
	Token string `json:"token"`
}

const (
	ArgonTime     = 1
	ArgonMemory   = 256 * 1024
	ArgonThreads  = 4
	ArgonKeySize  = 32
	ArgonSaltSize = 16

	UUIDSize  = 16
	TokenSize = 32
)

var isValidUsername = regexp.MustCompile(`^\w{1,16}$`).MatchString

func deriveArgon2IDKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeySize)
}
