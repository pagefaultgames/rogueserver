package api

import "golang.org/x/crypto/argon2"

const (
	ArgonTime     = 1
	ArgonMemory   = 256 * 1024
	ArgonThreads  = 4
	ArgonKeySize  = 32
	ArgonSaltSize = 16
)

func deriveArgon2IDKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeySize)
}
