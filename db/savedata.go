package db

func TryAddSeedCompletion(uuid []byte, seed string) (bool, error) {
	if len(seed) < 24 {
		for range 24 - len(seed) {
			seed += "0"
		}
	}

	var count int
	err := handle.QueryRow("SELECT COUNT(*) FROM seedCompletions WHERE uuid = ? AND seed = ?", uuid, seed).Scan(&count)
	if err != nil {
		return false, err
	} else if count > 0 {
		return false, nil
	}

	_, err = handle.Exec("INSERT INTO seedCompletions (uuid, seed, timestamp) VALUES (?, ?, UTC_TIMESTAMP())", uuid, seed)
	if err != nil {
		return false, err
	}

	return true, nil
}
