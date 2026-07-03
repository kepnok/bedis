package core

import (
	"sync"
	"time"

	"github.com/kepnok/bedis/config"
)

var (
	store map[string]*Obj
	mu    sync.Mutex
)

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, durationMs int64, oType uint8, oEnc uint8) *Obj {
	expiresAt := int64(-1)
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}

	return &Obj{
		TypeEncoding: oType | oEnc,
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

func Put(k string, obj *Obj) {
	mu.Lock()
	defer mu.Unlock()

	if len(store) > config.KeysLimit {
		evict()
	}
	store[k] = obj
}

func Get(k string) *Obj {
	v := store[k]

	//Here we do a passive delete
	if v != nil {
		if v.ExpiresAt != -1 && v.ExpiresAt <= time.Now().UnixMilli() {
			mu.Lock()
			delete(store, k)
			mu.Unlock()
			return nil
		}
	}
	return v
}

func Del(k string) bool {
	if _, ok := store[k]; ok {
		mu.Lock()
		delete(store, k)
		mu.Unlock()
		return true
	}
	return false
}
