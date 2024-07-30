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

package db

import (
	"github.com/pagefaultgames/rogueserver/defs"
)

func GetEnabledFeatureFlags() ([]defs.FeatureFlag, error) {
	var activeFlags []defs.FeatureFlag

	results, err := handle.Query("SELECT name, accessLevel FROM featureFlags WHERE enabled = 1")

	if err != nil {
		return activeFlags, err
	}

	defer results.Close()

	for results.Next() {
		var flag defs.FeatureFlag

		err = results.Scan(&flag.Name, &flag.AccessLevel)

		if err != nil {
			return activeFlags, err
		}

		activeFlags = append(activeFlags, flag)
	}

	return activeFlags, nil
}
