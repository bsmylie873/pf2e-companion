package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockPinGroupService struct{ mock.Mock }

func (m *MockPinGroupService) CreateGroup(gameID, callerID uuid.UUID, pinIDs []uuid.UUID) (models.PinGroupResponse, error) {
	args := m.Called(gameID, callerID, pinIDs)
	return args.Get(0).(models.PinGroupResponse), args.Error(1)
}

func (m *MockPinGroupService) GetGroup(id, callerID uuid.UUID) (models.PinGroupResponse, error) {
	args := m.Called(id, callerID)
	return args.Get(0).(models.PinGroupResponse), args.Error(1)
}

func (m *MockPinGroupService) AddPinToGroup(groupID, pinID, callerID uuid.UUID) (models.PinGroupResponse, error) {
	args := m.Called(groupID, pinID, callerID)
	return args.Get(0).(models.PinGroupResponse), args.Error(1)
}

func (m *MockPinGroupService) RemovePinFromGroup(groupID, pinID, callerID uuid.UUID) (models.PinGroupResponse, error) {
	args := m.Called(groupID, pinID, callerID)
	return args.Get(0).(models.PinGroupResponse), args.Error(1)
}

func (m *MockPinGroupService) UpdateGroup(id, callerID uuid.UUID, updates map[string]interface{}) (models.PinGroupResponse, error) {
	args := m.Called(id, callerID, updates)
	return args.Get(0).(models.PinGroupResponse), args.Error(1)
}

func (m *MockPinGroupService) DisbandGroup(id, callerID uuid.UUID) error {
	return m.Called(id, callerID).Error(0)
}

func (m *MockPinGroupService) ListGameGroups(gameID, callerID uuid.UUID) ([]models.PinGroupResponse, error) {
	args := m.Called(gameID, callerID)
	return args.Get(0).([]models.PinGroupResponse), args.Error(1)
}

func (m *MockPinGroupService) CreateMapGroup(mapID, callerID uuid.UUID, pinIDs []uuid.UUID) (models.PinGroupResponse, error) {
	args := m.Called(mapID, callerID, pinIDs)
	return args.Get(0).(models.PinGroupResponse), args.Error(1)
}

func (m *MockPinGroupService) ListMapGroups(mapID, callerID uuid.UUID) ([]models.PinGroupResponse, error) {
	args := m.Called(mapID, callerID)
	return args.Get(0).([]models.PinGroupResponse), args.Error(1)
}
