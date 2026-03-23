package services

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type UserService interface {
	ListUsers() ([]models.UserPublicResponse, error)
	GetUser(id uuid.UUID) (models.UserResponse, error)
	UpdateUser(id uuid.UUID, updates map[string]interface{}, callerID uuid.UUID) (models.UserResponse, error)
	DeleteUser(id uuid.UUID, callerID uuid.UUID) error
}

type userService struct {
	repo    repositories.UserRepository
	authSvc AuthService
}

func NewUserService(repo repositories.UserRepository, authSvc AuthService) UserService {
	return &userService{repo: repo, authSvc: authSvc}
}

func (s *userService) ListUsers() ([]models.UserPublicResponse, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	responses := make([]models.UserPublicResponse, len(users))
	for i, u := range users {
		responses[i] = models.FromUserPublic(u)
	}
	return responses, nil
}

func (s *userService) GetUser(id uuid.UUID) (models.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return models.UserResponse{}, err
	}
	return models.FromUser(user), nil
}

func (s *userService) UpdateUser(id uuid.UUID, updates map[string]interface{}, callerID uuid.UUID) (models.UserResponse, error) {
	if id != callerID {
		return models.UserResponse{}, ErrForbidden
	}
	if _, err := s.repo.FindByID(id); err != nil {
		return models.UserResponse{}, err
	}
	if pwRaw, hasPw := updates["password"]; hasPw {
		pw, ok := pwRaw.(string)
		if !ok || pw == "" {
			return models.UserResponse{}, errors.New("invalid password")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			return models.UserResponse{}, err
		}
		updates["password_hash"] = string(hash)
		delete(updates, "password")
		defer s.authSvc.InvalidateAllSessions(id)
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	updated, err := s.repo.Update(id, updates)
	if err != nil {
		return models.UserResponse{}, err
	}
	return models.FromUser(updated), nil
}

func (s *userService) DeleteUser(id uuid.UUID, callerID uuid.UUID) error {
	if id != callerID {
		return ErrForbidden
	}
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
