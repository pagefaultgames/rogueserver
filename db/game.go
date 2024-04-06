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
	err := handle.QueryRow("SELECT SUM(battles) FROM accountStats").Scan(&battleCount)
	if err != nil {
		return 0, err
	}

	return battleCount, nil
}

func FetchClassicSessionCount() (int, error) {
	var classicSessionCount int
	err := handle.QueryRow("SELECT SUM(classicSessionsPlayed) FROM accountStats").Scan(&classicSessionCount)
	if err != nil {
		return 0, err
	}

	return classicSessionCount, nil
}
