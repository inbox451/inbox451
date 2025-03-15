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

func setupInboxTestCore(t *testing.T) (*Core, *mocks.Repository) {
	mockRepo := mocks.NewRepository(t)
	logger := logger.New(io.Discard, logger.DEBUG)

	core := &Core{
		Logger:     logger,
		Repository: mockRepo,
	}
	core.InboxService = NewInboxService(core)

	return core, mockRepo
}

func TestInboxService_Create(t *testing.T) {
	tests := []struct {
		name    string
		inbox   models.Inbox
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful creation",
			inbox: models.Inbox{
				ProjectID: 1,
				Email:     "test@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateInbox", mock.AnythingOfType("models.Inbox")).
					Return(models.Inbox{
						ProjectID: 1,
						Email:     "test@example.com",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			inbox: models.Inbox{
				ProjectID: 1,
				Email:     "test@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateInbox", mock.AnythingOfType("models.Inbox")).
					Return(models.Inbox{}, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			_, err := core.InboxService.Create(tt.inbox)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInboxService_Get(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		id      int
		mockFn  func(*mocks.Repository)
		want    models.Inbox
		wantErr bool
		errType error
	}{
		{
			name: "existing inbox",
			id:   1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetInbox", 1).Return(models.Inbox{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					ProjectID: 1,
					Email:     "test@example.com",
				}, nil)
			},
			want: models.Inbox{
				Base: models.Base{
					ID:        1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				ProjectID: 1,
				Email:     "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "non-existent inbox",
			id:   999,
			mockFn: func(m *mocks.Repository) {
				m.On("GetInbox", 999).Return(models.Inbox{}, storage.ErrNotFound)
			},
			want:    models.Inbox{},
			wantErr: true,
			errType: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.InboxService.Get(tt.id)
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

func TestInboxService_Update(t *testing.T) {
	tests := []struct {
		name    string
		inbox   models.Inbox
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful update",
			inbox: models.Inbox{
				Base:      models.Base{ID: 1},
				ProjectID: 1,
				Email:     "updated@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateInbox", mock.AnythingOfType("models.Inbox")).
					Return(models.Inbox{
						Base:      models.Base{ID: 1},
						ProjectID: 1,
						Email:     "updated@example.com",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "update non-existent inbox",
			inbox: models.Inbox{
				Base:      models.Base{ID: 999},
				ProjectID: 1,
				Email:     "updated@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateInbox", mock.AnythingOfType("models.Inbox")).
					Return(models.Inbox{}, storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			_, err := core.InboxService.Update(tt.inbox)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInboxService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   1,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteInbox", 1).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "delete non-existent inbox",
			id:   999,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteInbox", 999).Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			err := core.InboxService.Delete(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInboxService_ListByProject(t *testing.T) {
	tests := []struct {
		name      string
		projectID int
		limit     int
		offset    int
		mockFn    func(*mocks.Repository)
		want      models.PaginatedResponse
		wantErr   bool
	}{
		{
			name:      "successful list",
			projectID: 1,
			limit:     10,
			offset:    0,
			mockFn: func(m *mocks.Repository) {
				inboxes := []models.Inbox{
					{
						Base:      models.Base{ID: 1},
						ProjectID: 1,
						Email:     "inbox1@example.com",
					},
					{
						Base:      models.Base{ID: 2},
						ProjectID: 1,
						Email:     "inbox2@example.com",
					},
				}
				m.On("ListInboxesByProject", 1, 10, 0).Return(inboxes, 2, nil)
			},
			want: models.PaginatedResponse{
				Data: []*models.Inbox{
					{
						Base:      models.Base{ID: 1},
						ProjectID: 1,
						Email:     "inbox1@example.com",
					},
					{
						Base:      models.Base{ID: 2},
						ProjectID: 1,
						Email:     "inbox2@example.com",
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
			name:      "repository error",
			projectID: 1,
			limit:     10,
			offset:    0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListInboxesByProject", 1, 10, 0).
					Return([]*models.Inbox(nil), 0, errors.New("database error"))
			},
			want:    models.PaginatedResponse{},
			wantErr: true,
		},
		{
			name:      "empty project",
			projectID: 2,
			limit:     10,
			offset:    0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListInboxesByProject", 2, 10, 0).
					Return([]*models.Inbox{}, 0, nil)
			},
			want: models.PaginatedResponse{
				Data: []models.Inbox{},
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
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.InboxService.ListByProject(tt.projectID, tt.limit, tt.offset)
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
