package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockFolderRepository struct{ mock.Mock }

func (m *MockFolderRepository) Create(folder *models.Folder) error {
	return m.Called(folder).Error(0)
}

func (m *MockFolderRepository) FindByID(id uuid.UUID) (models.Folder, error) {
	args := m.Called(id)
	return args.Get(0).(models.Folder), args.Error(1)
}

func (m *MockFolderRepository) FindSessionFolders(gameID uuid.UUID) ([]models.Folder, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.Folder), args.Error(1)
}

func (m *MockFolderRepository) FindNoteFolders(gameID, userID uuid.UUID) ([]models.Folder, error) {
	args := m.Called(gameID, userID)
	return args.Get(0).([]models.Folder), args.Error(1)
}

func (m *MockFolderRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Folder, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.Folder), args.Error(1)
}

func (m *MockFolderRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockFolderRepository) BatchUpdatePositions(ids []uuid.UUID, positions []int) error {
	return m.Called(ids, positions).Error(0)
}

func (m *MockFolderRepository) MaxPosition(gameID uuid.UUID, folderType string, parentID *uuid.UUID) (int, error) {
	args := m.Called(gameID, folderType, parentID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockFolderRepository) FindAllByGameID(gameID uuid.UUID) ([]models.Folder, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.Folder), args.Error(1)
}
