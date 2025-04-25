/*
	Copyright (C) 2024 - 2025  Pagefault Games

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package api

import (
	"log"
	"time"

	"github.com/pagefaultgames/rogueserver/db"
	"github.com/robfig/cron/v3"
)

var (
	scheduler           = cron.New(cron.WithLocation(time.UTC))
	playerCount         int
	battleCount         int
	classicSessionCount int
)

func scheduleStatRefresh() error {
	_, err := scheduler.AddFunc("@every 30s", func() {
		err := updateStats()
		if err != nil {
			log.Printf("failed to update stats: %s", err)
		}
	})
	if err != nil {
		return err
	}

	scheduler.Start()
	return nil
}

func updateStats() error {
	var err error
	playerCount, err = db.FetchPlayerCount()
	if err != nil {
		return err
	}

	battleCount, err = db.FetchBattleCount()
	if err != nil {
		return err
	}

	classicSessionCount, err = db.FetchClassicSessionCount()
	if err != nil {
		return err
	}

	return nil
}
