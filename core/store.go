package core

import (
	"sync"
	"time"

	"github.com/kepnok/bedis/config"
)

var (
	store   map[string]*Obj
	expires map[*Obj]uint64
	mu      sync.Mutex
)

func init() {
	store = make(map[string]*Obj)
	expires = make(map[*Obj]uint64)
}

func NewObj(value interface{}, durationMs int64, oType uint8, oEnc uint8) *Obj {

	obj := &Obj{
		TypeEncoding: oType | oEnc,
		Value:        value,
		LastAccessAt: getCurrentClock(),
	}

	if durationMs > 0 {
		SetExpiry(obj, durationMs)
	}

	return obj
}

func SetExpiry(obj *Obj, durationMs int64) {
	expires[obj] = uint64(time.Now().UnixMilli()) + uint64(durationMs)
}

func hasExpired(obj *Obj) bool {
	exp, ok := expires[obj]
	if !ok {
		return false
	}

	return exp <= uint64(time.Now().UnixMilli())
}

func getExpiry(obj *Obj) (uint64, bool) {
	exp, ok := expires[obj]
	return exp, ok
}

func Put(k string, obj *Obj) {
	mu.Lock()
	defer mu.Unlock()
	if len(store) > config.KeysLimit {
		evict()
	}
	store[k] = obj

	if KeyStats == nil {
		KeyStats = make(map[string]int)
	}
	UpdateDBStat(KEY_METRIC, len(store))
}

func Get(k string) *Obj {
	v := store[k]

	//Here we do a passive delete
	if v != nil {
		if hasExpired(v) {
			Del(k)
			return nil
		}
	}
	return v
}

func Del(k string) bool {
	if obj, ok := store[k]; ok {
		mu.Lock()
		delete(store, k)
		delete(expires, obj)
		mu.Unlock()
		UpdateDBStat(KEY_METRIC, len(store))
		return true
	}
	return false
}
