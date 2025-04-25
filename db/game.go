/*
	Copyright (C) 2024 - 2025  Pagefault Games

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

package db

func FetchPlayerCount() (int, error) {
	var playerCount int
	err := handle.QueryRow("SELECT COUNT(*) FROM accounts WHERE lastActivity > DATE_SUB(UTC_TIMESTAMP(), INTERVAL 5 MINUTE)").Scan(&playerCount)
	if err != nil {
		return 0, err
	}

	return playerCount, nil
}

func FetchBattleCount() (int, error) {
	var battleCount int
	err := handle.QueryRow("SELECT COALESCE(SUM(s.battles), 0) FROM accountStats s JOIN accounts a ON a.uuid = s.uuid WHERE a.banned = 0").Scan(&battleCount)
	if err != nil {
		return 0, err
	}

	return battleCount, nil
}

func FetchClassicSessionCount() (int, error) {
	var classicSessionCount int
	err := handle.QueryRow("SELECT COALESCE(SUM(s.classicSessionsPlayed), 0) FROM accountStats s JOIN accounts a ON a.uuid = s.uuid WHERE a.banned = 0").Scan(&classicSessionCount)
	if err != nil {
		return 0, err
	}

	return classicSessionCount, nil
}
