package daily

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/pagefaultgames/pokerogue-server/db"
)

const secondsPerDay = 60 * 60 * 24

var (
	dailyRunScheduler = gocron.NewScheduler(time.UTC)
	dailyRunSecret    []byte
)

func Init() error {
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

	err = db.TryAddDailyRun(Seed())
	if err != nil {
		log.Print(err)
	}

	log.Printf("Daily Run Seed: %s", Seed())

	scheduleRefresh()

	return nil
}

func Seed() string {
	return base64.StdEncoding.EncodeToString(deriveDailyRunSeed(time.Now().UTC()))
}

func scheduleRefresh() {
	dailyRunScheduler.Every(1).Day().At("00:00").Do(func() error {
		err := Init()
		if err != nil {
			log.Fatal(err)
		}

		return nil
	}())
	dailyRunScheduler.StartAsync()
}

func deriveDailyRunSeed(seedTime time.Time) []byte {
	day := make([]byte, 8)
	binary.BigEndian.PutUint64(day, uint64(seedTime.Unix()/secondsPerDay))

	hashedSeed := md5.Sum(append(day, dailyRunSecret...))

	return hashedSeed[:]
}
