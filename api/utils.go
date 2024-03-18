package api

import (
	"crypto/rand"
)

const randRunes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const lenRandRunes = len(randRunes)

func RandString(length int) string {
	b := make([]byte, length)

	rand.Read(b)

	for i := range b {
		b[i] = randRunes[int(b[i])%lenRandRunes]
	}

	return string(b)
}
