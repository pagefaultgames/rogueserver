// Copyright (C) 2024 Pagefault Games - All Rights Reserved
// https://github.com/pagefaultgames

package daily

import (
	"log"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

// /daily/rankings - fetch daily rankings
func Rankings(category, page int) ([]defs.DailyRanking, error) {
	rankings, err := db.FetchRankings(category, page)
	if err != nil {
		log.Print("failed to retrieve rankings")
	}

	return rankings, nil
}
