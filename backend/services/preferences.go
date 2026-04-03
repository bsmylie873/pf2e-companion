package services

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

var validPinColours = map[string]bool{
	"grey": true, "red": true, "orange": true, "gold": true,
	"green": true, "blue": true, "purple": true, "brown": true,
}

var validPinIcons = map[string]bool{
	"position-marker": true, "castle": true, "crossed-swords": true,
	"skull": true, "treasure-map": true, "campfire": true, "forest-camp": true,
	"mountain-cave": true, "village": true, "temple-gate": true, "sailboat": true,
	"crown": true, "dragon-head": true, "tombstone": true, "bridge": true,
	"mine-entrance": true, "tower-flag": true, "cauldron": true,
	"wood-cabin": true, "portal": true,
}

var validMapEditorModes = map[string]bool{
	"modal": true, "navigate": true,
}

type PreferenceService interface {
	GetPreferences(userID uuid.UUID) (models.UserPreferenceResponse, error)
	UpdatePreferences(userID uuid.UUID, updates map[string]interface{}) (models.UserPreferenceResponse, error)
	ClearDefaultGameForMembership(userID, gameID uuid.UUID) error
}

type preferenceService struct {
	repo           repositories.PreferenceRepository
	membershipRepo repositories.MembershipRepository
}

func NewPreferenceService(repo repositories.PreferenceRepository, membershipRepo repositories.MembershipRepository) PreferenceService {
	return &preferenceService{repo: repo, membershipRepo: membershipRepo}
}

func (s *preferenceService) GetPreferences(userID uuid.UUID) (models.UserPreferenceResponse, error) {
	pref, err := s.repo.FindByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.UserPreferenceResponse{}, nil
		}
		return models.UserPreferenceResponse{}, err
	}
	return models.FromUserPreference(pref), nil
}

func (s *preferenceService) UpdatePreferences(userID uuid.UUID, updates map[string]interface{}) (models.UserPreferenceResponse, error) {
	// Load or initialise
	pref, err := s.repo.FindByUserID(userID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return models.UserPreferenceResponse{}, err
	}
	if err == gorm.ErrRecordNotFound {
		pref = models.UserPreference{UserID: userID}
	}

	if v, ok := updates["default_game_id"]; ok {
		if v == nil {
			pref.DefaultGameID = nil
		} else {
			var gameID uuid.UUID
			switch val := v.(type) {
			case string:
				gameID, err = uuid.Parse(val)
				if err != nil {
					return models.UserPreferenceResponse{}, fmt.Errorf("%w: invalid UUID for default_game_id", ErrValidation)
				}
			case uuid.UUID:
				gameID = val
			default:
				return models.UserPreferenceResponse{}, fmt.Errorf("%w: invalid type for default_game_id", ErrValidation)
			}
			// Verify membership
			if _, merr := s.membershipRepo.FindByUserAndGameID(userID, gameID); merr != nil {
				return models.UserPreferenceResponse{}, fmt.Errorf("%w: user is not a member of this game", ErrValidation)
			}
			pref.DefaultGameID = &gameID
		}
	}

	if v, ok := updates["default_pin_colour"]; ok {
		if v == nil {
			pref.DefaultPinColour = nil
		} else {
			colour, ok2 := v.(string)
			if !ok2 || !validPinColours[colour] {
				return models.UserPreferenceResponse{}, fmt.Errorf("%w: invalid default_pin_colour", ErrValidation)
			}
			pref.DefaultPinColour = &colour
		}
	}

	if v, ok := updates["default_pin_icon"]; ok {
		if v == nil {
			pref.DefaultPinIcon = nil
		} else {
			icon, ok2 := v.(string)
			if !ok2 || !validPinIcons[icon] {
				return models.UserPreferenceResponse{}, fmt.Errorf("%w: invalid default_pin_icon", ErrValidation)
			}
			pref.DefaultPinIcon = &icon
		}
	}

	if v, ok := updates["sidebar_state"]; ok {
		if v == nil {
			pref.SidebarState = nil
		} else {
			raw, err := json.Marshal(v)
			if err != nil {
				return models.UserPreferenceResponse{}, fmt.Errorf("%w: invalid sidebar_state", ErrValidation)
			}
			pref.SidebarState = datatypes.JSON(raw)
		}
	}

	if v, ok := updates["default_view_mode"]; ok {
		if v == nil {
			pref.DefaultViewMode = nil
		} else {
			raw, err := json.Marshal(v)
			if err != nil {
				return models.UserPreferenceResponse{}, fmt.Errorf("%w: invalid default_view_mode", ErrValidation)
			}
			pref.DefaultViewMode = datatypes.JSON(raw)
		}
	}

	if v, ok := updates["map_editor_mode"]; ok {
		mode, ok2 := v.(string)
		if !ok2 || !validMapEditorModes[mode] {
			return models.UserPreferenceResponse{}, fmt.Errorf("%w: invalid map_editor_mode, must be 'modal' or 'navigate'", ErrValidation)
		}
		pref.MapEditorMode = mode
	}

	if v, ok := updates["page_size"]; ok {
		if v == nil {
			pref.PageSize = nil
		} else {
			raw, err := json.Marshal(v)
			if err != nil {
				return models.UserPreferenceResponse{}, fmt.Errorf("%w: invalid page_size", ErrValidation)
			}
			pref.PageSize = datatypes.JSON(raw)
		}
	}

	if err := s.repo.Upsert(&pref); err != nil {
		return models.UserPreferenceResponse{}, err
	}
	return models.FromUserPreference(pref), nil
}

func (s *preferenceService) ClearDefaultGameForMembership(userID, gameID uuid.UUID) error {
	return s.repo.ClearDefaultGameForMembership(userID, gameID)
}
