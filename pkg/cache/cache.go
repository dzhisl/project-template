package cache

import (
	"github.com/bluele/gcache"
)

var Client gcache.Cache

func InitCache() {
	Client = gcache.New(10000).LRU().Build()
}
