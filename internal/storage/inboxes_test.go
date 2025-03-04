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

func setupInboxTestDB(t *testing.T) (*repository, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	mock.ExpectPrepare("SELECT (.+) FROM inboxes")             // ListInboxes
	mock.ExpectPrepare("SELECT COUNT(.+) FROM inboxes")        // CountInboxes
	mock.ExpectPrepare("SELECT (.+) FROM inboxes WHERE id")    // GetInbox
	mock.ExpectPrepare("INSERT INTO inboxes")                  // CreateInbox
	mock.ExpectPrepare("UPDATE inboxes")                       // UpdateInbox
	mock.ExpectPrepare("DELETE FROM inboxes")                  // DeleteInbox
	mock.ExpectPrepare("SELECT (.+) FROM inboxes WHERE email") // GetInboxByEmail

	listInboxes, err := sqlxDB.Preparex("SELECT id, project_id, email, created_at, updated_at FROM inboxes WHERE project_id = ? LIMIT ? OFFSET ?")
	require.NoError(t, err)

	countInboxes, err := sqlxDB.Preparex("SELECT COUNT(*) FROM inboxes WHERE project_id = ?")
	require.NoError(t, err)

	getInbox, err := sqlxDB.Preparex("SELECT id, project_id, email, created_at, updated_at FROM inboxes WHERE id = ?")
	require.NoError(t, err)

	createInbox, err := sqlxDB.Preparex("INSERT INTO inboxes (project_id, email) VALUES (?, ?)")
	require.NoError(t, err)

	updateInbox, err := sqlxDB.Preparex("UPDATE inboxes SET email = ? WHERE id = ?")
	require.NoError(t, err)

	deleteInbox, err := sqlxDB.Preparex("DELETE FROM inboxes WHERE id = ?")
	require.NoError(t, err)

	getInboxByEmail, err := sqlxDB.Preparex("SELECT id, project_id, email, created_at, updated_at FROM inboxes WHERE email = ?")
	require.NoError(t, err)

	queries := &Queries{
		ListInboxesByProject:  listInboxes,
		CountInboxesByProject: countInboxes,
		GetInbox:              getInbox,
		CreateInbox:           createInbox,
		UpdateInbox:           updateInbox,
		DeleteInbox:           deleteInbox,
		GetInboxByEmail:       getInboxByEmail,
	}

	repo := &repository{
		db:      sqlxDB,
		queries: queries,
	}

	return repo, mock
}

func TestRepository_CreateInbox(t *testing.T) {
	tests := []struct {
		name    string
		inbox   models.Inbox
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful creation",
			inbox: models.Inbox{
				ProjectID: 1,
				Email:     "test@example.com",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO inboxes").
					WithArgs(1, "test@example.com").
					WillReturnRows(
						sqlmock.NewRows([]string{"id"}).AddRow(1),
					)
				// after insert, we expect a select to get the created inbox
				now := time.Now()
				mock.ExpectQuery("SELECT (.+) FROM inboxes WHERE id").
					WithArgs(1).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "project_id", "email", "created_at", "updated_at"}).
							AddRow(1, 1, "test@example.com", now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			inbox: models.Inbox{
				ProjectID: 1,
				Email:     "existing@example.com",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO inboxes").
					WithArgs(1, "existing@example.com").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupInboxTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			inbox, err := repo.CreateInbox(tt.inbox)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, inbox.ID)
			assert.NotZero(t, inbox.CreatedAt)
			assert.NotZero(t, inbox.UpdatedAt)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetInbox(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		mockFn  func(sqlmock.Sqlmock)
		want    models.Inbox
		wantErr bool
		errType error
	}{
		{
			name: "existing inbox",
			id:   1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "project_id", "email", "created_at", "updated_at",
				}).AddRow(1, 1, "test@example.com", now, now)

				mock.ExpectQuery("SELECT (.+) FROM inboxes").
					WithArgs(1).
					WillReturnRows(rows)
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
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM inboxes").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			want:    models.Inbox{},
			wantErr: true,
			errType: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupInboxTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.GetInbox(tt.id)
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

func TestRepository_GetInboxByEmail(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		email   string
		mockFn  func(sqlmock.Sqlmock)
		want    models.Inbox
		wantErr bool
	}{
		{
			name:  "existing inbox",
			email: "test@example.com",
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "project_id", "email", "created_at", "updated_at",
				}).AddRow(1, 1, "test@example.com", now, now)

				mock.ExpectQuery("SELECT (.+) FROM inboxes").
					WithArgs("test@example.com").
					WillReturnRows(rows)
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
			name:  "non-existent inbox",
			email: "nonexistent@example.com",
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM inboxes").
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			want:    models.Inbox{},
			wantErr: true, // Note: This is false because GetInboxByEmail returns nil, nil for not found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupInboxTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.GetInboxByEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_UpdateInbox(t *testing.T) {
	tests := []struct {
		name    string
		inbox   models.Inbox
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful update",
			inbox: models.Inbox{
				Base:      models.Base{ID: 1},
				ProjectID: 1,
				Email:     "updated@example.com",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE inboxes").
					WithArgs("updated@example.com", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// after update, we expect a select to get the updated inbox
				now := time.Now()
				mock.ExpectQuery("SELECT (.+) FROM inboxes WHERE id").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "project_id", "email", "created_at", "updated_at",
					}).AddRow(1, 1, "updated@example.com", now, now))
			},
			wantErr: false,
		},
		{
			name: "non-existent inbox",
			inbox: models.Inbox{
				Base:      models.Base{ID: 999},
				ProjectID: 1,
				Email:     "updated@example.com",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE inboxes").
					WithArgs("updated@example.com", 999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupInboxTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			_, err := repo.UpdateInbox(tt.inbox)
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

func TestRepository_DeleteInbox(t *testing.T) {
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
				mock.ExpectExec("DELETE FROM inboxes").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "non-existent inbox",
			id:   999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM inboxes").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupInboxTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.DeleteInbox(tt.id)
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

func TestRepository_ListInboxesByProject(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		projectID int
		limit     int
		offset    int
		mockFn    func(sqlmock.Sqlmock)
		want      []models.Inbox
		total     int
		wantErr   bool
	}{
		{
			name:      "successful list",
			projectID: 1,
			limit:     10,
			offset:    0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"id", "project_id", "email", "created_at", "updated_at",
				}).
					AddRow(1, 1, "inbox1@example.com", now, now).
					AddRow(2, 1, "inbox2@example.com", now, now)

				mock.ExpectQuery("SELECT (.+) FROM inboxes").
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			want: []models.Inbox{
				{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					ProjectID: 1,
					Email:     "inbox1@example.com",
				},
				{
					Base: models.Base{
						ID:        2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					ProjectID: 1,
					Email:     "inbox2@example.com",
				},
			},
			total:   2,
			wantErr: false,
		},
		{
			name:      "empty project",
			projectID: 2,
			limit:     10,
			offset:    0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(2).
					WillReturnRows(countRows)

				mock.ExpectQuery("SELECT (.+) FROM inboxes").
					WithArgs(2, 10, 0).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "project_id", "email", "created_at", "updated_at",
					}))
			},
			want:    nil,
			total:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupInboxTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListInboxesByProject(tt.projectID, tt.limit, tt.offset)
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
