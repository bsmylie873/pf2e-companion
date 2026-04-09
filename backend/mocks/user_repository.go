package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(user *models.User) error {
	return m.Called(user).Error(0)
}

func (m *MockUserRepository) FindAll() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id uuid.UUID) (models.User, error) {
	args := m.Called(id)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.User, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockUserRepository) FindByEmail(email string) (models.User, error) {
	args := m.Called(email)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (models.User, error) {
	args := m.Called(username)
	return args.Get(0).(models.User), args.Error(1)
}
