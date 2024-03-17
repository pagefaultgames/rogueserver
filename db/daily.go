package db

func TryAddDailyRun(seed string) error {
	_, err := handle.Exec("INSERT INTO dailyRuns (seed, date) VALUES (?, UTC_DATE()) ON DUPLICATE KEY UPDATE date = date", seed)
	if err != nil {
		return err
	}

	return nil
}
