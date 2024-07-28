/*
	Copyright (C) 2024  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package account

import (
	"regexp"
	"runtime"

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

	UUIDSize  = 16
	TokenSize = 32
)

var (
	ArgonMaxInstances = runtime.NumCPU()

	isValidUsername = regexp.MustCompile(`^\w{1,16}$`).MatchString
	semaphore       = make(chan bool, ArgonMaxInstances)

	GameURL string
	OAuthCallbackURL string
)

func deriveArgon2IDKey(password, salt []byte) []byte {
	semaphore <- true
	defer func() { <-semaphore }()

	return argon2.IDKey(password, salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeySize)
}
