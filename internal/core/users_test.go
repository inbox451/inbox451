package core

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"inbox451/internal/logger"
	"inbox451/internal/mocks"
	"inbox451/internal/models"
	"inbox451/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	null "github.com/volatiletech/null/v9"
)

func setupTestCore(t *testing.T) (*Core, *mocks.Repository) {
	mockRepo := mocks.NewRepository(t)
	logger := logger.New(io.Discard, logger.DEBUG)

	core := &Core{
		Logger:     logger,
		Repository: mockRepo,
	}
	core.UserService = NewUserService(core)

	return core, mockRepo
}

func TestUserService_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    *models.User
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful creation",
			user: &models.User{
				Name:     "Test User",
				Username: "testuser",
				Email:    "test@example.com",
				Status:   "active",
				Role:     "user",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Repository error",
			user: &models.User{
				Name:     "Test User",
				Username: "testuser",
				Email:    "test@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTestCore(t)
			tt.mockFn(mockRepo)

			err := core.UserService.Create(context.Background(), tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Get(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		userID  int
		mockFn  func(*mocks.Repository)
		want    *models.User
		wantErr bool
		errType error
	}{
		{
			name:   "existing user",
			userID: 1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetUser", mock.Anything, 1).Return(&models.User{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					Name:     "Test User",
					Username: "testuser",
					Email:    "test@example.com",
				}, nil)
			},
			want: &models.User{
				Base: models.Base{
					ID:        1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				Name:     "Test User",
				Username: "testuser",
				Email:    "test@example.com",
			},
			wantErr: false,
		},
		{
			name:   "non-existent user",
			userID: 999,
			mockFn: func(m *mocks.Repository) {
				m.On("GetUser", mock.Anything, 999).Return(nil, storage.ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errType: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.UserService.Get(context.Background(), tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType, "expected error type does not match")
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_List(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		offset  int
		mockFn  func(*mocks.Repository)
		want    *models.PaginatedResponse
		wantErr bool
	}{
		{
			name:   "successful list",
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				users := []*models.User{
					{Base: models.Base{ID: 1}, Name: "User 1"},
					{Base: models.Base{ID: 2}, Name: "User 2"},
				}
				m.On("ListUsers", mock.Anything, 10, 0).Return(users, 2, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.User{
					{Base: models.Base{ID: 1}, Name: "User 1"},
					{Base: models.Base{ID: 2}, Name: "User 2"},
				},
				Pagination: models.Pagination{
					Total:  2,
					Limit:  10,
					Offset: 0,
				},
			},
			wantErr: false,
		},
		{
			name:   "Repository error",
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListUsers", mock.Anything, 10, 0).Return([]*models.User(nil), 0, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.UserService.List(context.Background(), tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Update(t *testing.T) {
	tests := []struct {
		name    string
		user    *models.User
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful update",
			user: &models.User{
				Base: models.Base{ID: 1},
				Name: "Updated Name",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "update non-existent user",
			user: &models.User{
				Base: models.Base{ID: 999},
				Name: "Updated Name",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateUser", mock.Anything, mock.AnythingOfType("*models.User")).
					Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTestCore(t)
			tt.mockFn(mockRepo)

			err := core.UserService.Update(context.Background(), tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name:   "successful deletion",
			userID: 1,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteUser", mock.Anything, 1).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "delete non-existent user",
			userID: 999,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteUser", mock.Anything, 999).Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTestCore(t)
			tt.mockFn(mockRepo)

			err := core.UserService.Delete(context.Background(), tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_LoginWithPassword(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		mockFn   func(*mocks.Repository)
		want     *models.User
		wantErr  bool
		errType  error
	}{
		{
			name:     "successful login",
			username: "testuser",
			password: "password123",
			mockFn: func(m *mocks.Repository) {
				user := &models.User{
					Base:          models.Base{ID: 1},
					Username:      "testuser",
					Status:        "active",
					Password:      null.StringFrom("$2a$10$LhYtVKJdkcACbejxhtJU8eoDyMNhsntXzXDc3uQVpAXSBSSreyK6i"),
					PasswordLogin: true,
				}
				m.On("GetUserByUsername", mock.Anything, "testuser").Return(user, nil)
			},
			want: &models.User{
				Base:          models.Base{ID: 1},
				Username:      "testuser",
				Status:        "active",
				Password:      null.StringFrom("$2a$10$LhYtVKJdkcACbejxhtJU8eoDyMNhsntXzXDc3uQVpAXSBSSreyK6i"),
				PasswordLogin: true,
			},
			wantErr: false,
		},
		{
			name:     "non-existent user",
			username: "nonexistent",
			password: "password123",
			mockFn: func(m *mocks.Repository) {
				m.On("GetUserByUsername", mock.Anything, "nonexistent").Return((*models.User)(nil), storage.ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errType: ErrAuthFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.UserService.LoginWithPassword(context.Background(), tt.username, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType, "expected error type does not match")
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
