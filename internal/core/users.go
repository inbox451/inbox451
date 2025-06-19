package core

import (
	"context"
	"errors"
	"fmt"

	"inbox451/internal/storage"

	"github.com/volatiletech/null/v9"

	"inbox451/internal/models"
)

type UserService struct {
	core *Core
}

func NewUserService(core *Core) UserService {
	return UserService{core: core}
}

func (s *UserService) Create(ctx context.Context, user *models.User) error {
	s.core.Logger.Info("Creating new user: %s", user.Name)

	// Hash the password before saving
	if user.Password.Valid && user.Password.String != "" {
		if err := user.HashPassword(user.Password.String); err != nil {
			s.core.Logger.Error("Failed to hash password for user %s: %v", user.Username, err)
			return fmt.Errorf("password hashing failed: %w", err)
		}
	} else {
		// Ensure password is null if empty or invalid
		user.Password = null.StringFromPtr(nil)
	}

	// Set default status and role if not provided
	if user.Status == "" {
		user.Status = "active" // Or 'inactive' depending on desired default
	}
	if user.Role == "" {
		user.Role = "user"
	}

	if err := s.core.Repository.CreateUser(ctx, user); err != nil {
		s.core.Logger.Error("Failed to create user: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully created user with ID: %d", user.ID)
	return nil
}

func (s *UserService) Get(ctx context.Context, userID string) (*models.User, error) {
	s.core.Logger.Debug("Fetching user with ID: %s", userID)

	user, err := s.core.Repository.GetUser(ctx, userID)
	if err != nil {
		s.core.Logger.Error("Failed to fetch user: %v", err)
		return nil, err
	}

	if user == nil {
		s.core.Logger.Info("User not found with ID: %s", userID)
		return nil, ErrNotFound
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, user *models.User) error {
	s.core.Logger.Info("Updating user with ID: %s", user.ID)

	// Hash the password ONLY if it's being changed (i.e., not empty)
	if user.Password.Valid && user.Password.String != "" {
		if err := user.HashPassword(user.Password.String); err != nil {
			s.core.Logger.Error("Failed to hash password during update for user %s: %v", user.ID, err)
			return fmt.Errorf("password hashing failed: %w", err)
		}
	}

	// If user.Password.String is empty, the repository layer should handle not updating it.
	if err := s.core.Repository.UpdateUser(ctx, user); err != nil {
		s.core.Logger.Error("Failed to update user: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully updated user with ID: %s", user.ID)
	return nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	s.core.Logger.Info("Deleting user with ID: %s", id)

	if err := s.core.Repository.DeleteUser(ctx, id); err != nil {
		s.core.Logger.Error("Failed to delete user: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully deleted user with ID: %s", id)
	return nil
}

func (s *UserService) List(ctx context.Context, limit, offset int) (*models.PaginatedResponse, error) {
	s.core.Logger.Info("Listing users with limit: %d and offset: %d", limit, offset)

	users, total, err := s.core.Repository.ListUsers(ctx, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list users: %v", err)
		return nil, err
	}

	response := &models.PaginatedResponse{
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

// LoginWithPassword validates user credentials.
func (s *UserService) LoginWithPassword(ctx context.Context, username, password string) (*models.User, error) {
	s.core.Logger.Info("Attempting login for username: %s", username)
	user, err := s.core.Repository.GetUserByUsername(ctx, username)
	if err != nil {
		// If user not found or other DB error
		if errors.Is(err, storage.ErrNotFound) {
			s.core.Logger.Warn("Login failed: User not found for username: %s", username)
			return nil, ErrAuthFailed // Use a specific auth error
		}
		s.core.Logger.Error("Database error during login for username %s: %v", username, err)
		return nil, err // Return the original DB error for logging/debugging
	}

	// Check if user is active and allows password login
	if user.Status != "active" {
		s.core.Logger.Warn("Login failed: User account is inactive for username: %s", username)
		return nil, ErrAccountInactive
	}
	if !user.PasswordLogin {
		s.core.Logger.Warn("Login failed: Password login disabled for username: %s", username)
		return nil, ErrPasswordLoginDisabled
	}
	// Check password
	match, err := user.CheckPassword(password)
	if err != nil {
		s.core.Logger.Error("Error checking password for username %s: %v", username, err)
		return nil, fmt.Errorf("error during password verification: %w", err)
	}
	if !match {
		s.core.Logger.Warn("Login failed: Invalid password for username: %s", username)
		return nil, ErrAuthFailed // Specific error for bad credentials
	}

	// Login successful
	s.core.Logger.Info("Login successful for username: %s", username)
	// Optionally: Update last login time here if needed
	// s.core.Repository.UpdateUserLoginTimestamp(ctx, user.ID)
	return user, nil
}

func (s *UserService) LoginWithToken(ctx context.Context, username, tokenValue string) (*models.User, error) {
	s.core.Logger.Info("Attempting login for username: %s", username)
	user, err := s.core.Repository.GetUserByUsername(ctx, username)

	if err != nil {
		// If user not found or other DB error
		if errors.Is(err, storage.ErrNotFound) {
			s.core.Logger.Warn("Login failed: User not found for username: %s", username)
			return nil, ErrAuthFailed
		}
		s.core.Logger.Error("Database error during login for username %s: %v", username, err)
		return nil, err
	}

	// Check if user is active and allows password login
	if user.Status != "active" {
		s.core.Logger.Warn("Login failed: User account is inactive for username: %s", username)
		return nil, ErrAccountInactive
	}

	token, err := s.core.TokenService.GetByValue(ctx, tokenValue)
	if err != nil {
		s.core.Logger.Error("Error retrieving token for username %s: %v", username, err)
		if errors.Is(err, storage.ErrNotFound) {
			s.core.Logger.Warn("Login failed: Token not found for username: %s", username)
			return nil, ErrAuthFailed
		}
	}

	if token == nil || (token.UserID != user.ID) {
		s.core.Logger.Warn("Login failed: Token does not match user for username: %s", username)
		return nil, ErrAuthFailed
	}

	s.core.Logger.Info("Login successful for username: %s with token", username)
	return user, nil
}
