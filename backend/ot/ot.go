package ot

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// ErrVersionMismatch is returned when a client submits steps against a stale version.
var ErrVersionMismatch = errors.New("version mismatch")

// Step represents a single OT operation sent by a client.
type Step struct {
	Version  int             `json:"version"`
	ClientID string          `json:"client_id"`
	Data     json.RawMessage `json:"data"`
}

// Document holds the in-memory OT state for a single entity.
type Document struct {
	mu          sync.Mutex
	Version     int
	Steps       []Step
	dirty       bool
	LastContent json.RawMessage
}

// DocumentStore caches Document instances by entity UUID.
type DocumentStore struct {
	mu   sync.RWMutex
	docs map[uuid.UUID]*Document
}

// NewDocumentStore returns an initialized DocumentStore.
func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		docs: make(map[uuid.UUID]*Document),
	}
}

// GetOrCreate returns the Document for entityID, creating one seeded with currentVersion if absent.
func (s *DocumentStore) GetOrCreate(entityID uuid.UUID, currentVersion int) *Document {
	s.mu.RLock()
	doc, ok := s.docs[entityID]
	s.mu.RUnlock()
	if ok {
		return doc
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// Double-check after acquiring write lock.
	if doc, ok = s.docs[entityID]; ok {
		return doc
	}
	doc = &Document{Version: currentVersion}
	s.docs[entityID] = doc
	return doc
}

// ApplySteps validates the client version and appends steps to the document.
// Returns the new server version or ErrVersionMismatch.
func (d *Document) ApplySteps(clientVersion int, steps []Step) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if clientVersion != d.Version {
		return d.Version, ErrVersionMismatch
	}
	for i := range steps {
		d.Version++
		steps[i].Version = d.Version
		d.Steps = append(d.Steps, steps[i])
	}
	// Keep only the last 1000 steps.
	const maxSteps = 1000
	if len(d.Steps) > maxSteps {
		d.Steps = d.Steps[len(d.Steps)-maxSteps:]
	}
	d.dirty = true
	return d.Version, nil
}

// StepsSince returns all steps with Version greater than fromVersion.
func (d *Document) StepsSince(fromVersion int) []Step {
	d.mu.Lock()
	defer d.mu.Unlock()
	var result []Step
	for _, s := range d.Steps {
		if s.Version > fromVersion {
			result = append(result, s)
		}
	}
	return result
}

// Evict removes the document for entityID from the cache.
func (s *DocumentStore) Evict(entityID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.docs, entityID)
}
