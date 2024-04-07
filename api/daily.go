package api

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Flashfyre/pokerogue-server/db"
	"github.com/go-co-op/gocron"
)

const secondsPerDay = 60 * 60 * 24

var (
	dailyRunScheduler = gocron.NewScheduler(time.UTC)
	dailyRunSecret    []byte
	dailyRunSeed      string
)

func ScheduleDailyRunRefresh() {
	dailyRunScheduler.Every(1).Day().At("00:00").Do(InitDailyRun)
	dailyRunScheduler.StartAsync()
}

func InitDailyRun() {
	secret, err := os.ReadFile("secret.key")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("failed to read daily seed secret: %s", err)
		}

		newSecret := make([]byte, 32)
		_, err := rand.Read(newSecret)
		if err != nil {
			log.Fatalf("failed to generate daily seed secret: %s", err)
		}

		err = os.WriteFile("secret.key", newSecret, 0400)
		if err != nil {
			log.Fatalf("failed to write daily seed secret: %s", err)
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
}

func deriveDailyRunSeed(seedTime time.Time) []byte {
	day := make([]byte, 8)
	binary.BigEndian.PutUint64(day, uint64(seedTime.Unix()/secondsPerDay))

	hashedSeed := md5.Sum(append(day, dailyRunSecret...))

	return hashedSeed[:]
}

// /daily/seed - fetch daily run seed
func (s *Server) handleSeed(w http.ResponseWriter) {
	w.Write([]byte(dailyRunSeed))
}

// /daily/rankings - fetch daily rankings
func (s *Server) handleRankings(w http.ResponseWriter, r *http.Request) {
	uuid, err := getUuidFromRequest(r)
	if err != nil {
		httpError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateAccountLastActivity(uuid)
	if err != nil {
		log.Print("failed to update account last activity")
	}

	var category int
	if r.URL.Query().Has("category") {
		category, err = strconv.Atoi(r.URL.Query().Get("category"))
		if err != nil {
			httpError(w, r, fmt.Sprintf("failed to convert category: %s", err), http.StatusBadRequest)
			return
		}
	}

	page := 1
	if r.URL.Query().Has("page") {
		page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			httpError(w, r, fmt.Sprintf("failed to convert page: %s", err), http.StatusBadRequest)
			return
		}
	}

	rankings, err := db.FetchRankings(category, page)
	if err != nil {
		log.Print("failed to retrieve rankings")
	}

	response, err := json.Marshal(rankings)
	if err != nil {
		httpError(w, r, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

// /daily/rankingpagecount - fetch daily ranking page count
func (s *Server) handleRankingPageCount(w http.ResponseWriter, r *http.Request) {
	var err error
	var category int

	if r.URL.Query().Has("category") {
		category, err = strconv.Atoi(r.URL.Query().Get("category"))
		if err != nil {
			httpError(w, r, fmt.Sprintf("failed to convert category: %s", err), http.StatusBadRequest)
			return
		}
	}

	pageCount, err := db.FetchRankingPageCount(category)
	if err != nil {
		log.Print("failed to retrieve ranking page count")
	}

	response, err := json.Marshal(pageCount)
	if err != nil {
		httpError(w, r, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}
