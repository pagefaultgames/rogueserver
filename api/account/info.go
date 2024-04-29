// Copyright (C) 2024 Pagefault Games - All Rights Reserved
// https://github.com/pagefaultgames

package account

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pagefaultgames/rogueserver/defs"
)

type InfoResponse struct {
	Username        string `json:"username"`
	LastSessionSlot int    `json:"lastSessionSlot"`
}

// /account/info - get account info
func Info(username string, uuid []byte) (InfoResponse, error) {
	var latestSave time.Time
	latestSaveID := -1
	for id := range defs.SessionSlotCount {
		fileName := "session"
		if id != 0 {
			fileName += strconv.Itoa(id)
		}

		stat, err := os.Stat(fmt.Sprintf("userdata/%x/%s.pzs", uuid, fileName))
		if err != nil {
			continue
		}

		if stat.ModTime().After(latestSave) {
			latestSave = stat.ModTime()
			latestSaveID = id
		}
	}

	return InfoResponse{Username: username, LastSessionSlot: latestSaveID}, nil
}
