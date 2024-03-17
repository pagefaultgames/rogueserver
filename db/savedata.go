package db

func TryAddSeedCompletion(uuid []byte, seed string, mode int, score int) (bool, error) {
	if len(seed) < 24 {
		for range 24 - len(seed) {
			seed += "0"
		}
	}

	newCompletion := true

	var count int
	err := handle.QueryRow("SELECT COUNT(*) FROM seedCompletions WHERE uuid = ? AND seed = ?", uuid, seed).Scan(&count)
	if err != nil {
		return false, err
	} else if count > 0 {
		newCompletion = false
	}

	_, err = handle.Exec("INSERT INTO seedCompletions (uuid, seed, mode, score, timestamp) VALUES (?, ?, ?, ?, UTC_TIMESTAMP()) ON DUPLICATE KEY UPDATE score = ?, timestamp = IF(score < ?, UTC_TIMESTAMP(), timestamp)", uuid, seed, mode, score, score, score)
	if err != nil {
		return false, err
	}

	return newCompletion, nil
}
