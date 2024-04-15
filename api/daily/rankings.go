package daily

import (
	"log"

	"github.com/pagefaultgames/pokerogue-server/db"
	"github.com/pagefaultgames/pokerogue-server/defs"
)

// /daily/rankings - fetch daily rankings
func Rankings(uuid []byte, category, page int) ([]defs.DailyRanking, error) {
	err := db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	rankings, err := db.FetchRankings(category, page)
	if err != nil {
		log.Print("failed to retrieve rankings")
	}

	return rankings, nil
}
