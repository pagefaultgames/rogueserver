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

package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var c *cache.Cache

func init() {
	// Create a new cache with a default expiration time of 5 minutes, and which purges expired items every 10 minutes
	c = cache.New(5*time.Minute, 10*time.Minute)
}

func Set(key string, value interface{}) {
	c.Set(key, value, cache.DefaultExpiration)
}

func Get(key string) (interface{}, bool) {
	return c.Get(key)
}
