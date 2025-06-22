package core

import (
	"context"
	"errors"
	"inbox451/internal/test"
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

func setupRuleTestCore(t *testing.T) (*Core, *mocks.Repository) {
	mockRepo := mocks.NewRepository(t)
	logger := logger.New(io.Discard, logger.DEBUG)

	core := &Core{
		Logger:     logger,
		Repository: mockRepo,
	}
	core.RuleService = NewRuleService(core)

	return core, mockRepo
}

func TestRuleService_Create(t *testing.T) {
	testInboxID1 := test.RandomTestUUID()
	tests := []struct {
		name    string
		rule    *models.ForwardRule
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful creation",
			rule: &models.ForwardRule{
				InboxID:  testInboxID1,
				Sender:   "sender@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Test Subject",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateRule", mock.Anything, mock.AnythingOfType("*models.ForwardRule")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			rule: &models.ForwardRule{
				InboxID:  testInboxID1,
				Sender:   "sender@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Test Subject",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateRule", mock.Anything, mock.AnythingOfType("*models.ForwardRule")).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupRuleTestCore(t)
			tt.mockFn(mockRepo)

			err := core.RuleService.Create(context.Background(), tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRuleService_Get(t *testing.T) {
	now := time.Now()
	testInboxID1 := test.RandomTestUUID()
	testRuleID1 := test.RandomTestUUID()
	nonExistingRuleID := test.RandomTestUUID()
	tests := []struct {
		name    string
		id      string
		mockFn  func(*mocks.Repository)
		want    *models.ForwardRule
		wantErr bool
		errType error
	}{
		{
			name: "existing rule",
			id:   testRuleID1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetRule", mock.Anything, 1).Return(&models.ForwardRule{
					Base: models.Base{
						ID:        testRuleID1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  testInboxID1,
					Sender:   "sender@example.com",
					Receiver: "receiver@example.com",
					Subject:  "Test Subject",
				}, nil)
			},
			want: &models.ForwardRule{
				Base: models.Base{
					ID:        testRuleID1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				InboxID:  testInboxID1,
				Sender:   "sender@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Test Subject",
			},
			wantErr: false,
		},
		{
			name: "non-existent rule",
			id:   nonExistingRuleID,
			mockFn: func(m *mocks.Repository) {
				m.On("GetRule", mock.Anything, nonExistingRuleID).Return(nil, storage.ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errType: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupRuleTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.RuleService.Get(context.Background(), tt.id)
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

func TestRuleService_Update(t *testing.T) {
	testInboxID1 := test.RandomTestUUID()
	testRuleID1 := test.RandomTestUUID()
	nonExistingRuleID := test.RandomTestUUID()
	tests := []struct {
		name    string
		rule    *models.ForwardRule
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful update",
			rule: &models.ForwardRule{
				Base:     models.Base{ID: testRuleID1},
				InboxID:  testInboxID1,
				Sender:   "updated@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Updated Subject",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateRule", mock.Anything, mock.AnythingOfType("*models.ForwardRule")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "update non-existent rule",
			rule: &models.ForwardRule{
				Base:     models.Base{ID: nonExistingRuleID},
				InboxID:  testInboxID1,
				Sender:   "updated@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Updated Subject",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateRule", mock.Anything, mock.AnythingOfType("*models.ForwardRule")).
					Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupRuleTestCore(t)
			tt.mockFn(mockRepo)

			err := core.RuleService.Update(context.Background(), tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRuleService_Delete(t *testing.T) {
	testRuleID1 := test.RandomTestUUID()
	nonExistingRuleID := test.RandomTestUUID()
	tests := []struct {
		name    string
		id      string
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   testRuleID1,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteRule", mock.Anything, testRuleID1).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "delete non-existent rule",
			id:   nonExistingRuleID,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteRule", mock.Anything, nonExistingRuleID).Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupRuleTestCore(t)
			tt.mockFn(mockRepo)

			err := core.RuleService.Delete(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRuleService_ListByInbox(t *testing.T) {
	testInboxID1 := test.RandomTestUUID()
	testInboxID2 := test.RandomTestUUID()
	testRuleID1 := test.RandomTestUUID()
	testRuleID2 := test.RandomTestUUID()
	tests := []struct {
		name    string
		inboxID string
		limit   int
		offset  int
		mockFn  func(*mocks.Repository)
		want    *models.PaginatedResponse
		wantErr bool
	}{
		{
			name:    "successful list",
			inboxID: testInboxID1,
			limit:   10,
			offset:  0,
			mockFn: func(m *mocks.Repository) {
				rules := []*models.ForwardRule{
					{
						Base:     models.Base{ID: testRuleID1},
						InboxID:  testInboxID1,
						Sender:   "sender1@example.com",
						Receiver: "receiver1@example.com",
						Subject:  "Subject 1",
					},
					{
						Base:     models.Base{ID: testRuleID2},
						InboxID:  testInboxID1,
						Sender:   "sender2@example.com",
						Receiver: "receiver2@example.com",
						Subject:  "Subject 2",
					},
				}
				m.On("ListRulesByInbox", mock.Anything, testInboxID1, 10, 0).Return(rules, 2, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.ForwardRule{
					{
						Base:     models.Base{ID: testRuleID1},
						InboxID:  testInboxID1,
						Sender:   "sender1@example.com",
						Receiver: "receiver1@example.com",
						Subject:  "Subject 1",
					},
					{
						Base:     models.Base{ID: testRuleID2},
						InboxID:  testInboxID1,
						Sender:   "sender2@example.com",
						Receiver: "receiver2@example.com",
						Subject:  "Subject 2",
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
			name:    "repository error",
			inboxID: testInboxID1,
			limit:   10,
			offset:  0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListRulesByInbox", mock.Anything, testInboxID1, 10, 0).
					Return([]*models.ForwardRule(nil), 0, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty inbox",
			inboxID: testInboxID2,
			limit:   10,
			offset:  0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListRulesByInbox", mock.Anything, testInboxID2, 10, 0).
					Return([]*models.ForwardRule{}, 0, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.ForwardRule{},
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
			core, mockRepo := setupRuleTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.RuleService.ListByInbox(context.Background(), tt.inboxID, tt.limit, tt.offset)
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
