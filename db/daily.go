package db

import (
	"github.com/Flashfyre/pokerogue-server/defs"
)

func TryAddDailyRun(seed string) error {
	_, err := handle.Exec("INSERT INTO dailyRuns (seed, date) VALUES (?, UTC_DATE()) ON DUPLICATE KEY UPDATE date = date", seed)
	if err != nil {
		return err
	}

	return nil
}

func AddOrUpdateAccountDailyRun(uuid []byte, score int, wave int) error {
	_, err := handle.Exec("INSERT INTO accountDailyRuns (uuid, date, score, wave, timestamp) VALUES (?, UTC_DATE(), ?, ?, UTC_TIMESTAMP()) ON DUPLICATE KEY UPDATE score = GREATEST(score, ?), wave = GREATEST(wave, ?), timestamp = IF(score < ?, UTC_TIMESTAMP(), timestamp)", uuid, score, wave, score, wave, score)
	if err != nil {
		return err
	}

	return nil
}

func FetchRankings(page int) ([]defs.DailyRanking, error) {
	var rankings []defs.DailyRanking

	offset := (page - 1) * 10

	results, err := handle.Query("SELECT RANK() OVER (ORDER BY adr.score DESC, adr.timestamp), a.username, adr.score, adr.wave FROM accountDailyRuns adr JOIN dailyRuns dr ON dr.date = adr.date JOIN accounts a ON adr.uuid = a.uuid WHERE dr.date = UTC_DATE() LIMIT 10 OFFSET ?", offset)
	if err != nil {
		return rankings, err
	}

	defer results.Close()

	for results.Next() {
		ranking := defs.DailyRanking{}
		err = results.Scan(&ranking.Rank, &ranking.Username, &ranking.Score, &ranking.Wave)
		if err != nil {
			return rankings, err
		}

		rankings = append(rankings, ranking)
	}

	return rankings, nil
}
