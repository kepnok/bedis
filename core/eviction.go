package core

import "github.com/kepnok/bedis/config"

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

func evict() {
	switch config.EvictionStratery {
	case "simple-string":
		evictFirst()
	case "allkeys-random":
		evictAllkeysRandom()
	default:
		evictFirst()
	}
}