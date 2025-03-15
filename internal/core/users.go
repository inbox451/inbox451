package core

import (
	"inbox451/internal/models"
)

type UserService struct {
	core *Core
}

func NewUserService(core *Core) UserService {
	return UserService{core: core}
}

func (s *UserService) Create(user models.User) (models.User, error) {
	s.core.Logger.Info("Creating new user: %s", user.Name)

	user, err := s.core.Repository.CreateUser(user)
	if err != nil {
		s.core.Logger.Error("Failed to create user: %v", err)
		return user, err
	}

	s.core.Logger.Info("Successfully created user with ID: %d", user.ID)
	return user, nil
}

func (s *UserService) Get(userID int) (models.User, error) {
	s.core.Logger.Debug("Fetching user with ID: %d", userID)

	user, err := s.core.Repository.GetUser(userID)
	if err != nil {
		s.core.Logger.Error("Failed to fetch user: %v", err)
		return user, err
	}

	return user, nil
}

func (s *UserService) Update(user models.User) (models.User, error) {
	s.core.Logger.Info("Updating user with ID: %d", user.ID)

	update, err := s.core.Repository.UpdateUser(user)
	if err != nil {
		s.core.Logger.Error("Failed to update user: %v", err)
		return update, err
	}

	s.core.Logger.Info("Successfully updated user with ID: %d", user.ID)
	return update, nil
}

func (s *UserService) Delete(id int) error {
	s.core.Logger.Info("Deleting user with ID: %d", id)

	if err := s.core.Repository.DeleteUser(id); err != nil {
		s.core.Logger.Error("Failed to delete user: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully deleted user with ID: %d", id)
	return nil
}

func (s *UserService) List(limit, offset int) (models.PaginatedResponse, error) {
	s.core.Logger.Info("Listing users with limit: %d and offset: %d", limit, offset)

	users, total, err := s.core.Repository.ListUsers(limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list users: %v", err)
		return models.PaginatedResponse{}, err
	}

	response := models.PaginatedResponse{
		Data: users,
		Pagination: models.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	s.core.Logger.Info("Successfully retrieved %d users (total: %d)", len(users), total)
	return response, nil
}
