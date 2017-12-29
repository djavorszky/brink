package store

import "sync"

// CStore stands for ConcurrentStore and is simply an abstract of a map with a mutex.
type CStore struct {
	mu    sync.RWMutex
	store map[string]string
}

// New returns an initialized CStore
func New() CStore {
	return CStore{store: make(map[string]string)}
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

// Exists checks if a key exists in the store.
func (cs *CStore) Exists(key string) bool {
	_, ok := cs.Load(key)

	return ok
}

// StoreKey simply stores the key with empty value
func (cs *CStore) StoreKey(key string) {
	cs.Store(key, "")
}
