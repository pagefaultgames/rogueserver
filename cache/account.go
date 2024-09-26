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
	"fmt"
	"time"
)

func AddAccountSession(uuid []byte, token []byte) bool {
	key := fmt.Sprintf("session:%s", token)
	err := rdb.Set(key, string(uuid), 24*time.Hour).Err()
	return err == nil
}

func FetchUsernameBySessionToken(token []byte) (string, bool) {
	key := fmt.Sprintf("session:%s", token)
	username, err := rdb.Get(key).Result()
	if err != nil {
		return "", false
	}

	return username, true
}

func UpdateAccountLastActivity(uuid []byte) bool {
	key := fmt.Sprintf("account:%s", uuid)
	err := rdb.HSet(key, "lastActivity", time.Now().Format("2006-01-02 15:04:05")).Err()
	if err != nil {
		return false
	}
	err = rdb.Expire(key, 5*time.Minute).Err()
	return err == nil
}

func FetchTrainerIds(uuid []byte) (int, int, bool) {
	key := fmt.Sprintf("account:%s", uuid)
	vals, err := rdb.HMGet(key, "trainerId", "secretId").Result()
	if err == nil && len(vals) == 2 && vals[0] != nil && vals[1] != nil {
		trainerId, ok1 := vals[0].(int)
		secretId, ok2 := vals[1].(int)
		if ok1 && ok2 {
			return trainerId, secretId, true
		}
	}

	return 0, 0, false
}

func UpdateTrainerIds(trainerId, secretId int, uuid []byte) bool {
	key := fmt.Sprintf("account:%s", uuid)
	err := rdb.HMSet(key, map[string]interface{}{
		"trainerId": trainerId,
		"secretId":  secretId,
	}).Err()
	if err != nil {
		return false
	}

	err = rdb.Expire(key, 5*time.Minute).Err()
	return err == nil
}

func IsActiveSession(uuid []byte, sessionId string) (bool, bool) {
	key := fmt.Sprintf("active_sessions:%s", uuid)
	id, err := rdb.Get(key).Result()
	return id == sessionId, err == nil
}

func UpdateActiveSession(uuid []byte, sessionId string) bool {
	key := fmt.Sprintf("active_sessions:%s", uuid)
	err := rdb.Set(key, sessionId, 0).Err()
	if err != nil {
		return false
	}
	err = rdb.Expire(key, 5*time.Minute).Err()
	if err != nil {
		return false
	}

	err = rdb.SAdd("active_players", uuid).Err()
	if err != nil {
		return false
	}
	err = rdb.Expire("active_players", 5*time.Minute).Err()

	return err == nil
}

func FetchUUIDFromToken(token []byte) ([]byte, bool) {
	key := fmt.Sprintf("session:%s", token)
	uuid, err := rdb.Get(key).Bytes()
	return uuid, err == nil
}

func RemoveSessionFromToken(token []byte) bool {
	key := fmt.Sprintf("session:%s", token)
	err := rdb.Del(key).Err()
	return err == nil
}
