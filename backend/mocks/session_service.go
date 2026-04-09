package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockSessionService struct{ mock.Mock }

func (m *MockSessionService) CreateSession(gameID, callerID uuid.UUID, session *models.Session) (models.Session, error) {
	args := m.Called(gameID, callerID, session)
	return args.Get(0).(models.Session), args.Error(1)
}

func (m *MockSessionService) ListGameSessions(gameID, callerID uuid.UUID) ([]models.Session, error) {
	args := m.Called(gameID, callerID)
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockSessionService) ListGameSessionsPaginated(gameID, callerID uuid.UUID, offset, limit int) ([]models.Session, int64, error) {
	args := m.Called(gameID, callerID, offset, limit)
	return args.Get(0).([]models.Session), args.Get(1).(int64), args.Error(2)
}

func (m *MockSessionService) GetSession(id, callerID uuid.UUID) (models.Session, error) {
	args := m.Called(id, callerID)
	return args.Get(0).(models.Session), args.Error(1)
}

func (m *MockSessionService) UpdateSession(id, callerID uuid.UUID, updates map[string]interface{}) (models.Session, error) {
	args := m.Called(id, callerID, updates)
	return args.Get(0).(models.Session), args.Error(1)
}

func (m *MockSessionService) DeleteSession(id, callerID uuid.UUID) error {
	return m.Called(id, callerID).Error(0)
}
