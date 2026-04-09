package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockMapService struct{ mock.Mock }

func (m *MockMapService) CreateMap(gameID, callerID uuid.UUID, name string, description *string) (models.GameMap, error) {
	args := m.Called(gameID, callerID, name, description)
	return args.Get(0).(models.GameMap), args.Error(1)
}

func (m *MockMapService) ListMaps(gameID, callerID uuid.UUID) ([]models.GameMap, error) {
	args := m.Called(gameID, callerID)
	return args.Get(0).([]models.GameMap), args.Error(1)
}

func (m *MockMapService) ListArchivedMaps(gameID, callerID uuid.UUID) ([]models.GameMap, error) {
	args := m.Called(gameID, callerID)
	return args.Get(0).([]models.GameMap), args.Error(1)
}

func (m *MockMapService) GetMap(id, callerID uuid.UUID) (models.GameMap, error) {
	args := m.Called(id, callerID)
	return args.Get(0).(models.GameMap), args.Error(1)
}

func (m *MockMapService) RenameMap(id, callerID uuid.UUID, name string, description *string) (models.GameMap, error) {
	args := m.Called(id, callerID, name, description)
	return args.Get(0).(models.GameMap), args.Error(1)
}

func (m *MockMapService) ReorderMaps(gameID, callerID uuid.UUID, orderedIDs []uuid.UUID) error {
	return m.Called(gameID, callerID, orderedIDs).Error(0)
}

func (m *MockMapService) ArchiveMap(id, callerID uuid.UUID) error {
	return m.Called(id, callerID).Error(0)
}

func (m *MockMapService) RestoreMap(id, callerID uuid.UUID) (models.GameMap, error) {
	args := m.Called(id, callerID)
	return args.Get(0).(models.GameMap), args.Error(1)
}

func (m *MockMapService) SetMapImage(id, callerID uuid.UUID, imageURL string) (models.GameMap, error) {
	args := m.Called(id, callerID, imageURL)
	return args.Get(0).(models.GameMap), args.Error(1)
}

func (m *MockMapService) DeleteMapImage(id, callerID uuid.UUID) (models.GameMap, error) {
	args := m.Called(id, callerID)
	return args.Get(0).(models.GameMap), args.Error(1)
}

func (m *MockMapService) CleanupExpiredMaps() error {
	return m.Called().Error(0)
}
