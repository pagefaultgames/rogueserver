package api

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Flashfyre/pokerogue-server/db"
	"github.com/Flashfyre/pokerogue-server/defs"
	"github.com/go-co-op/gocron"
)

const secondsPerDay = 60 * 60 * 24

var (
	dailyRunScheduler = gocron.NewScheduler(time.UTC)
	dailyRunSecret    []byte
	dailyRunSeed      string
)

func ScheduleDailyRunRefresh() {
	dailyRunScheduler.Every(1).Day().At("00:00").Do(func() error {
		err := InitDailyRun()
		if err != nil {
			log.Fatal(err)
		}

		return nil
	}())
	dailyRunScheduler.StartAsync()
}

func InitDailyRun() error {
	secret, err := os.ReadFile("secret.key")
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read daily seed secret: %s", err)
		}

		newSecret := make([]byte, 32)
		_, err := rand.Read(newSecret)
		if err != nil {
			return fmt.Errorf("failed to generate daily seed secret: %s", err)
		}

		err = os.WriteFile("secret.key", newSecret, 0400)
		if err != nil {
			return fmt.Errorf("failed to write daily seed secret: %s", err)
		}

		secret = newSecret
	}

	dailyRunSecret = secret

	dailyRunSeed = base64.StdEncoding.EncodeToString(deriveDailyRunSeed(time.Now().UTC()))

	err = db.TryAddDailyRun(dailyRunSeed)
	if err != nil {
		log.Print(err)
	}

	log.Printf("Daily Run Seed: %s", dailyRunSeed)

	return nil
}

func deriveDailyRunSeed(seedTime time.Time) []byte {
	day := make([]byte, 8)
	binary.BigEndian.PutUint64(day, uint64(seedTime.Unix()/secondsPerDay))

	hashedSeed := md5.Sum(append(day, dailyRunSecret...))

	return hashedSeed[:]
}

// /daily/rankings - fetch daily rankings
func handleRankings(uuid []byte, category, page int) ([]defs.DailyRanking, error) {
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

// /daily/rankingpagecount - fetch daily ranking page count
func handleRankingPageCount(category int) (int, error) {
	pageCount, err := db.FetchRankingPageCount(category)
	if err != nil {
		log.Print("failed to retrieve ranking page count")
	}

	return pageCount, nil
}
