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
	now := time.Now()
	tests := []struct {
		name    string
		message models.Message
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful store",
			message: models.Message{
				InboxID:  1,
				Sender:   "sender@example.com",
				Receiver: "inbox@example.com",
				Subject:  "Test Subject",
				Body:     "Test Body",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateMessage", mock.AnythingOfType("models.Message")).
					Return(models.Message{
						Base: models.Base{
							ID:        1,
							CreatedAt: null.TimeFrom(now),
							UpdatedAt: null.TimeFrom(now),
						},
						InboxID:  1,
						Sender:   "sender@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Test Subject",
						Body:     "Test Body",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			message: models.Message{
				InboxID:  1,
				Sender:   "sender@example.com",
				Receiver: "inbox@example.com",
				Subject:  "Test Subject",
				Body:     "Test Body",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateMessage", mock.AnythingOfType("models.Message")).
					Return(models.Message{}, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			_, err := core.MessageService.Store(tt.message)
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
	now := time.Now()
	tests := []struct {
		name    string
		id      int
		mockFn  func(*mocks.Repository)
		want    models.Message
		wantErr bool
		errType error
	}{
		{
			name: "existing message",
			id:   1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetMessage", 1).Return(models.Message{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  1,
					Sender:   "sender@example.com",
					Receiver: "inbox@example.com",
					Subject:  "Test Subject",
					Body:     "Test Body",
				}, nil)
			},
			want: models.Message{
				Base: models.Base{
					ID:        1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				InboxID:  1,
				Sender:   "sender@example.com",
				Receiver: "inbox@example.com",
				Subject:  "Test Subject",
				Body:     "Test Body",
			},
			wantErr: false,
		},
		{
			name: "non-existent message",
			id:   999,
			mockFn: func(m *mocks.Repository) {
				m.On("GetMessage", 999).Return(models.Message{}, storage.ErrNotFound)
			},
			want:    models.Message{},
			wantErr: true,
			errType: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.MessageService.Get(tt.id)
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
	isRead := true
	tests := []struct {
		name    string
		inboxID int
		limit   int
		offset  int
		isRead  *bool
		mockFn  func(*mocks.Repository)
		want    models.PaginatedResponse
		wantErr bool
	}{
		{
			name:    "successful list with read filter",
			inboxID: 1,
			limit:   10,
			offset:  0,
			isRead:  &isRead,
			mockFn: func(m *mocks.Repository) {
				messages := []models.Message{
					{
						Base:     models.Base{ID: 1},
						InboxID:  1,
						Sender:   "sender1@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 1",
						Body:     "Body 1",
						IsRead:   true,
					},
				}
				m.On("ListMessagesByInboxWithFilter", 1, &isRead, 10, 0).
					Return(messages, 1, nil)
			},
			want: models.PaginatedResponse{
				Data: []models.Message{
					{
						Base:     models.Base{ID: 1},
						InboxID:  1,
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
			inboxID: 1,
			limit:   10,
			offset:  0,
			isRead:  nil,
			mockFn: func(m *mocks.Repository) {
				messages := []models.Message{
					{
						Base:     models.Base{ID: 1},
						InboxID:  1,
						Sender:   "sender1@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 1",
						Body:     "Body 1",
						IsRead:   true,
					},
					{
						Base:     models.Base{ID: 2},
						InboxID:  1,
						Sender:   "sender2@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 2",
						Body:     "Body 2",
						IsRead:   false,
					},
				}
				m.On("ListMessagesByInboxWithFilter", 1, (*bool)(nil), 10, 0).
					Return(messages, 2, nil)
			},
			want: models.PaginatedResponse{
				Data: []models.Message{
					{
						Base:     models.Base{ID: 1},
						InboxID:  1,
						Sender:   "sender1@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 1",
						Body:     "Body 1",
						IsRead:   true,
					},
					{
						Base:     models.Base{ID: 2},
						InboxID:  1,
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
			inboxID: 1,
			limit:   10,
			offset:  0,
			isRead:  nil,
			mockFn: func(m *mocks.Repository) {
				m.On("ListMessagesByInboxWithFilter", 1, (*bool)(nil), 10, 0).
					Return(nil, 0, errors.New("database error"))
			},
			want:    models.PaginatedResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.MessageService.ListByInbox(tt.inboxID, tt.limit, tt.offset, tt.isRead)
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
	tests := []struct {
		name      string
		messageID int
		mockFn    func(*mocks.Repository)
		wantErr   bool
	}{
		{
			name:      "successful mark as read",
			messageID: 1,
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateMessageReadStatus", 1, true).Return(
					models.Message{
						Base:     models.Base{ID: 1},
						InboxID:  1,
						Sender:   "sender1@example.com",
						Receiver: "inbox@example.com",
						Subject:  "Subject 1",
						Body:     "Body 1",
						IsRead:   true,
					}, nil)
			},
			wantErr: false,
		},
		{
			name:      "non-existent message",
			messageID: 999,
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateMessageReadStatus", 999, true).Return(models.Message{}, storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			_, err := core.MessageService.MarkAsRead(tt.messageID)
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
	tests := []struct {
		name      string
		messageID int
		mockFn    func(*mocks.Repository)
		wantErr   bool
	}{
		{
			name:      "successful mark as unread",
			messageID: 1,
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateMessageReadStatus", 1, false).Return(models.Message{
					Base:     models.Base{ID: 1},
					InboxID:  1,
					Sender:   "sender1@example.com",
					Receiver: "inbox@example.com",
					Subject:  "Subject 1",
					Body:     "Body 1",
					IsRead:   false,
				}, nil)
			},
			wantErr: false,
		},
		{
			name:      "non-existent message",
			messageID: 999,
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateMessageReadStatus", 999, false).Return(models.Message{}, storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			_, err := core.MessageService.MarkAsUnread(tt.messageID)
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
	tests := []struct {
		name      string
		messageID int
		mockFn    func(*mocks.Repository)
		wantErr   bool
	}{
		{
			name:      "successful deletion",
			messageID: 1,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteMessage", 1).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "non-existent message",
			messageID: 999,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteMessage", 999).Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupMessageTestCore(t)
			tt.mockFn(mockRepo)

			err := core.MessageService.Delete(tt.messageID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
