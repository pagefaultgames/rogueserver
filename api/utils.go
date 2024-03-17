package api

import (
	"crypto/md5"
	"encoding/binary"
	"math"
	"time"
)

var seedKey []byte // 32 bytes

func SetSeedKey(key []byte) {
	seedKey = key
}

func SeedFromTime(seedTime time.Time) []byte {
	day := make([]byte, 8)
	binary.BigEndian.PutUint64(day, uint64(math.Floor(float64(seedTime.Unix())/float64(time.Hour*24))))

	sum := md5.Sum(append(seedKey, day...))

	return sum[:]
}
