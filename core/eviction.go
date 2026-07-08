package core

import (
	"time"

	"github.com/kepnok/bedis/config"
)

var LRU_CLOCK_MAX uint32 = 0x00FFFFFF

func getCurrentClock() uint32 {
	return uint32(time.Now().Unix()) & LRU_CLOCK_MAX
}

//simplest eviction staratergy, evicting the first key it finds in the random traversal of the store
func evictFirst() {
	for k := range store {
		Del(k)
		return
	}
}

func evictAllkeysRandom() {
	// Randomly evicts keys to make space for new keys.

	numberOfKeys := int64(config.EvictionRatio * float64(config.KeysLimit));

	// Golang iteration of maps are fairly random
	for k := range store {
		Del(k)
		numberOfKeys--
		if numberOfKeys <= 0 {
			break
		}
	}

}

func getIdleTime(lastAccessAt uint32) uint32 {
	c := getCurrentClock()
	if c >= lastAccessAt {
		return c - lastAccessAt
	}
	return (LRU_CLOCK_MAX - lastAccessAt) + c
}


func populateEvictionPool() {
	sampleSize := 5
	for k := range store {
		ePool.Push(k, store[k].LastAccessAt)
		sampleSize--
		if sampleSize == 0 {
			break
		}
	}
}

// TODO: no need to populate everytime. should populate
// only when the number of keys to evict is less than what we have in the pool
func evictAllkeysLRU() {
	
	evictCount := int16(config.EvictionRatio * float64(config.KeysLimit))
	if len(ePool.pool) < int(evictCount) {
		populateEvictionPool()
	}
	for i := 0; i < int(evictCount) && len(ePool.pool) > 0; i++ {
		item := ePool.Pop()
		if item == nil {
			return
		}
		Del(item.key)
	}
}

func evict() {
	switch config.EvictionStratery {
	case "simple-string":
		evictFirst()
	case "allkeys-random":
		evictAllkeysRandom()
	case "allkeys-lru":
		evictAllkeysLRU()
	default:
		evictFirst()
	}
}