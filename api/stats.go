package api

import (
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/pagefaultgames/pokerogue-server/db"
)

var (
	statScheduler       = gocron.NewScheduler(time.UTC)
	playerCount         int
	battleCount         int
	classicSessionCount int
)

func scheduleStatRefresh() {
	statScheduler.Every(10).Second().Do(updateStats)
	statScheduler.StartAsync()
}

func updateStats() {
	var err error
	playerCount, err = db.FetchPlayerCount()
	if err != nil {
		log.Print(err)
	}

	battleCount, err = db.FetchBattleCount()
	if err != nil {
		log.Print(err)
	}

	classicSessionCount, err = db.FetchClassicSessionCount()
	if err != nil {
		log.Print(err)
	}
}
