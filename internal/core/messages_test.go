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

func setupMessageTestCore(t *testing.T) (*Core, *mocks.Repository) {
	mockRepo := mocks.NewRepository(t)
	logger := logger.New(io.Discard, logger.DEBUG)

	core := &Core{
		Logger:     logger,
		Repository: mockRepo,
	}
	core.MessageService = NewMessageService(core)

	return core, mockRepo
}

func TestMessageService_Store(t *testing.T) {
	testInboxID := test.RandomTestUUID()
	tests := []struct {
		name    string
		message *models.Message
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful store",
			message: &models.Message{
				InboxID:  testInboxID,
				Sender:   "sender@example.com",
				Receiver: "inbox@example.com",
				Subject:  "Test Subject",
				Body:     "Test Body",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateMessage", mock.Anything, mock.AnythingOfType("*models.Message")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			message: &models.Message{
				InboxID:  testInboxID,
				Sender:   "sender@example.com",
				Receiver: "inbox@example.com",
				Subject:  "Test Subject",
				Body:     "Test Body",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateMessage", mock.Anything, mock.AnythingOfType("*models.Message")).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			err := core.MessageService.Store(context.Background(), tt.message)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMessageService_Get(t *testing.T) {
	testInboxID := test.RandomTestUUID()
	testMessageID := test.RandomTestUUID()
	nonExistingMessageID := test.RandomTestUUID()
	now := time.Now()
	tests := []struct {
		name    string
		id      string
		mockFn  func(*mocks.Repository)
		want    *models.Message
		wantErr bool
		errType error
	}{
		{
			name: "existing message",
			id:   testMessageID,
			mockFn: func(m *mocks.Repository) {
				m.On("GetMessage", mock.Anything, testMessageID).Return(&models.Message{
					Base: models.Base{
						ID:        testMessageID,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  testInboxID,
					Sender:   "sender@example.com",
					Receiver: "inbox@example.com",
					Subject:  "Test Subject",
					Body:     "Test Body",
				}, nil)
			},
			want: &models.Message{
				Base: models.Base{
					ID:        testMessageID,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				InboxID:  testInboxID,
				Sender:   "sender@example.com",
				Receiver: "inbox@example.com",
				Subject:  "Test Subject",
				Body:     "Test Body",
			},
			wantErr: false,
		},
		{
			name: "non-existent message",
			id:   nonExistingMessageID,
			mockFn: func(m *mocks.Repository) {
				m.On("GetMessage", mock.Anything, nonExistingMessageID).Return(nil, storage.ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errType: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.MessageService.Get(context.Background(), tt.id)
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

func TestMessageService_ListByInbox(t *testing.T) {
	testInboxID1 := test.RandomTestUUID()
	testMessageID1 := test.RandomTestUUID()
	testMessageID2 := test.RandomTestUUID()
	isRead := true
	tests := []struct {
		name    string
		inboxID string
		limit   int
		offset  int
		isRead  *bool
		mockFn  func(*mocks.Repository)
		want    *models.PaginatedResponse
		wantErr bool
	}{
		{
			name:    "successful list with read filter",
			inboxID: testInboxID1,
			limit:   10,
			offset:  0,
			isRead:  &isRead,
			mockFn: func(m *mocks.Repository) {
				messages := []*models.Message{
					{
						Base:     models.Base{ID: testMessageID1},
						InboxID:  testInboxID1,
						Sender:   "sender1@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 1",
						Body:     "Body 1",
						IsRead:   isRead,
					},
				}
				m.On("ListMessagesByInboxWithFilter", mock.Anything, testInboxID1, &isRead, 10, 0).
					Return(messages, 1, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.Message{
					{
						Base:     models.Base{ID: testMessageID1},
						InboxID:  testInboxID1,
						Sender:   "sender1@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 1",
						Body:     "Body 1",
						IsRead:   true,
					},
				},
				Pagination: models.Pagination{
					Total:  1,
					Limit:  10,
					Offset: 0,
				},
			},
			wantErr: false,
		},
		{
			name:    "successful list without read filter",
			inboxID: testInboxID1,
			limit:   10,
			offset:  0,
			isRead:  nil,
			mockFn: func(m *mocks.Repository) {
				messages := []*models.Message{
					{
						Base:     models.Base{ID: testMessageID1},
						InboxID:  testInboxID1,
						Sender:   "sender1@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 1",
						Body:     "Body 1",
						IsRead:   true,
					},
					{
						Base:     models.Base{ID: testMessageID2},
						InboxID:  testInboxID1,
						Sender:   "sender2@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 2",
						Body:     "Body 2",
						IsRead:   false,
					},
				}
				m.On("ListMessagesByInboxWithFilter", mock.Anything, testInboxID1, (*bool)(nil), 10, 0).
					Return(messages, 2, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.Message{
					{
						Base:     models.Base{ID: testMessageID1},
						InboxID:  testInboxID1,
						Sender:   "sender1@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 1",
						Body:     "Body 1",
						IsRead:   true,
					},
					{
						Base:     models.Base{ID: testMessageID2},
						InboxID:  testInboxID1,
						Sender:   "sender2@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 2",
						Body:     "Body 2",
						IsRead:   false,
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
			isRead:  nil,
			mockFn: func(m *mocks.Repository) {
				m.On("ListMessagesByInboxWithFilter", mock.Anything, testInboxID1, (*bool)(nil), 10, 0).
					Return([]*models.Message(nil), 0, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.MessageService.ListByInbox(context.Background(), tt.inboxID, tt.limit, tt.offset, tt.isRead)
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

func TestMessageService_MarkAsRead(t *testing.T) {
	testMessageID1 := test.RandomTestUUID()
	nonExistingMessageID := test.RandomTestUUID()
	tests := []struct {
		name      string
		messageID string
		mockFn    func(*mocks.Repository)
		wantErr   bool
	}{
		{
			name:      "successful mark as read",
			messageID: testMessageID1,
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateMessageReadStatus", mock.Anything, testMessageID1, true).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "non-existent message",
			messageID: nonExistingMessageID,
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateMessageReadStatus", mock.Anything, nonExistingMessageID, true).Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			err := core.MessageService.MarkAsRead(context.Background(), tt.messageID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMessageService_MarkAsUnread(t *testing.T) {
	testMessageID1 := test.RandomTestUUID()
	nonExistingMessageID := test.RandomTestUUID()
	tests := []struct {
		name      string
		messageID string
		mockFn    func(*mocks.Repository)
		wantErr   bool
	}{
		{
			name:      "successful mark as unread",
			messageID: testMessageID1,
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateMessageReadStatus", mock.Anything, testMessageID1, false).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "non-existent message",
			messageID: nonExistingMessageID,
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateMessageReadStatus", mock.Anything, nonExistingMessageID, false).Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			err := core.MessageService.MarkAsUnread(context.Background(), tt.messageID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMessageService_Delete(t *testing.T) {
	testMessageID1 := test.RandomTestUUID()
	nonExistingMessageID := test.RandomTestUUID()
	tests := []struct {
		name      string
		messageID string
		mockFn    func(*mocks.Repository)
		wantErr   bool
	}{
		{
			name:      "successful deletion",
			messageID: testMessageID1,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteMessage", mock.Anything, testMessageID1).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "non-existent message",
			messageID: nonExistingMessageID,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteMessage", mock.Anything, nonExistingMessageID).Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			err := core.MessageService.Delete(context.Background(), tt.messageID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
