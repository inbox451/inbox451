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

func setupTestDB(t *testing.T) (*repository, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	mock.ExpectPrepare("SELECT (.+) FROM users")                // ListUsers
	mock.ExpectPrepare("SELECT COUNT(.+) FROM users")           // CountUsers
	mock.ExpectPrepare("SELECT (.+) FROM users WHERE id")       // GetUser
	mock.ExpectPrepare("SELECT (.+) FROM users WHERE username") // GetUserByUsername
	mock.ExpectPrepare("INSERT INTO users")                     // CreateUser
	mock.ExpectPrepare("UPDATE users")                          // UpdateUser
	mock.ExpectPrepare("DELETE FROM users")                     // DeleteUser

	listUsers, err := sqlxDB.Preparex("SELECT id, name, username, password, email, status, role, loggedin_at, created_at, updated_at FROM users ORDER BY id LIMIT ? OFFSET ?")
	require.NoError(t, err)

	countUsers, err := sqlxDB.Preparex("SELECT COUNT(*) FROM users")
	require.NoError(t, err)

	getUser, err := sqlxDB.Preparex("SELECT id, name, username, password, email, status, role, password_login, loggedin_at, created_at, updated_at FROM users WHERE id = ?")
	require.NoError(t, err)

	getUserByUsername, err := sqlxDB.Preparex("SELECT id, name, username, password, email, status, role, password_login, loggedin_at, created_at, updated_at FROM users WHERE username = ?")
	require.NoError(t, err)

	createUser, err := sqlxDB.Preparex("INSERT INTO users (name, username, password, email, status, role, password_login) VALUES (?, ?, ?, ?, ?, ?, ?)")
	require.NoError(t, err)

	updateUser, err := sqlxDB.Preparex("UPDATE users SET name = ?, username = ?, password = ?, email = ?, status = ?, role = ?, password_login = ? WHERE id = ?")
	require.NoError(t, err)

	deleteUser, err := sqlxDB.Preparex("DELETE FROM users WHERE id = ?")
	require.NoError(t, err)

	queries := &Queries{
		ListUsers:         listUsers,
		CountUsers:        countUsers,
		GetUser:           getUser,
		GetUserByUsername: getUserByUsername,
		CreateUser:        createUser,
		UpdateUser:        updateUser,
		DeleteUser:        deleteUser,
	}

	repo := &repository{
		db:      sqlxDB,
		queries: queries,
	}

	return repo, mock
}

func TestRepository_ListUsers(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []*models.User
		total   int
		wantErr bool
	}{
		{
			name:   "successful list",
			limit:  10,
			offset: 0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"id", "name", "username", "password", "email",
					"status", "role", "loggedin_at", "created_at", "updated_at",
				}).
					AddRow(1, "User 1", "user1", null.StringFrom("hash1"), "user1@example.com",
						"active", "user", nil, now, now).
					AddRow(2, "User 2", "user2", null.StringFrom("hash2"), "user2@example.com",
						"active", "user", nil, now, now)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			want: []*models.User{
				{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					Name:     "User 1",
					Username: "user1",
					Password: null.StringFrom("hash1"),
					Email:    "user1@example.com",
					Status:   "active",
					Role:     "user",
				},
				{
					Base: models.Base{
						ID:        2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					Name:     "User 2",
					Username: "user2",
					Password: null.StringFrom("hash2"),
					Email:    "user2@example.com",
					Status:   "active",
					Role:     "user",
				},
			},
			total:   2,
			wantErr: false,
		},
		{
			name:   "database error",
			limit:  10,
			offset: 0,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(sql.ErrConnDone)
			},
			want:    nil,
			total:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListUsers(context.Background(), tt.limit, tt.offset)
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

