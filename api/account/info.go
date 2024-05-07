/*
	Copyright (C) 2024  Pagefault Games

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

package account

import (
	"github.com/pagefaultgames/rogueserver/db"
	"github.com/pagefaultgames/rogueserver/defs"
)

type InfoResponse struct {
	Username        string `json:"username"`
	LastSessionSlot int    `json:"lastSessionSlot"`
}

// /account/info - get account info
func Info(username string, uuid []byte) (InfoResponse, error) {
	response := InfoResponse{Username: username, LastSessionSlot: -1}

	highest := -1
	for i := 0; i < defs.SessionSlotCount; i++ {
		data, err := db.ReadSessionSaveData(uuid, i)
		if err != nil {
			continue
		}

		if data.Timestamp > highest {
			highest = data.Timestamp
			response.LastSessionSlot = i
		}
	}

	if response.LastSessionSlot < 0 || response.LastSessionSlot >= defs.SessionSlotCount {
		response.LastSessionSlot = -1
	}

	return response, nil
}
