package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockBackupService struct{ mock.Mock }

func (m *MockBackupService) ExportGame(gameID, callerID uuid.UUID) (*models.BackupFile, error) {
	args := m.Called(gameID, callerID)
	return args.Get(0).(*models.BackupFile), args.Error(1)
}

func (m *MockBackupService) ExportSession(sessionID, callerID uuid.UUID) (*models.BackupFile, error) {
	args := m.Called(sessionID, callerID)
	return args.Get(0).(*models.BackupFile), args.Error(1)
}

func (m *MockBackupService) ExportNote(noteID, callerID uuid.UUID) (*models.BackupFile, error) {
	args := m.Called(noteID, callerID)
	return args.Get(0).(*models.BackupFile), args.Error(1)
}

func (m *MockBackupService) ImportGame(gameID, callerID uuid.UUID, mode string, backup *models.BackupFile) (*models.ImportSummary, error) {
	args := m.Called(gameID, callerID, mode, backup)
	return args.Get(0).(*models.ImportSummary), args.Error(1)
}
