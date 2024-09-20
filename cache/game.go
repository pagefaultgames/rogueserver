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

func FetchPlayerCount() (int, bool) {
	rdb.Do("SELECT", activePlayersDB)
	cachedPlayerCount, err := rdb.DBSize().Result()
	if err != nil {
		return 0, false
	}

	return int(cachedPlayerCount), true
}

func FetchBattleCount() (int, bool) {
	rdb.Do("SELECT", accountsDB)
	cachedBattleCount, err := rdb.Get("battleCount").Int()
	if err != nil {
		return 0, false
	}

	return cachedBattleCount, true
}

func UpdateBattleCount(battleCount int) bool {
	rdb.Do("SELECT", accountsDB)
	err := rdb.Set("battleCount", battleCount, 0).Err()
	return err == nil
}

func FetchClassicSessionCount() (int, bool) {
	rdb.Do("SELECT", accountsDB)
	cachedClassicSessionCount, err := rdb.Get("classicSessionCount").Int()
	if err != nil {
		return 0, false
	}

	return cachedClassicSessionCount, true
}

func UpdateClassicSessionCount(classicSessionCount int) bool {
	rdb.Do("SELECT", accountsDB)
	err := rdb.Set("classicSessionCount", classicSessionCount, 0).Err()
	return err == nil
}
