package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type PinGroupService interface {
	CreateGroup(gameID, userID uuid.UUID, pinIDs []uuid.UUID) (models.PinGroupResponse, error)
	GetGroup(groupID, userID uuid.UUID) (models.PinGroupResponse, error)
	AddPinToGroup(groupID, pinID, userID uuid.UUID) (models.PinGroupResponse, error)
	RemovePinFromGroup(groupID, pinID, userID uuid.UUID) (models.PinGroupResponse, error)
	UpdateGroup(groupID, userID uuid.UUID, updates map[string]interface{}) (models.PinGroupResponse, error)
	DisbandGroup(groupID, userID uuid.UUID) error
	ListGameGroups(gameID, userID uuid.UUID) ([]models.PinGroupResponse, error)
	CreateMapGroup(mapID, userID uuid.UUID, pinIDs []uuid.UUID) (models.PinGroupResponse, error)
	ListMapGroups(mapID, userID uuid.UUID) ([]models.PinGroupResponse, error)
}

type pinGroupService struct {
	pinGroupRepo   repositories.PinGroupRepository
	pinRepo        repositories.PinRepository
	noteRepo       repositories.NoteRepository
	membershipRepo repositories.MembershipRepository
	mapRepo        repositories.MapRepository
}

func NewPinGroupService(
	pinGroupRepo repositories.PinGroupRepository,
	pinRepo repositories.PinRepository,
	noteRepo repositories.NoteRepository,
	membershipRepo repositories.MembershipRepository,
	mapRepo repositories.MapRepository,
) PinGroupService {
	return &pinGroupService{
		pinGroupRepo:   pinGroupRepo,
		pinRepo:        pinRepo,
		noteRepo:       noteRepo,
		membershipRepo: membershipRepo,
		mapRepo:        mapRepo,
	}
}

func (s *pinGroupService) buildGroupResponse(group models.PinGroup) (models.PinGroupResponse, error) {
	pins, err := s.pinRepo.FindByGroupID(group.ID)
	if err != nil {
		return models.PinGroupResponse{}, err
	}
	return models.PinGroupResponse{
		ID:        group.ID,
		GameID:    group.GameID,
		X:         group.X,
		Y:         group.Y,
		Colour:    group.Colour,
		Icon:      group.Icon,
		PinCount:  len(pins),
		Pins:      pins,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}, nil
}

func (s *pinGroupService) GetGroup(groupID, userID uuid.UUID) (models.PinGroupResponse, error) {
	group, err := s.pinGroupRepo.FindByID(groupID)
	if err != nil {
		return models.PinGroupResponse{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, group.GameID); err != nil {
		return models.PinGroupResponse{}, ErrForbidden
	}
	return s.buildGroupResponse(group)
}

func (s *pinGroupService) CreateGroup(gameID, userID uuid.UUID, pinIDs []uuid.UUID) (models.PinGroupResponse, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return models.PinGroupResponse{}, ErrForbidden
	}
	if len(pinIDs) < 2 {
		return models.PinGroupResponse{}, errors.New("at least 2 pins required to create a group")
	}
	// Validate all pins
	pins := make([]models.SessionPin, 0, len(pinIDs))
	for _, pid := range pinIDs {
		pin, err := s.pinRepo.FindByID(pid)
		if err != nil {
			return models.PinGroupResponse{}, err
		}
		if pin.GameID != gameID {
			return models.PinGroupResponse{}, ErrForbidden
		}
		if pin.GroupID != nil {
			return models.PinGroupResponse{}, errors.New("pin " + pid.String() + " is already in a group")
		}
		pins = append(pins, pin)
	}
	// Inherit x/y/colour/icon from first pin
	first := pins[0]
	group := &models.PinGroup{
		GameID: gameID,
		X:      first.X,
		Y:      first.Y,
		Colour: first.Colour,
		Icon:   first.Icon,
	}
	if err := s.pinGroupRepo.Create(group); err != nil {
		return models.PinGroupResponse{}, err
	}
	for _, pid := range pinIDs {
		if err := s.pinRepo.SetGroupID(pid, group.ID); err != nil {
			return models.PinGroupResponse{}, err
		}
	}
	return s.buildGroupResponse(*group)
}

