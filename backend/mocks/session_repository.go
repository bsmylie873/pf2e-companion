package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockSessionRepository struct{ mock.Mock }

func (m *MockSessionRepository) Create(session *models.Session) error {
	return m.Called(session).Error(0)
}

func (m *MockSessionRepository) FindByGameID(gameID uuid.UUID) ([]models.Session, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.Session), args.Error(1)
}

func (m *MockSessionRepository) FindByGameIDPaginated(gameID uuid.UUID, offset, limit int) ([]models.Session, int64, error) {
	args := m.Called(gameID, offset, limit)
	return args.Get(0).([]models.Session), args.Get(1).(int64), args.Error(2)
}

func (m *MockSessionRepository) FindByID(id uuid.UUID) (models.Session, error) {
	args := m.Called(id)
	return args.Get(0).(models.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Session, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.Session), args.Error(1)
}

func (m *MockSessionRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}
