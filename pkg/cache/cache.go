package cache

import (
	"sync"

	"github.com/bluele/gcache"
)

var (
	client gcache.Cache
	once   sync.Once
)

func initCache() {
	client = gcache.New(10000).LRU().Build()
}

func GetClient() gcache.Cache {
	once.Do(func() {
		initCache()
	})
	return client
}
