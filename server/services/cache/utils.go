package cache

import (
	"github.com/goburrow/cache"
)

func key2Int64(key cache.Key) int64 {
	return key.(int64)
}
