package core

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"inbox451/internal/config"
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
		Config:     &config.Config{},
		Logger:     logger,
		Repository: mockRepo,
	}
	core.InboxService = NewInboxService(core)

	return core, mockRepo
}

func setupInboxTestCoreWithEmailDomain(t *testing.T, emailDomain string) (*Core, *mocks.Repository) {
	core, mockRepo := setupInboxTestCore(t)
	core.Config.Server.EmailDomain = emailDomain
	return core, mockRepo
}

func TestInboxService_Create(t *testing.T) {
	tests := []struct {
		name    string
		inbox   *models.Inbox
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful creation",
			inbox: &models.Inbox{
				ProjectID: 1,
				Email:     "test@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateInbox", mock.Anything, mock.AnythingOfType("*models.Inbox")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			inbox: &models.Inbox{
				ProjectID: 1,
				Email:     "test@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateInbox", mock.Anything, mock.AnythingOfType("*models.Inbox")).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			err := core.InboxService.Create(context.Background(), tt.inbox)
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
		want    *models.Inbox
		wantErr bool
		errType error
	}{
		{
			name: "existing inbox",
			id:   1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetInbox", mock.Anything, 1).Return(&models.Inbox{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					ProjectID: 1,
					Email:     "test@example.com",
				}, nil)
			},
			want: &models.Inbox{
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
				m.On("GetInbox", mock.Anything, 999).Return(nil, storage.ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errType: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.InboxService.Get(context.Background(), tt.id)
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
		inbox   *models.Inbox
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful update",
			inbox: &models.Inbox{
				Base:      models.Base{ID: 1},
				ProjectID: 1,
				Email:     "updated@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateInbox", mock.Anything, mock.AnythingOfType("*models.Inbox")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "update non-existent inbox",
			inbox: &models.Inbox{
				Base:      models.Base{ID: 999},
				ProjectID: 1,
				Email:     "updated@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateInbox", mock.Anything, mock.AnythingOfType("*models.Inbox")).
					Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			err := core.InboxService.Update(context.Background(), tt.inbox)
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
				m.On("DeleteInbox", mock.Anything, 1).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "delete non-existent inbox",
			id:   999,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteInbox", mock.Anything, 999).Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			err := core.InboxService.Delete(context.Background(), tt.id)
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
		want      *models.PaginatedResponse
		wantErr   bool
	}{
		{
			name:      "successful list",
			projectID: 1,
			limit:     10,
			offset:    0,
			mockFn: func(m *mocks.Repository) {
				inboxes := []*models.Inbox{
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
				m.On("ListInboxesByProject", mock.Anything, 1, 10, 0).Return(inboxes, 2, nil)
			},
			want: &models.PaginatedResponse{
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
				m.On("ListInboxesByProject", mock.Anything, 1, 10, 0).
					Return([]*models.Inbox(nil), 0, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:      "empty project",
			projectID: 2,
			limit:     10,
			offset:    0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListInboxesByProject", mock.Anything, 2, 10, 0).
					Return([]*models.Inbox{}, 0, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.Inbox{},
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

			got, err := core.InboxService.ListByProject(context.Background(), tt.projectID, tt.limit, tt.offset)
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

func TestInboxService_Create_WithEmailDomain(t *testing.T) {
	tests := []struct {
		name        string
		emailDomain string
		inbox       *models.Inbox
		mockFn      func(*mocks.Repository)
		wantEmail   string
		wantErr     bool
		errMessage  string
	}{
		{
			name:        "auto-append domain to local part",
			emailDomain: "example.com",
			inbox: &models.Inbox{
				ProjectID: 1,
				Email:     "testinbox",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateInbox", mock.Anything, mock.MatchedBy(func(inbox *models.Inbox) bool {
					return inbox.Email == "testinbox@example.com"
				})).Return(nil)
			},
			wantEmail: "testinbox@example.com",
			wantErr:   false,
		},
		{
			name:        "accept full email with correct domain",
			emailDomain: "example.com",
			inbox: &models.Inbox{
				ProjectID: 1,
				Email:     "test@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateInbox", mock.Anything, mock.MatchedBy(func(inbox *models.Inbox) bool {
					return inbox.Email == "test@example.com"
				})).Return(nil)
			},
			wantEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name:        "reject email with incorrect domain",
			emailDomain: "example.com",
			inbox: &models.Inbox{
				ProjectID: 1,
				Email:     "test@otherdomain.com",
			},
			mockFn:     func(m *mocks.Repository) {},
			wantErr:    true,
			errMessage: "Inbox email must end with @example.com",
		},
		{
			name:        "no domain configured - accept any email",
			emailDomain: "",
			inbox: &models.Inbox{
				ProjectID: 1,
				Email:     "test@anydomain.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateInbox", mock.Anything, mock.MatchedBy(func(inbox *models.Inbox) bool {
					return inbox.Email == "test@anydomain.com"
				})).Return(nil)
			},
			wantEmail: "test@anydomain.com",
			wantErr:   false,
		},
		{
			name:        "auto-append domain with special characters",
			emailDomain: "sub.example.com",
			inbox: &models.Inbox{
				ProjectID: 1,
				Email:     "user+tag",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateInbox", mock.Anything, mock.MatchedBy(func(inbox *models.Inbox) bool {
					return inbox.Email == "user+tag@sub.example.com"
				})).Return(nil)
			},
			wantEmail: "user+tag@sub.example.com",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCoreWithEmailDomain(t, tt.emailDomain)
			tt.mockFn(mockRepo)

			err := core.InboxService.Create(context.Background(), tt.inbox)
			if tt.wantErr {
				assert.Error(t, err)
				apiErr, ok := err.(*APIError)
				assert.True(t, ok, "error should be APIError")
				assert.Equal(t, http.StatusBadRequest, apiErr.Code)
				assert.Equal(t, tt.errMessage, apiErr.Message)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantEmail, tt.inbox.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestInboxService_Update_WithEmailDomain(t *testing.T) {
	tests := []struct {
		name        string
		emailDomain string
		inbox       *models.Inbox
		mockFn      func(*mocks.Repository)
		wantEmail   string
		wantErr     bool
		errMessage  string
	}{
		{
			name:        "auto-append domain to local part on update",
			emailDomain: "example.com",
			inbox: &models.Inbox{
				Base:      models.Base{ID: 1},
				ProjectID: 1,
				Email:     "updatedinbox",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateInbox", mock.Anything, mock.MatchedBy(func(inbox *models.Inbox) bool {
					return inbox.Email == "updatedinbox@example.com"
				})).Return(nil)
			},
			wantEmail: "updatedinbox@example.com",
			wantErr:   false,
		},
		{
			name:        "accept updated email with correct domain",
			emailDomain: "example.com",
			inbox: &models.Inbox{
				Base:      models.Base{ID: 1},
				ProjectID: 1,
				Email:     "updated@example.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateInbox", mock.Anything, mock.MatchedBy(func(inbox *models.Inbox) bool {
					return inbox.Email == "updated@example.com"
				})).Return(nil)
			},
			wantEmail: "updated@example.com",
			wantErr:   false,
		},
		{
			name:        "reject updated email with incorrect domain",
			emailDomain: "example.com",
			inbox: &models.Inbox{
				Base:      models.Base{ID: 1},
				ProjectID: 1,
				Email:     "updated@wrongdomain.com",
			},
			mockFn:     func(m *mocks.Repository) {},
			wantErr:    true,
			errMessage: "Inbox email must end with @example.com",
		},
		{
			name:        "no domain configured - accept any updated email",
			emailDomain: "",
			inbox: &models.Inbox{
				Base:      models.Base{ID: 1},
				ProjectID: 1,
				Email:     "updated@anydomain.com",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateInbox", mock.Anything, mock.MatchedBy(func(inbox *models.Inbox) bool {
					return inbox.Email == "updated@anydomain.com"
				})).Return(nil)
			},
			wantEmail: "updated@anydomain.com",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCoreWithEmailDomain(t, tt.emailDomain)
			tt.mockFn(mockRepo)

			err := core.InboxService.Update(context.Background(), tt.inbox)
			if tt.wantErr {
				assert.Error(t, err)
				apiErr, ok := err.(*APIError)
				assert.True(t, ok, "error should be APIError")
				assert.Equal(t, http.StatusBadRequest, apiErr.Code)
				assert.Equal(t, tt.errMessage, apiErr.Message)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantEmail, tt.inbox.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
