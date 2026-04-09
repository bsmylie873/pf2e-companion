package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockFolderService struct{ mock.Mock }

func (m *MockFolderService) CreateFolder(gameID, callerID uuid.UUID, name, folderType, visibility string) (models.Folder, error) {
	args := m.Called(gameID, callerID, name, folderType, visibility)
	return args.Get(0).(models.Folder), args.Error(1)
}

func (m *MockFolderService) ListFolders(gameID, callerID uuid.UUID, folderType string) ([]models.Folder, error) {
	args := m.Called(gameID, callerID, folderType)
	return args.Get(0).([]models.Folder), args.Error(1)
}

func (m *MockFolderService) RenameFolder(id, callerID uuid.UUID, name string) (models.Folder, error) {
	args := m.Called(id, callerID, name)
	return args.Get(0).(models.Folder), args.Error(1)
}

func (m *MockFolderService) DeleteFolder(id, callerID uuid.UUID) error {
	return m.Called(id, callerID).Error(0)
}

func (m *MockFolderService) ReorderFolders(gameID, callerID uuid.UUID, folderType string, orderedIDs []uuid.UUID) error {
	return m.Called(gameID, callerID, folderType, orderedIDs).Error(0)
}
