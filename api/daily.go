package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	var err error
	dailyRunSeed, err = db.GetDailyRunSeed()
	if err != nil {
		log.Printf("failed to generated daily run seed: %s", err.Error())
	}

	if dailyRunSeed == "" {
		dailyRunSeed = RandString(24)
		err := db.TryAddDailyRun(dailyRunSeed)
		if err != nil {
			log.Print(err.Error())
		} else {
			log.Printf("Daily Run Seed: %s", dailyRunSeed)
		}
	}
}

// /daily/seed - get daily run seed

func (s *Server) HandleSeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(dailyRunSeed))
}

// /daily/rankings - fetch daily rankings

func (s *Server) HandleRankings(w http.ResponseWriter, r *http.Request) {
	var err error
	var page int

	if r.URL.Query().Has("page") {
		page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to convert page: %s", err), http.StatusBadRequest)
			return
		}
	} else {
		page = 1
	}

	rankings, err := db.GetRankings(page)
	if err != nil {
		log.Print("failed to retrieve rankings")
	}

	response, err := json.Marshal(rankings)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal response json: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}
