package db

func FetchPlayerCount() (int, error) {
	var playerCount int
	err := handle.QueryRow("SELECT COUNT(*) FROM accounts WHERE lastActivity > DATE_SUB(UTC_TIMESTAMP(), INTERVAL 5 MINUTE)").Scan(&playerCount)
	if err != nil {
		return 0, err
	}

	return playerCount, nil
}
