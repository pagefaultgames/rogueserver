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

package cache

import (
	"fmt"
	"time"
)

func TryAddSeedCompletion(uuid []byte, seed string, mode int) bool {
	key := fmt.Sprintf("savedata:%s", uuid)
	err := rdb.HMSet(key, map[string]interface{}{
		"mode":      mode,
		"seed":      seed,
		"timestamp": time.Now().Unix(),
	}).Err()
	return err == nil
}

func ReadSeedCompletion(uuid []byte, seed string) (bool, bool) {
	key := fmt.Sprintf("savedata:%s", uuid)
	completed, err := rdb.HExists(key, seed).Result()
	return completed, err == nil
}
