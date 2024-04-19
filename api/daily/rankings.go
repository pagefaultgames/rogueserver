package daily

import (
	"log"

	"github.com/pagefaultgames/pokerogue-server/db"
	"github.com/pagefaultgames/pokerogue-server/defs"
)

// /daily/rankings - fetch daily rankings
func Rankings(category, page int) ([]defs.DailyRanking, error) {
	rankings, err := db.FetchRankings(category, page)
	if err != nil {
		log.Print("failed to retrieve rankings")
	}

	return rankings, nil
}
