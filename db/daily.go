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

package db

import (
	"math"
	"database/sql"

	"github.com/pagefaultgames/rogueserver/defs"
)

func TryAddDailyRun(seed string) (string, error) {
	var actualSeed string
	err := handle.QueryRow("INSERT INTO dailyRuns (seed, date) VALUES (?, UTC_DATE()) ON DUPLICATE KEY UPDATE date = date RETURNING seed", seed).Scan(&actualSeed)
	if err != nil {
		return "", err
	}

	return actualSeed, nil
}

func GetDailyRunSeed() (string, error) {
	var seed string
	err := handle.QueryRow("SELECT seed FROM dailyRuns WHERE date = UTC_DATE()").Scan(&seed)
	if err != nil {
		return "", err
	}

	return seed, nil

}

func AddOrUpdateAccountDailyRun(uuid []byte, score int, wave int) error {
	_, err := handle.Exec("INSERT INTO accountDailyRuns (uuid, date, score, wave, timestamp) VALUES (?, UTC_DATE(), ?, ?, UTC_TIMESTAMP()) ON DUPLICATE KEY UPDATE score = GREATEST(score, ?), wave = GREATEST(wave, ?), timestamp = IF(score < ?, UTC_TIMESTAMP(), timestamp)", uuid, score, wave, score, wave, score)
	if err != nil {
		return err
	}

	return nil
}

func FetchRankings(category int, page int, uuid []byte) ([]defs.DailyRanking, error) {
	var rankings []defs.DailyRanking

	username, err := FetchUsernameFromUUID(uuid);
	if err != nil {
		return rankings, err
	}

	offset := (page - 1) * 10

	var query string
	switch category {
	case 0:
		query = "SELECT RANK() OVER (ORDER BY adr.score DESC, adr.timestamp), a.username, adr.score, adr.wave FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date = UTC_DATE() AND a.banned = 0 LIMIT 10 OFFSET ?"
	case 1:
		query = "SELECT RANK() OVER (ORDER BY SUM(adr.score) DESC, adr.timestamp), a.username, SUM(adr.score), 0 FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date >= DATE_SUB(DATE(UTC_TIMESTAMP()), INTERVAL DAYOFWEEK(UTC_TIMESTAMP()) - 1 DAY) AND a.banned = 0 GROUP BY a.username ORDER BY 1 LIMIT 10 OFFSET ?"
	case 2:
		// We retrieve the friends of the user and the user itself
		query = `SELECT RANK() OVER (ORDER BY score DESC, timestamp) AS rank, username, score, wave
				FROM (
					SELECT a.username, adr.score, adr.wave, adr.timestamp
					FROM accountDailyRuns adr
					JOIN dailyRuns dr ON dr.date = adr.date
					JOIN accounts a ON adr.uuid = a.uuid
					JOIN friends f ON a.username = f.friend
					WHERE dr.date = UTC_DATE()
					AND a.banned = 0
					AND f.user = ?
					UNION
					SELECT a.username, adr.score, adr.wave, adr.timestamp
					FROM accountDailyRuns adr
					JOIN dailyRuns dr ON dr.date = adr.date
					JOIN accounts a ON adr.uuid = a.uuid
					WHERE dr.date = UTC_DATE()
					AND a.banned = 0
					AND a.username = ?
				) AS combined LIMIT 10 OFFSET ?;`
	}

	var results *sql.Rows
	if category == 2 {
		results, err = handle.Query(query, username, username, offset)
	} else {
		results, err = handle.Query(query, offset)
	}

	if err != nil {
		return rankings, err
	}

	defer results.Close()

	for results.Next() {
		var ranking defs.DailyRanking
		err = results.Scan(&ranking.Rank, &ranking.Username, &ranking.Score, &ranking.Wave)
		if err != nil {
			return rankings, err
		}

		rankings = append(rankings, ranking)
	}

	return rankings, nil
}

func FetchRankingPageCount(category int, uuid []byte) (int, error) {
	username, err := FetchUsernameFromUUID(uuid);
	if err != nil {
		return 0, err
	}

	var query string
	switch category {
	case 0:
		query = "SELECT COUNT(a.username) FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date = UTC_DATE()"
	case 1:
		query = "SELECT COUNT(DISTINCT a.username) FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date >= DATE_SUB(DATE(UTC_TIMESTAMP()), INTERVAL DAYOFWEEK(UTC_TIMESTAMP()) - 1 DAY)"
	case 2:
		query = `SELECT COUNT(a.username)
				 FROM accountDailyRuns adr
				 JOIN dailyRuns dr ON dr.date = adr.date
				 JOIN accounts a ON adr.uuid = a.uuid
				 JOIN friends f ON a.username = f.friend
				 WHERE dr.date = UTC_DATE()
				 AND f.user = ?
				 OR a.username = ?`
	}

	var recordCount int
	if category == 2 {	
		err = handle.QueryRow(query, username, username).Scan(&recordCount)
		// We only fetch friends of the account, not the account itself so adding +1 here.
		// this way, we don't have to do the big union query like in FetchRankings
		recordCount += 1 
	} else {
		err = handle.QueryRow(query).Scan(&recordCount)
	}

	return int(math.Ceil(float64(recordCount) / 10)), nil
}
