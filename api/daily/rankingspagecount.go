package daily

import (
	"log"

	"github.com/pagefaultgames/pokerogue-server/db"
)

// /daily/rankingpagecount - fetch daily ranking page count
func RankingPageCount(category int) (int, error) {
	pageCount, err := db.FetchRankingPageCount(category)
	if err != nil {
		log.Print("failed to retrieve ranking page count")
	}

	return pageCount, nil
}
