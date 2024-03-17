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

func GetRankings(page int) ([]defs.DailyRanking, error) {
	var rankings []defs.DailyRanking

	offset := (page - 1) * 10

	results, err := handle.Query("SELECT RANK() OVER (ORDER BY sc.score DESC, sc.timestamp), a.username, sc.score FROM seedCompletions sc JOIN dailyRuns dr ON dr.seed = sc.seed JOIN accounts a ON sc.uuid = a.uuid WHERE dr.date = UTC_DATE() LIMIT 10 OFFSET ?", offset)
	if err != nil {
		return rankings, err
	}

	defer results.Close()

	for results.Next() {
		ranking := defs.DailyRanking{}
		err = results.Scan(&ranking.Rank, &ranking.Username, &ranking.Score)
		if err != nil {
			return rankings, err
		}

		rankings = append(rankings, ranking)
	}

	return rankings, nil
}
