package core

import (
	"context"
	"errors"
	"inbox451/internal/test"
	"io"
	"net/http"
	"testing"
	"time"

	"inbox451/internal/test"

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
				ProjectID: test.StaticTestUUID(),
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
				ProjectID: test.StaticTestUUID(),
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
	testProjectID := test.RandomTestUUID()
	testInboxID := test.RandomTestUUID()
	nonExistingInboxID := test.RandomTestUUID()
	now := time.Now()
	tests := []struct {
		name    string
		id      string
		mockFn  func(*mocks.Repository)
		want    *models.Inbox
		wantErr bool
		errType error
	}{
		{
			name: "existing inbox",
			id:   testInboxID,
			mockFn: func(m *mocks.Repository) {
				m.On("GetInbox", mock.Anything, testInboxID).Return(&models.Inbox{
					Base: models.Base{
						ID:        testInboxID,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					ProjectID: testProjectID,
					Email:     "test@example.com",
				}, nil)
			},
			want: &models.Inbox{
				Base: models.Base{
					ID:        testInboxID,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				ProjectID: testProjectID,
				Email:     "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "non-existent inbox",
			id:   nonExistingInboxID,
			mockFn: func(m *mocks.Repository) {
				m.On("GetInbox", mock.Anything, nonExistingInboxID).Return(nil, storage.ErrNotFound)
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
	testProjectID := test.RandomTestUUID()
	testInboxID := test.RandomTestUUID()
	nonExistingInboxID := test.RandomTestUUID()
	tests := []struct {
		name    string
		inbox   *models.Inbox
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful update",
			inbox: &models.Inbox{
				Base:      models.Base{ID: testInboxID},
				ProjectID: testProjectID,
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
				Base:      models.Base{ID: nonExistingInboxID},
				ProjectID: testProjectID,
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
	testInboxID := test.RandomTestUUID()
	nonExistingInboxID := test.RandomTestUUID()
	tests := []struct {
		name    string
		id      string
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   testInboxID,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteInbox", mock.Anything, testInboxID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "delete non-existent inbox",
			id:   nonExistingInboxID,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteInbox", mock.Anything, nonExistingInboxID).Return(storage.ErrNotFound)
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
	testProjectID1 := test.RandomTestUUID()
	testInboxID1 := test.RandomTestUUID()
	testInboxID2 := test.RandomTestUUID()
	emptyProjectID := test.RandomTestUUID()
	tests := []struct {
		name      string
		projectID string
		limit     int
		offset    int
		mockFn    func(*mocks.Repository)
		want      *models.PaginatedResponse
		wantErr   bool
	}{
		{
			name:      "successful list",
			projectID: testProjectID1,
			limit:     10,
			offset:    0,
			mockFn: func(m *mocks.Repository) {
				inboxes := []*models.Inbox{
					{
						Base:      models.Base{ID: testInboxID1},
						ProjectID: testProjectID1,
						Email:     "inbox1@example.com",
					},
					{
						Base:      models.Base{ID: testInboxID2},
						ProjectID: testProjectID1,
						Email:     "inbox2@example.com",
					},
				}
				m.On("ListInboxesByProject", mock.Anything, testProjectID1, 10, 0).Return(inboxes, 2, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.Inbox{
					{
						Base:      models.Base{ID: testInboxID1},
						ProjectID: testProjectID1,
						Email:     "inbox1@example.com",
					},
					{
						Base:      models.Base{ID: testInboxID2},
						ProjectID: testProjectID1,
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
			projectID: testProjectID1,
			limit:     10,
			offset:    0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListInboxesByProject", mock.Anything, testProjectID1, 10, 0).
					Return([]*models.Inbox(nil), 0, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:      "empty project",
			projectID: emptyProjectID,
			limit:     10,
			offset:    0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListInboxesByProject", mock.Anything, emptyProjectID, 10, 0).
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
	testProjectID := test.RandomTestUUID()
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
				ProjectID: testProjectID,
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
				ProjectID: testProjectID,
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
				ProjectID: testProjectID,
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
				ProjectID: testProjectID,
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
				ProjectID: testProjectID,
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
	testProjectID := test.RandomTestUUID()
	testInboxID1 := test.RandomTestUUID()
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
				Base:      models.Base{ID: testInboxID1},
				ProjectID: testProjectID,
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
				Base:      models.Base{ID: testInboxID1},
				ProjectID: testProjectID,
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
				Base:      models.Base{ID: testInboxID1},
				ProjectID: testProjectID,
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
				Base:      models.Base{ID: testInboxID1},
				ProjectID: testProjectID,
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

func TestInboxService_ListByUser(t *testing.T) {
	now := time.Now()
	testUserID1 := test.RandomTestUUID()
	testUserID2 := test.RandomTestUUID()
	testInboxID1 := test.RandomTestUUID()
	testInboxID2 := test.RandomTestUUID()
	testProjectID1 := test.RandomTestUUID()
	testProjectID2 := test.RandomTestUUID()
	tests := []struct {
		name    string
		userID  string
		mockFn  func(*mocks.Repository)
		want    []*models.Inbox
		wantErr bool
	}{
		{
			name:   "successful list with multiple inboxes",
			userID: testUserID1,
			mockFn: func(m *mocks.Repository) {
				inboxes := []*models.Inbox{
					{
						Base: models.Base{
							ID:        testInboxID1,
							CreatedAt: null.TimeFrom(now),
							UpdatedAt: null.TimeFrom(now),
						},
						ProjectID: testProjectID1,
						Email:     "inbox1@example.com",
					},
					{
						Base: models.Base{
							ID:        testInboxID2,
							CreatedAt: null.TimeFrom(now),
							UpdatedAt: null.TimeFrom(now),
						},
						ProjectID: testProjectID2,
						Email:     "inbox2@example.com",
					},
				}
				m.On("ListInboxesByUser", mock.Anything, testUserID1).Return(inboxes, nil)
			},
			want: []*models.Inbox{
				{
					Base: models.Base{
						ID:        testInboxID1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					ProjectID: testProjectID1,
					Email:     "inbox1@example.com",
				},
				{
					Base: models.Base{
						ID:        testInboxID2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					ProjectID: testProjectID2,
					Email:     "inbox2@example.com",
				},
			},
			wantErr: false,
		},
		{
			name:   "successful list with no inboxes",
			userID: testUserID2,
			mockFn: func(m *mocks.Repository) {
				m.On("ListInboxesByUser", mock.Anything, testUserID2).Return([]*models.Inbox{}, nil)
			},
			want:    []*models.Inbox{},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("ListInboxesByUser", mock.Anything, testUserID1).Return([]*models.Inbox(nil), errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.InboxService.ListByUser(context.Background(), tt.userID)
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

func TestInboxService_GetByEmailAndUser(t *testing.T) {
	now := time.Now()
	testUserID1 := test.RandomTestUUID()
	testInboxID1 := test.RandomTestUUID()
	testProjectID1 := test.RandomTestUUID()
	tests := []struct {
		name    string
		email   string
		userID  string
		mockFn  func(*mocks.Repository)
		want    *models.Inbox
		wantErr bool
		errType error
	}{
		{
			name:   "existing inbox",
			email:  "inbox@example.com",
			userID: testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetInboxByEmailAndUser", mock.Anything, "inbox@example.com", testUserID1).Return(&models.Inbox{
					Base: models.Base{
						ID:        testInboxID1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					ProjectID: testProjectID1,
					Email:     "inbox@example.com",
				}, nil)
			},
			want: &models.Inbox{
				Base: models.Base{
					ID:        testInboxID1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				ProjectID: testProjectID1,
				Email:     "inbox@example.com",
			},
			wantErr: false,
		},
		{
			name:   "non-existent inbox",
			email:  "nonexistent@example.com",
			userID: testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetInboxByEmailAndUser", mock.Anything, "nonexistent@example.com", testUserID1).Return(nil, storage.ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errType: storage.ErrNotFound,
		},
		{
			name:   "repository error",
			email:  "inbox@example.com",
			userID: testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetInboxByEmailAndUser", mock.Anything, "inbox@example.com", testUserID1).Return(nil, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupInboxTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.InboxService.GetByEmailAndUser(context.Background(), tt.email, tt.userID)
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
