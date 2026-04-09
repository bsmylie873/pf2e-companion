package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type MockNoteRepository struct{ mock.Mock }

func (m *MockNoteRepository) Create(note *models.Note) error {
	return m.Called(note).Error(0)
}

func (m *MockNoteRepository) FindByGameID(gameID, userID uuid.UUID, isGM bool, filters repositories.NoteFilters) ([]models.Note, error) {
	args := m.Called(gameID, userID, isGM, filters)
	return args.Get(0).([]models.Note), args.Error(1)
}

func (m *MockNoteRepository) FindByGameIDPaginated(gameID, userID uuid.UUID, isGM bool, filters repositories.NoteFilters, offset, limit int) ([]models.Note, int64, error) {
	args := m.Called(gameID, userID, isGM, filters, offset, limit)
	return args.Get(0).([]models.Note), args.Get(1).(int64), args.Error(2)
}

func (m *MockNoteRepository) FindByID(id uuid.UUID) (models.Note, error) {
	args := m.Called(id)
	return args.Get(0).(models.Note), args.Error(1)
}

func (m *MockNoteRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Note, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.Note), args.Error(1)
}

func (m *MockNoteRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockNoteRepository) ClearNoteFromPins(noteID uuid.UUID) error {
	return m.Called(noteID).Error(0)
}
