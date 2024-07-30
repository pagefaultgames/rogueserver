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
)

type InfoResponse struct {
	Username        string   `json:"username"`
	DiscordId       string   `json:"discordId"`
	GoogleId        string   `json:"googleId"`
	LastSessionSlot int      `json:"lastSessionSlot"`
	FeatureFlags    []string `json:"featureFlags"`
}

// /account/info - get account info
func Info(username string, discordId string, googleId string, uuid []byte) (InfoResponse, error) {
	slot, _ := db.GetLatestSessionSaveDataSlot(uuid)
	featureFlags := getFeatureFlags(discordId)
	response := InfoResponse{
		Username:        username,
		LastSessionSlot: slot,
		DiscordId:       discordId,
		GoogleId:        googleId,
		FeatureFlags:    featureFlags,
	}
	return response, nil
}

func getFeatureFlags(discordId string) []string {
	var flags []string

	enabledFlags, err := db.GetEnabledFeatureFlags()
	if err != nil {
		return flags
	}

	for _, flag := range enabledFlags {
		var hasAccess = false

		if flag.AccessLevel == EVERYONE {
			hasAccess = true
		} else {
			accessGroup := GetAccessGroupByDiscordRole(discordId)

			if flag.AccessLevel == DEV_STAFF {
				hasAccess = accessGroup == DEV_STAFF
			} else if flag.AccessLevel == CONTRIBUTOR {
				hasAccess = accessGroup == CONTRIBUTOR || accessGroup == DEV_STAFF
			}
		}

		if hasAccess {
			flags = append(flags, flag.Name)
		}
	}

	return flags
}
