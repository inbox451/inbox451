package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"inbox451/internal/test"

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
	testInboxID1 := test.RandomTestUUID()

	tests := []struct {
		name    string
		rule    *models.ForwardRule
		mockFn  func(sqlmock.Sqlmock)
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
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO forward_rules").
					WithArgs(testInboxID1, "sender@example.com", "receiver@example.com", "Test Subject").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
							AddRow(testInboxID1, now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "database error",
			rule: &models.ForwardRule{
				InboxID:  testInboxID1,
				Sender:   "sender@example.com",
				Receiver: "receiver@example.com",
				Subject:  "Test Subject",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO forward_rules").
					WithArgs(testInboxID1, "sender@example.com", "receiver@example.com", "Test Subject").
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

			err := repo.CreateRule(context.Background(), tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, tt.rule.ID)
			assert.NotZero(t, tt.rule.CreatedAt)
			assert.NotZero(t, tt.rule.UpdatedAt)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetRule(t *testing.T) {
	now := time.Now()
	testInboxID1 := test.RandomTestUUID()
	testRuleID1 := test.RandomTestUUID()
	nonExistingRuleID := test.RandomTestUUID()

	tests := []struct {
		name    string
		id      string
		mockFn  func(sqlmock.Sqlmock)
		want    *models.ForwardRule
		wantErr bool
		errType error
	}{
		{
			name: "existing rule",
			id:   testRuleID1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "inbox_id", "sender", "receiver", "subject", "created_at", "updated_at",
				}).AddRow(testRuleID1, testInboxID1, "sender@example.com", "receiver@example.com", "Test Subject", now, now)

				mock.ExpectQuery("SELECT (.+) FROM forward_rules").
					WithArgs(testRuleID1).
					WillReturnRows(rows)
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
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM forward_rules").
					WithArgs(nonExistingRuleID).
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

			got, err := repo.GetRule(context.Background(), tt.id)
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
	testInboxID1 := test.RandomTestUUID()
	testRuleID1 := test.RandomTestUUID()
	nonExistingRuleID := test.RandomTestUUID()
	tests := []struct {
		name    string
		rule    *models.ForwardRule
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful update",
			rule: &models.ForwardRule{
				Base:     models.Base{ID: testRuleID1},
				InboxID:  testInboxID1,
				Sender:   "updated@example.com",
				Receiver: "newreceiver@example.com",
				Subject:  "Updated Subject",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE forward_rules").
					WithArgs("updated@example.com", "newreceiver@example.com", "Updated Subject", testRuleID1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "non-existent rule",
			rule: &models.ForwardRule{
				Base:     models.Base{ID: nonExistingRuleID},
				InboxID:  testInboxID1,
				Sender:   "updated@example.com",
				Receiver: "newreceiver@example.com",
				Subject:  "Updated Subject",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE forward_rules").
					WithArgs("updated@example.com", "newreceiver@example.com", "Updated Subject", nonExistingRuleID).
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

			err := repo.UpdateRule(context.Background(), tt.rule)
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

func TestRepository_DeleteRule(t *testing.T) {
	testRuleID1 := test.RandomTestUUID()
	nonExistingRuleID := test.RandomTestUUID()
	tests := []struct {
		name    string
		id      string
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   testRuleID1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM forward_rules").
					WithArgs(testRuleID1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "non-existent rule",
			id:   nonExistingRuleID,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM forward_rules").
					WithArgs(nonExistingRuleID).
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

			err := repo.DeleteRule(context.Background(), tt.id)
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
	testInboxID1 := test.RandomTestUUID()
	testRuleID1 := test.RandomTestUUID()
	testRuleID2 := test.RandomTestUUID()

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
					AddRow(testRuleID1, testInboxID1, "sender1@example.com", "receiver1@example.com", "Subject 1", now, now).
					AddRow(testRuleID2, testInboxID1, "sender2@example.com", "receiver2@example.com", "Subject 2", now, now)

				mock.ExpectQuery("SELECT (.+) FROM forward_rules").
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			want: []*models.ForwardRule{
				{
					Base: models.Base{
						ID:        testRuleID1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  testInboxID1,
					Sender:   "sender1@example.com",
					Receiver: "receiver1@example.com",
					Subject:  "Subject 1",
				},
				{
					Base: models.Base{
						ID:        testRuleID2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  testInboxID1,
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

			got, total, err := repo.ListRules(context.Background(), tt.limit, tt.offset)
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
	testInboxID1 := test.RandomTestUUID()
	testInboxID2 := test.RandomTestUUID()
	testRuleID1 := test.RandomTestUUID()
	testRuleID2 := test.RandomTestUUID()

	tests := []struct {
		name    string
		inboxID string
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []*models.ForwardRule
		total   int
		wantErr bool
	}{
		{
			name:    "successful list",
			inboxID: testInboxID1,
			limit:   10,
			offset:  0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(testInboxID1).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"id", "inbox_id", "sender", "receiver", "subject", "created_at", "updated_at",
				}).
					AddRow(testRuleID1, testInboxID1, "sender1@example.com", "receiver1@example.com", "Subject 1", now, now).
					AddRow(testRuleID2, testInboxID1, "sender2@example.com", "receiver2@example.com", "Subject 2", now, now)

				mock.ExpectQuery("SELECT (.+) FROM forward_rules").
					WithArgs(testInboxID1, 10, 0).
					WillReturnRows(rows)
			},
			want: []*models.ForwardRule{
				{
					Base: models.Base{
						ID:        testRuleID1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  testInboxID1,
					Sender:   "sender1@example.com",
					Receiver: "receiver1@example.com",
					Subject:  "Subject 1",
				},
				{
					Base: models.Base{
						ID:        testRuleID2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					InboxID:  testInboxID1,
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
			inboxID: testInboxID2,
			limit:   10,
			offset:  0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(testInboxID2).
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

			got, total, err := repo.ListRulesByInbox(context.Background(), tt.inboxID, tt.limit, tt.offset)
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
