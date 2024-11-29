package savedata

import (
	"github.com/patrickmn/go-cache"
)

var Cache = cache.New(cache.NoExpiration, cache.NoExpiration)
