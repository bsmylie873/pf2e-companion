package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type UserService interface {
	CreateUser(user *models.User) (models.UserResponse, error)
	ListUsers() ([]models.UserResponse, error)
	GetUser(id uuid.UUID) (models.UserResponse, error)
	UpdateUser(id uuid.UUID, updates map[string]interface{}) (models.UserResponse, error)
	DeleteUser(id uuid.UUID) error
}

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(user *models.User) (models.UserResponse, error) {
	if err := s.repo.Create(user); err != nil {
		return models.UserResponse{}, err
	}
	return models.FromUser(*user), nil
}

func (s *userService) ListUsers() ([]models.UserResponse, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	responses := make([]models.UserResponse, len(users))
	for i, u := range users {
		responses[i] = models.FromUser(u)
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

func (s *userService) UpdateUser(id uuid.UUID, updates map[string]interface{}) (models.UserResponse, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		return models.UserResponse{}, err
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

func (s *userService) DeleteUser(id uuid.UUID) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
