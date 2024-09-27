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

func TryAddDailyRun(seed string) bool {
	key := fmt.Sprintf("daily:%s", time.Now().Format("2006-01-02"))
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 1, 0, 0, now.Location())
	duration := time.Until(midnight)
	err := rdb.Set(key, seed, duration).Err()
	return err == nil
}

func GetDailyRunSeed() (string, bool) {
	key := fmt.Sprintf("daily:%s", time.Now().Format("2006-01-02"))
	cachedSeed, err := rdb.Get(key).Result()
	if err != nil {
		return "", false
	}

	return cachedSeed, true
}
