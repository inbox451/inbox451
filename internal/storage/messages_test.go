package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"inbox451/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	null "github.com/volatiletech/null/v9"
)

func setupMessageTestDB(t *testing.T) (*repository, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	mock.ExpectPrepare("SELECT (.+) FROM messages WHERE inbox_id")                              // ListMessagesByInbox
	mock.ExpectPrepare("SELECT COUNT(.+) FROM messages WHERE inbox_id")                         // CountMessagesByInbox
	mock.ExpectPrepare("SELECT (.+) FROM messages WHERE id")                                    // GetMessage
	mock.ExpectPrepare("INSERT INTO messages")                                                  // CreateMessage
	mock.ExpectPrepare("UPDATE messages")                                                       // UpdateMessageReadStatus
	mock.ExpectPrepare("DELETE FROM messages")                                                  // DeleteMessage
	mock.ExpectPrepare("SELECT (.+) FROM messages WHERE inbox_id = \\? AND is_read = \\?")      // ListMessagesWithFilter
	mock.ExpectPrepare("SELECT COUNT(.+) FROM messages WHERE inbox_id = \\? AND is_read = \\?") // CountMessagesWithFilter

	listMessages, err := sqlxDB.Preparex("SELECT id, inbox_id, sender, receiver, subject, body, is_read, created_at, updated_at FROM messages WHERE inbox_id = ? LIMIT ? OFFSET ?")
	require.NoError(t, err)

	countMessages, err := sqlxDB.Preparex("SELECT COUNT(*) FROM messages WHERE inbox_id = ?")
	require.NoError(t, err)

	getMessage, err := sqlxDB.Preparex("SELECT id, inbox_id, sender, receiver, subject, body, is_read, created_at, updated_at FROM messages WHERE id = ?")
	require.NoError(t, err)

	createMessage, err := sqlxDB.Preparex("INSERT INTO messages (inbox_id, sender, receiver, subject, body) VALUES (?, ?, ?, ?, ?)")
	require.NoError(t, err)

	updateMessageReadStatus, err := sqlxDB.Preparex("UPDATE messages SET is_read = ? WHERE id = ?")
	require.NoError(t, err)

	deleteMessage, err := sqlxDB.Preparex("DELETE FROM messages WHERE id = ?")
	require.NoError(t, err)

	listMessagesWithFilter, err := sqlxDB.Preparex("SELECT id, inbox_id, sender, receiver, subject, body, is_read, created_at, updated_at FROM messages WHERE inbox_id = ? AND is_read = ? LIMIT ? OFFSET ?")
	require.NoError(t, err)

	countMessagesWithFilter, err := sqlxDB.Preparex("SELECT COUNT(*) FROM messages WHERE inbox_id = ? AND is_read = ?")
	require.NoError(t, err)

	queries := &Queries{
		ListMessagesByInbox:                listMessages,
		CountMessagesByInbox:               countMessages,
		GetMessage:                         getMessage,
		CreateMessage:                      createMessage,
		UpdateMessageReadStatus:            updateMessageReadStatus,
		DeleteMessage:                      deleteMessage,
		ListMessagesByInboxWithReadFilter:  listMessagesWithFilter,
		CountMessagesByInboxWithReadFilter: countMessagesWithFilter,
	}

	repo := &repository{
		db:      sqlxDB,
		queries: queries,
	}

	return repo, mock
}

