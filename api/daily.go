package api

import (
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/Flashfyre/pokerogue-server/db"
)

var (
	dailyRunSeed string
)

func ScheduleDailyRunRefresh() {
	scheduler.Every(1).Day().At("00:00").Do(func() {
		InitDailyRun()
	})
}

func InitDailyRun() {
	dailyRunSeed = base64.StdEncoding.EncodeToString(SeedFromTime(time.Now().UTC()))
	err := db.TryAddDailyRun(dailyRunSeed)
	if err != nil {
		log.Print(err.Error())
	} else {
		log.Printf("Daily Run Seed: %s", dailyRunSeed)
	}
}

// /daily/seed - get daily run seed

func (s *Server) HandleSeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(dailyRunSeed))
}
