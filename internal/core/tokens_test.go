package core

import (
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

func setupTokenTestCore(t *testing.T) (*Core, *mocks.Repository) {
	mockRepo := mocks.NewRepository(t)
	logger := logger.New(io.Discard, logger.DEBUG)

	core := &Core{
		Logger:     logger,
		Repository: mockRepo,
	}
	core.TokenService = NewTokensService(core)

	return core, mockRepo
}

func TestTokenService_ListByUser(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		limit   int
		offset  int
		mockFn  func(*mocks.Repository)
		want    models.PaginatedResponse
		wantErr bool
	}{
		{
			name:   "successful list",
			userID: 1,
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				tokens := []models.Token{
					{
						Base:   models.Base{ID: 1},
						UserID: 1,
						Name:   "Token 1",
						Token:  "token1",
					},
					{
						Base:   models.Base{ID: 2},
						UserID: 1,
						Name:   "Token 2",
						Token:  "token2",
					},
				}
				m.On("ListTokensByUser", 1, 10, 0).Return(tokens, 2, nil)
			},
			want: models.PaginatedResponse{
				Data: []models.Token{
					{
						Base:   models.Base{ID: 1},
						UserID: 1,
						Name:   "Token 1",
						Token:  "token1",
					},
					{
						Base:   models.Base{ID: 2},
						UserID: 1,
						Name:   "Token 2",
						Token:  "token2",
					},
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
			name:   "repository error",
			userID: 1,
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListTokensByUser", 1, 10, 0).
					Return([]models.Token(nil), 0, errors.New("database error"))
			},
			want:    models.PaginatedResponse{},
			wantErr: true,
		},
		{
			name:   "empty token list",
			userID: 2,
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListTokensByUser", 2, 10, 0).
					Return([]models.Token{}, 0, nil)
			},
			want: models.PaginatedResponse{
				Data: []models.Token{},
				Pagination: models.Pagination{
					Total:  0,
					Limit:  10,
					Offset: 0,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTokenTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.TokenService.ListByUser(tt.userID, tt.limit, tt.offset)
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

func TestTokenService_GetByUser(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		tokenID int
		userID  int
		mockFn  func(*mocks.Repository)
		want    models.Token
		wantErr bool
		errType error
	}{
		{
			name:    "existing token",
			tokenID: 1,
			userID:  1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetTokenByUser", 1, 1).Return(models.Token{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					UserID: 1,
					Name:   "Test Token",
					Token:  "test-token",
				}, nil)
			},
			want: models.Token{
				Base: models.Base{
					ID:        1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				UserID: 1,
				Name:   "Test Token",
				Token:  "test-token",
			},
			wantErr: false,
		},
		{
			name:    "non-existent token",
			tokenID: 999,
			userID:  1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetTokenByUser", 999, 1).Return(models.Token{}, storage.ErrNotFound)
			},
			want:    models.Token{},
			wantErr: true,
			errType: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTokenTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.TokenService.GetByUser(tt.tokenID, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTokenService_CreateForUser(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		tokenData models.Token
		mockFn    func(*mocks.Repository)
		wantErr   bool
	}{
		{
			name:   "successful creation with default name",
			userID: 1,
			mockFn: func(m *mocks.Repository) {
				m.On("CreateToken", mock.MatchedBy(func(token models.Token) bool {
					return token.UserID == 1 &&
						token.Name == "API Token" &&
						len(token.Token) > 0 // Token should be generated
				})).Return(models.Token{Base: models.Base{ID: 1}, UserID: 1, Name: "API Token", Token: "token"}, nil)
			},
			wantErr: false,
		},
		{
			name:   "successful creation with custom name",
			userID: 1,
			tokenData: models.Token{
				Name: "Custom Token",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateToken", mock.MatchedBy(func(token *models.Token) bool {
					return token.UserID == 1 &&
						token.Name == "Custom Token" &&
						len(token.Token) > 0
				})).Return(models.Token{Base: models.Base{ID: 1}, UserID: 1, Name: "Custom Token", Token: "token"}, nil)
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: 1,
			mockFn: func(m *mocks.Repository) {
				m.On("CreateToken", mock.Anything).
					Return(models.Token{}, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTokenTestCore(t)
			tt.mockFn(mockRepo)

			token, err := core.TokenService.CreateForUser(tt.userID, tt.tokenData)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.NotEmpty(t, token.Token)
				assert.Equal(t, tt.userID, token.UserID)
				if tt.tokenData.Name != "" {
					assert.Equal(t, tt.tokenData.Name, token.Name)
				} else {
					assert.Equal(t, "API Token", token.Name)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTokenService_DeleteByUser(t *testing.T) {
	tests := []struct {
		name    string
		tokenID int
		userID  int
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name:    "successful deletion",
			tokenID: 1,
			userID:  1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetTokenByUser", 1, 1).
					Return(&models.Token{Base: models.Base{ID: 1}, UserID: 1}, nil)
				m.On("DeleteToken", 1).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "token not found",
			tokenID: 999,
			userID:  1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetTokenByUser", 1, 999).
					Return(models.Token{}, storage.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name:    "delete error",
			tokenID: 1,
			userID:  1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetTokenByUser", 1, 1).
					Return(&models.Token{Base: models.Base{ID: 1}, UserID: 1}, nil)
				m.On("DeleteToken", 1).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTokenTestCore(t)
			tt.mockFn(mockRepo)

			err := core.TokenService.DeleteByUser(tt.userID, tt.tokenID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
