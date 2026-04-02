package ot

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PersistFunc is called by the persistence loop to save document state to the database.
type PersistFunc func(entityID uuid.UUID, content json.RawMessage, version int) error

// StartPersistenceLoop runs a background goroutine that flushes dirty documents every 5 seconds.
func StartPersistenceLoop(store *DocumentStore, persist PersistFunc) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			store.mu.RLock()
			ids := make([]uuid.UUID, 0, len(store.docs))
			for id := range store.docs {
				ids = append(ids, id)
			}
			store.mu.RUnlock()

			for _, id := range ids {
				store.mu.RLock()
				doc, ok := store.docs[id]
				store.mu.RUnlock()
				if !ok {
					continue
				}

				doc.mu.Lock()
				if !doc.dirty {
					doc.mu.Unlock()
					continue
				}
				content := doc.LastContent
				version := doc.Version
				doc.mu.Unlock()

				if err := persist(id, content, version); err == nil {
					doc.mu.Lock()
					doc.dirty = false
					doc.mu.Unlock()
				}
			}
		}
	}()
}