func TestRepository_GetUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		userID  int
		mockFn  func(sqlmock.Sqlmock)
		want    *models.User
		wantErr bool
		errType error
	}{
		{
			name:   "existing user",
			userID: 1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "username", "password", "email",
					"status", "role", "password_login", "loggedin_at",
					"created_at", "updated_at",
				}).AddRow(
					1, "Test User", "testuser", null.StringFrom("hash"), "test@example.com",
					"active", "user", true, nil,
					now, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: &models.User{
				Base: models.Base{
					ID:        1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				Name:          "Test User",
				Username:      "testuser",
				Password:      null.StringFrom("hash"),
				Email:         "test@example.com",
				Status:        "active",
				Role:          "user",
				PasswordLogin: true,
			},
			wantErr: false,
		},
		{
			name:   "non-existent user",
			userID: 999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users").
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
			repo, mock := setupTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.GetUser(context.Background(), tt.userID)
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

func TestRepository_CreateUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		user    *models.User
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful creation",
			user: &models.User{
				Name:          "Test User",
				Username:      "testuser",
				Password:      null.StringFrom("hash"),
				Email:         "test@example.com",
				Status:        "active",
				Role:          "user",
				PasswordLogin: true,
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(
						"Test User",
						"testuser",
						"hash",
						"test@example.com",
						"active",
						"user",
						true,
					).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
							AddRow(1, now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "duplicate username",
			user: &models.User{
				Name:          "Test User",
				Username:      "existing",
				Password:      null.StringFrom("hash"),
				Email:         "test@example.com",
				Status:        "active",
				Role:          "user",
				PasswordLogin: true,
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(
						"Test User",
						"existing",
						"hash",
						"test@example.com",
						"active",
						"user",
						true,
					).
					WillReturnError(sql.ErrConnDone) // Using generic error for simplicity
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.CreateUser(context.Background(), tt.user)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, tt.user.ID)
			assert.NotZero(t, tt.user.CreatedAt)
			assert.NotZero(t, tt.user.UpdatedAt)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_UpdateUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		user    *models.User
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful update",
			user: &models.User{
				Base: models.Base{
					ID: 1,
				},
				Name:          "Updated User",
				Username:      "updateduser",
				Password:      null.StringFrom("newhash"),
				Email:         "updated@example.com",
				Status:        "active",
				Role:          "admin",
				PasswordLogin: true,
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("UPDATE users").
					WithArgs(
						"Updated User",
						"updateduser",
						"newhash",
						"updated@example.com",
						"active",
						"admin",
						true,
						1,
					).
					WillReturnRows(
						sqlmock.NewRows([]string{"updated_at"}).
							AddRow(now),
					)
			},
			wantErr: false,
		},
		{
			name: "non-existent user",
			user: &models.User{
				Base: models.Base{
					ID: 999,
				},
				Name:          "Updated User",
				Username:      "updateduser",
				Password:      null.StringFrom("newhash"),
				Email:         "updated@example.com",
				Status:        "active",
				Role:          "admin",
				PasswordLogin: true,
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("UPDATE users").
					WithArgs(
						"Updated User",
						"updateduser",
						"newhash",
						"updated@example.com",
						"active",
						"admin",
						true,
						999,
					).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.UpdateUser(context.Background(), tt.user)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, tt.user.UpdatedAt)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_DeleteUser(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "successful deletion",
			userID: 1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:   "non-existent user",
			userID: 999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.DeleteUser(context.Background(), tt.userID)
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

func TestRepository_GetUserByUsername(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		username string
		mockFn   func(sqlmock.Sqlmock)
		want     *models.User
		wantErr  bool
	}{
		{
			name:     "existing user",
			username: "testuser",
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "username", "password", "email",
					"status", "role", "password_login", "loggedin_at",
					"created_at", "updated_at",
				}).AddRow(
					1, "Test User", "testuser", null.StringFrom("hash"), "test@example.com",
					"active", "user", true, nil,
					now, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM users WHERE username").
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			want: &models.User{
				Base: models.Base{
					ID:        1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				Name:          "Test User",
				Username:      "testuser",
				Password:      null.StringFrom("hash"),
				Email:         "test@example.com",
				Status:        "active",
				Role:          "user",
				PasswordLogin: true,
			},
			wantErr: false,
		},
		{
			name:     "non-existent user",
			username: "nonexistent",
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE username").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:     "database error",
			username: "testuser",
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE username").
					WithArgs("testuser").
					WillReturnError(sql.ErrConnDone)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.GetUserByUsername(context.Background(), tt.username)
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
