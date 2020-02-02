package repository

import (
	"sync"
)

// KeyRepository storage interface for users authentication keys.
type KeyRepository interface {
	Find(username string) ([]byte, bool)
	Save(username string, key []byte) error
}

// NewKeyRepository returns a new key repository using the default implementation.
func NewKeyRepository() KeyRepository {
	return &keyMap{
		mu:   sync.RWMutex{},
		keys: make(map[string][]byte),
	}
}

// keyMap in memory KeyRepository implementation using a synced map.
type keyMap struct {
	mu   sync.RWMutex
	keys map[string][]byte
}

func (km *keyMap) Find(username string) ([]byte, bool) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	key, ok := km.keys[username]
	return key, ok
}

func (km *keyMap) Save(username string, key []byte) error {
	km.mu.Lock()
	defer km.mu.Unlock()
	km.keys[username] = key
	return nil
}
