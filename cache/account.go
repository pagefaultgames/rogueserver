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
)

func AddAccountSession(uuid []byte, token []byte) bool {
	rdb.Do("SELECT", sessionDB)
	err := rdb.Set(string(token), string(uuid), 7*24*time.Hour).Err()
	return err == nil
}

func FetchUsernameBySessionToken(token []byte) (string, bool) {
	rdb.Do("SELECT", sessionDB)
	username, err := rdb.Get(string(token)).Result()
	if err != nil {
		return "", false
	}

	return username, true
}

func updateActivePlayers(uuid []byte) bool {
	rdb.Do("SELECT", activePlayersDB)
	err := rdb.Set(string(uuid), 1, 0).Err()
	if err != nil {
		return false
	}
	err = rdb.Expire(string(uuid), 5*time.Minute).Err()
	return err == nil
}

func UpdateAccountLastActivity(uuid []byte) bool {
	rdb.Do("SELECT", accountsDB)
	err := rdb.HSet(string(uuid), "lastActivity", time.Now().Format("2006-01-02 15:04:05")).Err()
	if err != nil {
		return false
	}
	updateActivePlayers(uuid)

	return err == nil
}

// FIXME: This is not atomic
func UpdateAccountStats(uuid []byte, battles, classicSessionsPlayed int) bool {
	rdb.Do("SELECT", accountsDB)
	err := rdb.HIncrBy(string(uuid), "battles", int64(battles)).Err()
	if err != nil {
		return false
	}
	err = rdb.HIncrBy(string(uuid), "classicSessionsPlayed", int64(classicSessionsPlayed)).Err()
	return err == nil
}

func FetchTrainerIds(uuid []byte) (int, int, bool) {
	rdb.Do("SELECT", accountsDB)
	vals, err := rdb.HMGet(string(uuid), "trainerId", "secretId").Result()
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
	rdb.Do("SELECT", accountsDB)
	err := rdb.HMSet(string(uuid), map[string]interface{}{
		"trainerId": trainerId,
		"secretId":  secretId,
	}).Err()
	return err == nil
}

func IsActiveSession(uuid []byte, sessionId string) (bool, bool) {
	rdb.Do("SELECT", activeClientSessionsDB)
	id, err := rdb.Get(string(uuid)).Result()
	if err != nil {
		return false, false
	}

	return id == sessionId, true
}

func UpdateActiveSession(uuid []byte, sessionId string) bool {
	rdb.Do("SELECT", activeClientSessionsDB)
	err := rdb.Set(string(uuid), sessionId, 0).Err()
	return err == nil
}

func FetchUUIDFromToken(token []byte) ([]byte, bool) {
	rdb.Do("SELECT", sessionDB)
	uuid, err := rdb.Get(string(token)).Bytes()
	if err != nil {
		return nil, false
	}

	return uuid, true
}

func RemoveSessionFromToken(token []byte) bool {
	rdb.Do("SELECT", sessionDB)
	err := rdb.Del(string(token)).Err()
	return err == nil
}
