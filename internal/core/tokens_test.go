package core

import (
	"context"
	"errors"
	"inbox451/internal/test"
	"io"
	"testing"
	"time"

	"inbox451/internal/test"

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
	testUserID1 := test.RandomTestUUID()
	testUserID2 := test.RandomTestUUID()
	testTokenID1 := test.RandomTestUUID()
	testTokenID2 := test.RandomTestUUID()
	tests := []struct {
		name    string
		userID  string
		limit   int
		offset  int
		mockFn  func(*mocks.Repository)
		want    *models.PaginatedResponse
		wantErr bool
	}{
		{
			name:   "successful list",
			userID: testUserID1,
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				tokens := []*models.Token{
					{
						Base:   models.Base{ID: testTokenID1},
						UserID: testUserID1,
						Name:   "Token 1",
						Token:  "token1",
					},
					{
						Base:   models.Base{ID: testTokenID2},
						UserID: testUserID1,
						Name:   "Token 2",
						Token:  "token2",
					},
				}
				m.On("ListTokensByUser", mock.Anything, testUserID1, 10, 0).Return(tokens, 2, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.Token{
					{
						Base:   models.Base{ID: testTokenID1},
						UserID: testUserID1,
						Name:   "Token 1",
						Token:  "token1",
					},
					{
						Base:   models.Base{ID: testTokenID2},
						UserID: testUserID1,
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
			userID: testUserID1,
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListTokensByUser", mock.Anything, testUserID1, 10, 0).
					Return([]*models.Token(nil), 0, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "empty token list",
			userID: testUserID2,
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListTokensByUser", mock.Anything, testUserID2, 10, 0).
					Return([]*models.Token{}, 0, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.Token{},
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

			got, err := core.TokenService.ListByUser(context.Background(), tt.userID, tt.limit, tt.offset)
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
	testUserID1 := test.RandomTestUUID()
	testTokenID1 := test.RandomTestUUID()
	nonExistingTokenID := test.RandomTestUUID()
	now := time.Now()

	tests := []struct {
		name    string
		tokenID string
		userID  string
		mockFn  func(*mocks.Repository)
		want    *models.Token
		wantErr bool
		errType error
	}{
		{
			name:    "existing token",
			tokenID: testTokenID1,
			userID:  testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetTokenByUser", mock.Anything, testTokenID1, testUserID1).Return(&models.Token{
					Base: models.Base{
						ID:        testTokenID1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					UserID: testUserID1,
					Name:   "Test Token",
					Token:  "test-token",
				}, nil)
			},
			want: &models.Token{
				Base: models.Base{
					ID:        testTokenID1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				UserID: testUserID1,
				Name:   "Test Token",
				Token:  "test-token",
			},
			wantErr: false,
		},
		{
			name:    "non-existent token",
			tokenID: nonExistingTokenID,
			userID:  testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetTokenByUser", mock.Anything, nonExistingTokenID, testUserID1).Return(nil, storage.ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errType: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTokenTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.TokenService.GetByUser(context.Background(), tt.tokenID, tt.userID)
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
	testUserID1 := test.RandomTestUUID()
	tests := []struct {
		name      string
		userID    string
		tokenData *models.Token
		mockFn    func(*mocks.Repository)
		wantErr   bool
	}{
		{
			name:   "successful creation with default name",
			userID: testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("CreateToken", mock.Anything, mock.MatchedBy(func(token *models.Token) bool {
					return token.UserID == testUserID1 &&
						token.Name == "API Token" &&
						len(token.Token) > 0 // Token should be generated
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "successful creation with custom name",
			userID: testUserID1,
			tokenData: &models.Token{
				Name: "Custom Token",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateToken", mock.Anything, mock.MatchedBy(func(token *models.Token) bool {
					return token.UserID == testUserID1 &&
						token.Name == "Custom Token" &&
						len(token.Token) > 0
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("CreateToken", mock.Anything, mock.Anything).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTokenTestCore(t)
			tt.mockFn(mockRepo)

			token, err := core.TokenService.CreateForUser(context.Background(), tt.userID, tt.tokenData)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.NotEmpty(t, token.Token)
				assert.Equal(t, tt.userID, token.UserID)
				if tt.tokenData != nil && tt.tokenData.Name != "" {
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
	testUserID1 := test.RandomTestUUID()
	testTokenID1 := test.RandomTestUUID()
	nonExistingTokenID := test.RandomTestUUID()
	tests := []struct {
		name    string
		tokenID string
		userID  string
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name:    "successful deletion",
			tokenID: testTokenID1,
			userID:  testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetTokenByUser", mock.Anything, testUserID1, testTokenID1).
					Return(&models.Token{Base: models.Base{ID: testTokenID1}, UserID: testUserID1}, nil)
				m.On("DeleteToken", mock.Anything, testTokenID1).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "token not found",
			tokenID: nonExistingTokenID,
			userID:  testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetTokenByUser", mock.Anything, testUserID1, nonExistingTokenID).
					Return(nil, storage.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name:    "delete error",
			tokenID: testTokenID1,
			userID:  testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetTokenByUser", mock.Anything, testUserID1, testTokenID1).
					Return(&models.Token{Base: models.Base{ID: testTokenID1}, UserID: testUserID1}, nil)
				m.On("DeleteToken", mock.Anything, testTokenID1).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupTokenTestCore(t)
			tt.mockFn(mockRepo)

			err := core.TokenService.DeleteByUser(context.Background(), tt.userID, tt.tokenID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
