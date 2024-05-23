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

func FetchRankings(category int, page int) ([]defs.DailyRanking, error) {
	var rankings []defs.DailyRanking

	offset := (page - 1) * 10

	var query string
	switch category {
	case 0:
		query = "SELECT RANK() OVER (ORDER BY adr.score DESC, adr.timestamp), a.username, adr.score, adr.wave FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date = UTC_DATE() AND a.banned = 0 LIMIT 10 OFFSET ?"
	case 1:
		query = "SELECT RANK() OVER (ORDER BY SUM(adr.score) DESC, adr.timestamp), a.username, SUM(adr.score), 0 FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date >= DATE_SUB(DATE(UTC_TIMESTAMP()), INTERVAL DAYOFWEEK(UTC_TIMESTAMP()) - 1 DAY) AND a.banned = 0 GROUP BY a.username ORDER BY 1 LIMIT 10 OFFSET ?"
	}

	results, err := handle.Query(query, offset)
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

func FetchRankingPageCount(category int) (int, error) {
	var query string
	switch category {
	case 0:
		query = "SELECT COUNT(a.username) FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date = UTC_DATE()"
	case 1:
		query = "SELECT COUNT(DISTINCT a.username) FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date >= DATE_SUB(DATE(UTC_TIMESTAMP()), INTERVAL DAYOFWEEK(UTC_TIMESTAMP()) - 1 DAY)"
	}

	var recordCount int
	err := handle.QueryRow(query).Scan(&recordCount)
	if err != nil {
		return 0, err
	}

	return int(math.Ceil(float64(recordCount) / 10)), nil
}
