// Copyright (C) 2024 Pagefault Games - All Rights Reserved
// https://github.com/pagefaultgames

package daily

import (
	"log"

	"github.com/pagefaultgames/rogueserver/db"
)

// /daily/rankingpagecount - fetch daily ranking page count
func RankingPageCount(category int) (int, error) {
	pageCount, err := db.FetchRankingPageCount(category)
	if err != nil {
		log.Print("failed to retrieve ranking page count")
	}

	return pageCount, nil
}
