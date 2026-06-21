package core

//simplest eviction staratergy, evicting the first key it finds in the random traversal of the store
func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}

func evict() {
	evictFirst()
}