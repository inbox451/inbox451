package storage

import (
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

func setupRuleTestDB(t *testing.T) (*repository, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	mock.ExpectPrepare("SELECT (.+) FROM forward_rules WHERE inbox_id")      // ListRulesByInbox
	mock.ExpectPrepare("SELECT COUNT(.+) FROM forward_rules WHERE inbox_id") // CountRulesByInbox
	mock.ExpectPrepare("SELECT (.+) FROM forward_rules ORDER BY")            // ListRules
	mock.ExpectPrepare("SELECT COUNT(.+) FROM forward_rules$")               // CountRules
	mock.ExpectPrepare("SELECT (.+) FROM forward_rules WHERE id")            // GetRule
	mock.ExpectPrepare("INSERT INTO forward_rules")                          // CreateRule
	mock.ExpectPrepare("UPDATE forward_rules")                               // UpdateRule
	mock.ExpectPrepare("DELETE FROM forward_rules")                          // DeleteRule

	listRulesByInbox, err := sqlxDB.Preparex("SELECT id, inbox_id, sender, receiver, subject, created_at, updated_at FROM forward_rules WHERE inbox_id = ? LIMIT ? OFFSET ?")
	require.NoError(t, err)

	countRulesByInbox, err := sqlxDB.Preparex("SELECT COUNT(*) FROM forward_rules WHERE inbox_id = ?")
	require.NoError(t, err)

	listRules, err := sqlxDB.Preparex("SELECT id, inbox_id, sender, receiver, subject, created_at, updated_at FROM forward_rules ORDER BY id LIMIT ? OFFSET ?")
	require.NoError(t, err)

	countRules, err := sqlxDB.Preparex("SELECT COUNT(*) FROM forward_rules")
	require.NoError(t, err)

	getRule, err := sqlxDB.Preparex("SELECT id, inbox_id, sender, receiver, subject, created_at, updated_at FROM forward_rules WHERE id = ?")
	require.NoError(t, err)

	createRule, err := sqlxDB.Preparex("INSERT INTO forward_rules (inbox_id, sender, receiver, subject) VALUES (?, ?, ?, ?)")
	require.NoError(t, err)

	updateRule, err := sqlxDB.Preparex("UPDATE forward_rules SET sender = ?, receiver = ?, subject = ? WHERE id = ?")
	require.NoError(t, err)

	deleteRule, err := sqlxDB.Preparex("DELETE FROM forward_rules WHERE id = ?")
	require.NoError(t, err)

	// Initialize queries struct in the same order
	queries := &Queries{
		ListRulesByInbox:  listRulesByInbox,
		CountRulesByInbox: countRulesByInbox,
		ListRules:         listRules,
		CountRules:        countRules,
		GetRule:           getRule,
		CreateRule:        createRule,
		UpdateRule:        updateRule,
		DeleteRule:        deleteRule,
	}

	repo := &repository{
		db:      sqlxDB,
		queries: queries,
	}

	return repo, mock
}

func TestRepository_CreateRule(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		rule    models.ForwardRule
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful creation",
			rule: models.ForwardRule{
				InboxID:  1,
				Sender:   "sender@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Test Subject",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO forward_rules").
					WithArgs(1, "sender@example.com", "receiver@example.com", "Test Subject").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
							AddRow(1, now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "database error",
			rule: models.ForwardRule{
				InboxID:  1,
				Sender:   "sender@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Test Subject",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO forward_rules").
					WithArgs(1, "sender@example.com", "receiver@example.com", "Test Subject").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupRuleTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.CreateRule(tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, got.ID)
			assert.NotZero(t, got.CreatedAt)
			assert.NotZero(t, got.UpdatedAt)
			assert.Equal(t, tt.rule.Subject, got.Subject)
			assert.Equal(t, tt.rule.Sender, got.Sender)
			assert.Equal(t, tt.rule.Receiver, got.Receiver)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetRule(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		mockFn  func(sqlmock.Sqlmock)
		want    *models.ForwardRule
		wantErr bool
		errType error
	}{
		{
			name: "existing rule",
			id:   1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "inbox_id", "sender", "receiver", "subject", "created_at", "updated_at",
				}).AddRow(1, 1, "sender@example.com", "receiver@example.com", "Test Subject", now, now)

				mock.ExpectQuery("SELECT (.+) FROM forward_rules").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: &models.ForwardRule{
				Base: models.Base{
					ID:        1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				InboxID:  1,
				Sender:   "sender@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Test Subject",
			},
			wantErr: false,
		},
		{
			name: "non-existent rule",
			id:   999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM forward_rules").
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
			repo, mock := setupRuleTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.GetRule(tt.id)
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

func TestRepository_UpdateRule(t *testing.T) {
	tests := []struct {
		name    string
		rule    models.ForwardRule
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful update",
			rule: models.ForwardRule{
				Base:     models.Base{ID: 1},
				InboxID:  1,
				Sender:   "updated@example.com",
				Receiver: "newreceiver@example.com",
				Subject:  "Updated Subject",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE forward_rules").
					WithArgs("updated@example.com", "newreceiver@example.com", "Updated Subject", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "non-existent rule",
			rule: models.ForwardRule{
				Base:     models.Base{ID: 999},
				InboxID:  1,
				Sender:   "updated@example.com",
				Receiver: "newreceiver@example.com",
				Subject:  "Updated Subject",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE forward_rules").
					WithArgs("updated@example.com", "newreceiver@example.com", "Updated Subject", 999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupRuleTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.UpdateRule(tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.rule.Sender, got.Sender)
			assert.Equal(t, tt.rule.Receiver, got.Receiver)
			assert.Equal(t, tt.rule.Subject, got.Subject)

			assert.NoError(t, err)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_DeleteRule(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM forward_rules").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "non-existent rule",
			id:   999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM forward_rules").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupRuleTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.DeleteRule(tt.id)
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

func TestRepository_ListRules(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []*models.ForwardRule
		total   int
		wantErr bool
	}{
		{
			name:   "successful list",
			limit:  10,
			offset: 0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"id", "inbox_id", "sender", "receiver", "subject", "created_at", "updated_at",
				}).
					AddRow(1, 1, "sender1@example.com", "receiver1@example.com", "Subject 1", now, now).
					AddRow(2, 1, "sender2@example.com", "receiver2@example.com", "Subject 2", now, now)

				mock.ExpectQuery("SELECT (.+) FROM forward_rules").
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			want: []*models.ForwardRule{
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
				},
			},
			total:   2,
			wantErr: false,
		},
		{
			name:   "empty list",
			limit:  10,
			offset: 0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)
			},
			want:    []*models.ForwardRule{},
			total:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupRuleTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListRules(tt.limit, tt.offset)
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

func TestRepository_ListRulesByInbox(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		inboxID int
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []*models.ForwardRule
		total   int
		wantErr bool
	}{
		{
			name:    "successful list",
			inboxID: 1,
			limit:   10,
			offset:  0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"id", "inbox_id", "sender", "receiver", "subject", "created_at", "updated_at",
				}).
					AddRow(1, 1, "sender1@example.com", "receiver1@example.com", "Subject 1", now, now).
					AddRow(2, 1, "sender2@example.com", "receiver2@example.com", "Subject 2", now, now)

				mock.ExpectQuery("SELECT (.+) FROM forward_rules").
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			want: []*models.ForwardRule{
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
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(2).
					WillReturnRows(countRows)
			},
			want:    []*models.ForwardRule{},
			total:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupRuleTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListRulesByInbox(tt.inboxID, tt.limit, tt.offset)
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
