package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type MockNoteService struct{ mock.Mock }

func (m *MockNoteService) CreateNote(gameID, callerID uuid.UUID, note *models.Note) (models.Note, error) {
	args := m.Called(gameID, callerID, note)
	return args.Get(0).(models.Note), args.Error(1)
}

func (m *MockNoteService) ListGameNotes(gameID, callerID uuid.UUID, filters repositories.NoteFilters) ([]models.Note, error) {
	args := m.Called(gameID, callerID, filters)
	return args.Get(0).([]models.Note), args.Error(1)
}

func (m *MockNoteService) ListGameNotesPaginated(gameID, callerID uuid.UUID, filters repositories.NoteFilters, offset, limit int) ([]models.Note, int64, error) {
	args := m.Called(gameID, callerID, filters, offset, limit)
	return args.Get(0).([]models.Note), args.Get(1).(int64), args.Error(2)
}

func (m *MockNoteService) GetNote(id, callerID uuid.UUID) (models.Note, error) {
	args := m.Called(id, callerID)
	return args.Get(0).(models.Note), args.Error(1)
}

func (m *MockNoteService) UpdateNote(id, callerID uuid.UUID, updates map[string]interface{}) (models.Note, error) {
	args := m.Called(id, callerID, updates)
	return args.Get(0).(models.Note), args.Error(1)
}

func (m *MockNoteService) DeleteNote(id, callerID uuid.UUID) error {
	return m.Called(id, callerID).Error(0)
}
