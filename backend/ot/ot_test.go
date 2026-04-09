package ot

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocumentStore(t *testing.T) {
	store := NewDocumentStore()
	require.NotNil(t, store)
	assert.NotNil(t, store.docs)
}

func TestDocumentStore_GetOrCreate_NewDocument(t *testing.T) {
	store := NewDocumentStore()
	entityID := uuid.New()

	doc := store.GetOrCreate(entityID, 5)
	require.NotNil(t, doc)
	assert.Equal(t, 5, doc.Version)
}

func TestDocumentStore_GetOrCreate_ExistingDocument(t *testing.T) {
	store := NewDocumentStore()
	entityID := uuid.New()

	doc1 := store.GetOrCreate(entityID, 1)
	doc2 := store.GetOrCreate(entityID, 99) // different version — should return same doc

	assert.Same(t, doc1, doc2, "GetOrCreate should return the same document on second call")
	assert.Equal(t, 1, doc2.Version, "version should not be overwritten")
}

func TestDocument_ApplySteps_Success(t *testing.T) {
	store := NewDocumentStore()
	entityID := uuid.New()
	doc := store.GetOrCreate(entityID, 0)

	steps := []Step{
		{ClientID: "client-a", Data: json.RawMessage(`{"op":"insert"}`)},
		{ClientID: "client-a", Data: json.RawMessage(`{"op":"delete"}`)},
	}

	newVersion, err := doc.ApplySteps(0, steps)
	require.NoError(t, err)
	assert.Equal(t, 2, newVersion)
	assert.Equal(t, 2, doc.Version)
	assert.Len(t, doc.Steps, 2)
	assert.Equal(t, 1, doc.Steps[0].Version)
	assert.Equal(t, 2, doc.Steps[1].Version)
	assert.True(t, doc.dirty)
}

func TestDocument_ApplySteps_VersionMismatch(t *testing.T) {
	store := NewDocumentStore()
	entityID := uuid.New()
	doc := store.GetOrCreate(entityID, 3)

	steps := []Step{
		{ClientID: "client-b", Data: json.RawMessage(`{"op":"insert"}`)},
	}

	// Client thinks version is 0 but server is at 3
	_, err := doc.ApplySteps(0, steps)
	require.Error(t, err)
	assert.Equal(t, ErrVersionMismatch, err)
	assert.Equal(t, 3, doc.Version, "version should not change on mismatch")
}

func TestDocument_StepsSince(t *testing.T) {
	store := NewDocumentStore()
	entityID := uuid.New()
	doc := store.GetOrCreate(entityID, 0)

	steps := []Step{
		{ClientID: "c1", Data: json.RawMessage(`{}`)},
		{ClientID: "c1", Data: json.RawMessage(`{}`)},
		{ClientID: "c1", Data: json.RawMessage(`{}`)},
	}
	_, err := doc.ApplySteps(0, steps)
	require.NoError(t, err)
	// Doc is now at version 3 with steps versioned 1, 2, 3

	result := doc.StepsSince(1)
	assert.Len(t, result, 2, "should return steps with version > 1")
	assert.Equal(t, 2, result[0].Version)
	assert.Equal(t, 3, result[1].Version)

	all := doc.StepsSince(0)
	assert.Len(t, all, 3)

	none := doc.StepsSince(3)
	assert.Len(t, none, 0)
}

func TestDocumentStore_Evict(t *testing.T) {
	store := NewDocumentStore()
	entityID := uuid.New()

	store.GetOrCreate(entityID, 0)

	store.mu.RLock()
	_, exists := store.docs[entityID]
	store.mu.RUnlock()
	assert.True(t, exists, "doc should exist before eviction")

	store.Evict(entityID)

	store.mu.RLock()
	_, exists = store.docs[entityID]
	store.mu.RUnlock()
	assert.False(t, exists, "doc should be gone after eviction")
}

func TestDocumentStore_Evict_NonExistent(t *testing.T) {
	store := NewDocumentStore()
	// Evicting a non-existent ID should not panic
	assert.NotPanics(t, func() {
		store.Evict(uuid.New())
	})
}

func TestDocument_ApplySteps_MaxStepsPruning(t *testing.T) {
	store := NewDocumentStore()
	entityID := uuid.New()
	doc := store.GetOrCreate(entityID, 0)

	// Apply 1000 steps in batches of 100 (not yet pruned)
	version := 0
	for batch := 0; batch < 10; batch++ {
		steps := make([]Step, 100)
		for i := range steps {
			steps[i] = Step{ClientID: "client", Data: json.RawMessage(`{}`)}
		}
		newVer, err := doc.ApplySteps(version, steps)
		require.NoError(t, err)
		version = newVer
	}
	assert.Equal(t, 1000, doc.Version)
	assert.Len(t, doc.Steps, 1000)

	// Apply 1 more step — total 1001 triggers the pruning (max is 1000)
	extraStep := []Step{{ClientID: "client", Data: json.RawMessage(`{}`)}}
	newVer, err := doc.ApplySteps(version, extraStep)
	require.NoError(t, err)
	_ = newVer

	assert.Equal(t, 1001, doc.Version, "version should be 1001")
	assert.Len(t, doc.Steps, 1000, "steps should be pruned to max 1000")
	// Step at version 1 was pruned; oldest remaining should be version 2
	assert.Equal(t, 2, doc.Steps[0].Version, "oldest remaining step should be version 2")
	assert.Equal(t, 1001, doc.Steps[999].Version, "newest step should be version 1001")
}
