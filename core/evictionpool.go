package core

import "slices"

type PoolItem struct {
	key            string
	lastAccessedAt uint32
}

// TODO: When last accessed at of object changes
// update the poolItem correponding to that
type EvictionPool struct {
	pool   []*PoolItem
	keyset map[string]*PoolItem
}

func (pq *EvictionPool) Push(key string, lastAccessedAt uint32) {

	if _, ok := pq.keyset[key]; ok {
		return
	}

	item := &PoolItem{
		key:            key,
		lastAccessedAt: lastAccessedAt,
	}

	if len(pq.pool) == ePoolSizeMax && lastAccessedAt >= pq.pool[len(pq.pool)-1].lastAccessedAt {
		return
	}	

	idx, _ := slices.BinarySearchFunc(pq.pool, item, func(a, b *PoolItem) int {
		if getIdleTime(a.lastAccessedAt) > getIdleTime(b.lastAccessedAt) {
			return -1
		} else if getIdleTime(a.lastAccessedAt) < getIdleTime(b.lastAccessedAt) {
			return 1
		}

		return 0
	})

	pq.keyset[key] = item
	if len(pq.pool) < ePoolSizeMax {
		//since the pool is not full yet we can say for sure that the len is not as much as the limit we have set so inserting nil will grow our pool by 1
		pq.pool = append(pq.pool, nil)

		// this shift works similary to the memmove in c, that redis uses in its implementation; we move the elemnts right by 1
		copy(pq.pool[idx+1:], pq.pool[idx:])

		pq.pool[idx] = item

	} else {
		// here our slice is completely full, so we will remove the last element from the keySet and then shift the elemts and insert item to index
		lastElement := pq.pool[len(pq.pool) - 1]
		delete(pq.keyset, lastElement.key)

		copy(pq.pool[idx + 1:], pq.pool[idx:])
		pq.pool[idx] = item
	}
	// ***** old logic *****

	// _, ok := pq.keyset[key]
	// if ok {
	// 	return
	// }
	// item := &PoolItem{key: key, lastAccessedAt: lastAccessedAt}
	// if len(pq.pool) < ePoolSizeMax {
	// 	pq.keyset[key] = item
	// 	pq.pool = append(pq.pool, item)

	// 	// Performance bottleneck
	// 	sort.Sort(ByIdleTime(pq.pool))
	// } else if lastAccessedAt > pq.pool[0].lastAccessedAt {
	// 	pq.pool = pq.pool[1:]
	// 	pq.keyset[key] = item
	// 	pq.pool = append(pq.pool, item)
	// 	sort.Sort(ByIdleTime(pq.pool))
	// }

	// ***** old logic *****
}

func (pq *EvictionPool) Pop() *PoolItem {
	if len(pq.pool) == 0 {
		return nil
	}
	item := pq.pool[0]
	pq.pool = pq.pool[1:]
	delete(pq.keyset, item.key)
	return item
}

func newEvictionPool(size int) *EvictionPool {
	return &EvictionPool{
		pool:   make([]*PoolItem, size),
		keyset: make(map[string]*PoolItem),
	}
}

var ePoolSizeMax int = 16
var ePool *EvictionPool = newEvictionPool(0)
