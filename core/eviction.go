package core

import "github.com/kepnok/bedis/config"

//simplest eviction staratergy, evicting the first key it finds in the random traversal of the store
func evictFirst() {
	for k := range store {
		mu.Lock()
		delete(store, k)
		mu.Unlock()
		return
	}
}

func evict() {
	switch config.EvictionStratery {
	case "simple-string":
		evictFirst()
	default:
		evictFirst()
	}
}