func (s *pinGroupService) AddPinToGroup(groupID, pinID, userID uuid.UUID) (models.PinGroupResponse, error) {
	group, err := s.pinGroupRepo.FindByID(groupID)
	if err != nil {
		return models.PinGroupResponse{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, group.GameID); err != nil {
		return models.PinGroupResponse{}, ErrForbidden
	}
	pin, err := s.pinRepo.FindByID(pinID)
	if err != nil {
		return models.PinGroupResponse{}, err
	}
	if pin.GameID != group.GameID {
		return models.PinGroupResponse{}, ErrForbidden
	}
	if pin.GroupID != nil {
		return models.PinGroupResponse{}, errors.New("pin " + pinID.String() + " is already in a group")
	}
	if err := s.pinRepo.SetGroupID(pinID, groupID); err != nil {
		return models.PinGroupResponse{}, err
	}
	return s.buildGroupResponse(group)
}

func (s *pinGroupService) RemovePinFromGroup(groupID, pinID, userID uuid.UUID) (models.PinGroupResponse, error) {
	group, err := s.pinGroupRepo.FindByID(groupID)
	if err != nil {
		return models.PinGroupResponse{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, group.GameID); err != nil {
		return models.PinGroupResponse{}, ErrForbidden
	}
	if err := s.pinRepo.ClearGroupID(pinID); err != nil {
		return models.PinGroupResponse{}, err
	}
	// Auto-dissolve if <= 1 member remains
	count, err := s.pinGroupRepo.CountMembers(groupID)
	if err != nil {
		return models.PinGroupResponse{}, err
	}
	if count <= 1 {
		if count == 1 {
			remaining, err := s.pinRepo.FindByGroupID(groupID)
			if err != nil {
				return models.PinGroupResponse{}, err
			}
			for _, p := range remaining {
				if err := s.pinRepo.ClearGroupID(p.ID); err != nil {
					return models.PinGroupResponse{}, err
				}
			}
		}
		if err := s.pinGroupRepo.Delete(groupID); err != nil {
			return models.PinGroupResponse{}, err
		}
		return models.PinGroupResponse{}, nil
	}
	return s.buildGroupResponse(group)
}

func (s *pinGroupService) UpdateGroup(groupID, userID uuid.UUID, updates map[string]interface{}) (models.PinGroupResponse, error) {
	group, err := s.pinGroupRepo.FindByID(groupID)
	if err != nil {
		return models.PinGroupResponse{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, group.GameID); err != nil {
		return models.PinGroupResponse{}, ErrForbidden
	}
	delete(updates, "id")
	delete(updates, "game_id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	updatedGroup, err := s.pinGroupRepo.Update(groupID, updates)
	if err != nil {
		return models.PinGroupResponse{}, err
	}
	return s.buildGroupResponse(updatedGroup)
}

func (s *pinGroupService) DisbandGroup(groupID, userID uuid.UUID) error {
	group, err := s.pinGroupRepo.FindByID(groupID)
	if err != nil {
		return err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, group.GameID); err != nil {
		return ErrForbidden
	}
	pins, err := s.pinRepo.FindByGroupID(groupID)
	if err != nil {
		return err
	}
	for _, p := range pins {
		if err := s.pinRepo.ClearGroupID(p.ID); err != nil {
			return err
		}
	}
	return s.pinGroupRepo.Delete(groupID)
}

func (s *pinGroupService) ListGameGroups(gameID, userID uuid.UUID) ([]models.PinGroupResponse, error) {
	membership, err := s.membershipRepo.FindByUserAndGameID(userID, gameID)
	if err != nil {
		return nil, ErrForbidden
	}
	groups, err := s.pinGroupRepo.FindByGameID(gameID)
	if err != nil {
		return nil, err
	}
	result := make([]models.PinGroupResponse, 0, len(groups))
	for _, g := range groups {
		pins, err := s.pinRepo.FindByGroupID(g.ID)
		if err != nil {
			return nil, err
		}
		if !membership.IsGM {
			filtered := make([]models.SessionPin, 0, len(pins))
			for _, p := range pins {
				if p.NoteID != nil {
					note, err := s.noteRepo.FindByID(*p.NoteID)
					if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, err
					}
					if err == nil && note.Visibility == "private" && note.UserID != userID {
						continue
					}
				}
				filtered = append(filtered, p)
			}
			pins = filtered
		}
		result = append(result, models.PinGroupResponse{
			ID:        g.ID,
			GameID:    g.GameID,
			X:         g.X,
			Y:         g.Y,
			Colour:    g.Colour,
			Icon:      g.Icon,
			PinCount:  len(pins),
			Pins:      pins,
			CreatedAt: g.CreatedAt,
			UpdatedAt: g.UpdatedAt,
		})
	}
	return result, nil
}

func (s *pinGroupService) CreateMapGroup(mapID, userID uuid.UUID, pinIDs []uuid.UUID) (models.PinGroupResponse, error) {
	m, err := s.mapRepo.FindByID(mapID)
	if err != nil {
		return models.PinGroupResponse{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, m.GameID); err != nil {
		return models.PinGroupResponse{}, ErrForbidden
	}
	if len(pinIDs) < 2 {
		return models.PinGroupResponse{}, errors.New("at least 2 pins required to create a group")
	}
	pins := make([]models.SessionPin, 0, len(pinIDs))
	for _, pid := range pinIDs {
		pin, err := s.pinRepo.FindByID(pid)
		if err != nil {
			return models.PinGroupResponse{}, err
		}
		if pin.GameID != m.GameID {
			return models.PinGroupResponse{}, ErrForbidden
		}
		if pin.GroupID != nil {
			return models.PinGroupResponse{}, errors.New("pin " + pid.String() + " is already in a group")
		}
		pins = append(pins, pin)
	}
	first := pins[0]
	group := &models.PinGroup{
		GameID: m.GameID,
		MapID:  &mapID,
		X:      first.X,
		Y:      first.Y,
		Colour: first.Colour,
		Icon:   first.Icon,
	}
	if err := s.pinGroupRepo.Create(group); err != nil {
		return models.PinGroupResponse{}, err
	}
	for _, pid := range pinIDs {
		if err := s.pinRepo.SetGroupID(pid, group.ID); err != nil {
			return models.PinGroupResponse{}, err
		}
	}
	return s.buildGroupResponse(*group)
}

func (s *pinGroupService) ListMapGroups(mapID, userID uuid.UUID) ([]models.PinGroupResponse, error) {
	m, err := s.mapRepo.FindByID(mapID)
	if err != nil {
		return nil, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, m.GameID); err != nil {
		return nil, ErrForbidden
	}
	groups, err := s.pinGroupRepo.FindByMapID(mapID)
	if err != nil {
		return nil, err
	}
	result := make([]models.PinGroupResponse, 0, len(groups))
	for _, g := range groups {
		resp, err := s.buildGroupResponse(g)
		if err != nil {
			return nil, err
		}
		result = append(result, resp)
	}
	return result, nil
}
