package store

import (
	"strings"
	"sync"
)

// CStore stands for ConcurrentStore and is simply an abstract of a map with a mutex.
type CStore struct {
	mu    sync.RWMutex
	store map[string]string
}

// New returns an initialized CStore
func New() CStore {
	return CStore{
		store: make(map[string]string),
	}
}

// Store stores the value with the provided key
func (cs *CStore) Store(key, value string) {
	cs.mu.Lock()
	cs.store[key] = value
	cs.mu.Unlock()
}

// Load loads the value saved by the provided key.
func (cs *CStore) Load(key string) (string, bool) {
	cs.mu.RLock()
	val, ok := cs.store[key]
	cs.mu.RUnlock()

	return val, ok
}

// Contains checks if a key exists in the store.
func (cs *CStore) Contains(key string) bool {
	_, ok := cs.Load(key)

	return ok
}

// AnyContains checks whether the passed needle is contained
// in any of the keys that the store has.
func (cs *CStore) AnyContains(needle string) bool {
	for key := range cs.store {
		if strings.Contains(key, needle) {
			return true
		}
	}

	return false
}

// AnyContainsReverse checks whether any of the keys stored by
// the store is contained in the passed haystack
func (cs *CStore) AnyContainsReverse(haystack string) bool {
	for key := range cs.store {
		if strings.Contains(haystack, key) {
			return true
		}
	}

	return false
}

// StoreKey simply stores the key with empty value
func (cs *CStore) StoreKey(key string) {
	cs.Store(key, "")
}

// Delete deletes an entry from the underlying map
func (cs *CStore) Delete(key string) {
	cs.mu.Lock()
	delete(cs.store, key)
	cs.mu.Unlock()
}

// Size returns the size of the underlying map
func (cs *CStore) Size() int {
	return len(cs.store)
}

// ToMap creates a copy of the underlying map and returns it
func (cs *CStore) ToMap() map[string]string {
	newMap := make(map[string]string, len(cs.store))

	for k, v := range cs.store {
		newMap[k] = v
	}

	return newMap
}
