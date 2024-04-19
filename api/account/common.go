package account

import (
	"regexp"

	"golang.org/x/crypto/argon2"
)

type GenericAuthResponse struct {
	Token string `json:"token"`
}

const (
	ArgonTime     = 1
	ArgonMemory   = 256 * 1024
	ArgonThreads  = 4
	ArgonKeySize  = 32
	ArgonSaltSize = 16

	ArgonMaxInstances = 16

	UUIDSize  = 16
	TokenSize = 32
)

var (
	isValidUsername = regexp.MustCompile(`^\w{1,16}$`).MatchString
	semaphore       = make(chan bool, ArgonMaxInstances)
)

func deriveArgon2IDKey(password, salt []byte) []byte {
	semaphore <- true
	defer func() { <-semaphore }()

	return argon2.IDKey(password, salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeySize)
}