func TestRepository_CreateMessage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		message *models.Message
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful creation",
			message: &models.Message{
				InboxID:  1,
				Sender:   "sender@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Test Subject",
				Body:     "Test Body",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO messages").
					WithArgs(
						1,
						"sender@example.com",
						"receiver@example.com",
						"Test Subject",
						"Test Body",
					).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
							AddRow(1, now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "database error",
			message: &models.Message{
				InboxID:  1,
				Sender:   "sender@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Test Subject",
				Body:     "Test Body",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO messages").
					WithArgs(
						1,
						"sender@example.com",
						"receiver@example.com",
						"Test Subject",
						"Test Body",
					).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupMessageTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.CreateMessage(context.Background(), tt.message)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, tt.message.ID)
			assert.NotZero(t, tt.message.CreatedAt)
			assert.NotZero(t, tt.message.UpdatedAt)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetMessage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		mockFn  func(sqlmock.Sqlmock)
		want    *models.Message
		wantErr bool
		errType error
	}{
		{
			name: "existing message",
			id:   1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "inbox_id", "sender", "receiver", "subject",
					"body", "is_read", "created_at", "updated_at",
				}).AddRow(
					1, 1, "sender@example.com", "receiver@example.com",
					"Test Subject", "Test Body", false, now, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM messages").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: &models.Message{
				Base: models.Base{
					ID:        1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				InboxID:  1,
				Sender:   "sender@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Test Subject",
				Body:     "Test Body",
				IsRead:   false,
			},
			wantErr: false,
		},
		{
			name: "non-existent message",
			id:   999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM messages").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
			errType: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupMessageTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.GetMessage(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_UpdateMessageReadStatus(t *testing.T) {
	tests := []struct {
		name      string
		messageID int
		isRead    bool
		mockFn    func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "successful update to read",
			messageID: 1,
			isRead:    true,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE messages").
					WithArgs(true, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "successful update to unread",
			messageID: 1,
			isRead:    false,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE messages").
					WithArgs(false, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "non-existent message",
			messageID: 999,
			isRead:    true,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE messages").
					WithArgs(true, 999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupMessageTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.UpdateMessageReadStatus(context.Background(), tt.messageID, tt.isRead)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_DeleteMessage(t *testing.T) {
	tests := []struct {
		name      string
		messageID int
		mockFn    func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "successful deletion",
			messageID: 1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM messages").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "non-existent message",
			messageID: 999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM messages").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupMessageTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.DeleteMessage(context.Background(), tt.messageID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_ListMessagesByInboxWithFilter(t *testing.T) {
	now := time.Now()
	isRead := true

	tests := []struct {
		name    string
		inboxID int
		isRead  *bool
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []*models.Message
		total   int
		wantErr bool
	}{
		{
			name:    "list with read filter",
			inboxID: 1,
			isRead:  &isRead,
			limit:   10,
			offset:  0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1, true).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"id", "inbox_id", "sender", "receiver", "subject",
					"body", "is_read", "created_at", "updated_at",
				}).AddRow(
					1, 1, "sender@example.com", "receiver@example.com",
					"Test Subject", "Test Body", true, now, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM messages").
					WithArgs(1, true, 10, 0).
					WillReturnRows(rows)
			},
			want: []*models.Message{
				{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  1,
					Sender:   "sender@example.com",
					Receiver: "receiver@example.com",
					Subject:  "Test Subject",
					Body:     "Test Body",
					IsRead:   true,
				},
			},
			total:   1,
			wantErr: false,
		},
		{
			name:    "list without filter",
			inboxID: 1,
			isRead:  nil,
			limit:   10,
			offset:  0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"id", "inbox_id", "sender", "receiver", "subject",
					"body", "is_read", "created_at", "updated_at",
				}).
					AddRow(1, 1, "sender1@example.com", "receiver1@example.com",
						"Subject 1", "Body 1", true, now, now).
					AddRow(2, 1, "sender2@example.com", "receiver2@example.com",
						"Subject 2", "Body 2", false, now, now)

				mock.ExpectQuery("SELECT (.+) FROM messages").
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			want: []*models.Message{
				{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  1,
					Sender:   "sender1@example.com",
					Receiver: "receiver1@example.com",
					Subject:  "Subject 1",
					Body:     "Body 1",
					IsRead:   true,
				},
				{
					Base: models.Base{
						ID:        2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  1,
					Sender:   "sender2@example.com",
					Receiver: "receiver2@example.com",
					Subject:  "Subject 2",
					Body:     "Body 2",
					IsRead:   false,
				},
			},
			total:   2,
			wantErr: false,
		},
		{
			name:    "empty result",
			inboxID: 2,
			isRead:  &isRead,
			limit:   10,
			offset:  0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(2, true).
					WillReturnRows(countRows)
			},
			want:    []*models.Message{},
			total:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupMessageTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListMessagesByInboxWithFilter(context.Background(), tt.inboxID, tt.isRead, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.total, total)
			assert.Equal(t, tt.want, got)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_ListMessagesByInbox(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		inboxID int
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []*models.Message
		total   int
		wantErr bool
	}{
		{
			name:    "successful list with messages",
			inboxID: 1,
			limit:   10,
			offset:  0,
			mockFn: func(mock sqlmock.Sqlmock) {
				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1).
					WillReturnRows(countRows)

				// Mock list query
				rows := sqlmock.NewRows([]string{
					"id", "inbox_id", "sender", "receiver", "subject",
					"body", "is_read", "created_at", "updated_at",
				}).
					AddRow(1, 1, "sender1@example.com", "receiver1@example.com",
						"Subject 1", "Body 1", true, now, now).
					AddRow(2, 1, "sender2@example.com", "receiver2@example.com",
						"Subject 2", "Body 2", false, now, now)

				mock.ExpectQuery("SELECT (.+) FROM messages").
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			want: []*models.Message{
				{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  1,
					Sender:   "sender1@example.com",
					Receiver: "receiver1@example.com",
					Subject:  "Subject 1",
					Body:     "Body 1",
					IsRead:   true,
				},
				{
					Base: models.Base{
						ID:        2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  1,
					Sender:   "sender2@example.com",
					Receiver: "receiver2@example.com",
					Subject:  "Subject 2",
					Body:     "Body 2",
					IsRead:   false,
				},
			},
			total:   2,
			wantErr: false,
		},
		{
			name:    "empty inbox",
			inboxID: 2,
			limit:   10,
			offset:  0,
			mockFn: func(mock sqlmock.Sqlmock) {
				// Mock count query returning 0
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(2).
					WillReturnRows(countRows)

				// No need to expect list query when count is 0
			},
			want:    []*models.Message{},
			total:   0,
			wantErr: false,
		},
		{
			name:    "count query error",
			inboxID: 1,
			limit:   10,
			offset:  0,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			want:    nil,
			total:   0,
			wantErr: true,
		},
		{
			name:    "list query error",
			inboxID: 1,
			limit:   10,
			offset:  0,
			mockFn: func(mock sqlmock.Sqlmock) {
				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1).
					WillReturnRows(countRows)

				// Mock list query error
				mock.ExpectQuery("SELECT (.+) FROM messages").
					WithArgs(1, 10, 0).
					WillReturnError(sql.ErrConnDone)
			},
			want:    nil,
			total:   0,
			wantErr: true,
		},
		{
			name:    "with pagination",
			inboxID: 1,
			limit:   2,
			offset:  2,
			mockFn: func(mock sqlmock.Sqlmock) {
				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(4)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1).
					WillReturnRows(countRows)

				// Mock list query with pagination
				rows := sqlmock.NewRows([]string{
					"id", "inbox_id", "sender", "receiver", "subject",
					"body", "is_read", "created_at", "updated_at",
				}).
					AddRow(3, 1, "sender3@example.com", "receiver3@example.com",
						"Subject 3", "Body 3", true, now, now).
					AddRow(4, 1, "sender4@example.com", "receiver4@example.com",
						"Subject 4", "Body 4", false, now, now)

				mock.ExpectQuery("SELECT (.+) FROM messages").
					WithArgs(1, 2, 2).
					WillReturnRows(rows)
			},
			want: []*models.Message{
				{
					Base: models.Base{
						ID:        3,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  1,
					Sender:   "sender3@example.com",
					Receiver: "receiver3@example.com",
					Subject:  "Subject 3",
					Body:     "Body 3",
					IsRead:   true,
				},
				{
					Base: models.Base{
						ID:        4,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  1,
					Sender:   "sender4@example.com",
					Receiver: "receiver4@example.com",
					Subject:  "Subject 4",
					Body:     "Body 4",
					IsRead:   false,
				},
			},
			total:   4,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupMessageTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListMessagesByInbox(context.Background(), tt.inboxID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.total, total)
			assert.Equal(t, tt.want, got)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
