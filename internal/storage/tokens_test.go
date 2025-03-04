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

func setupTokenTestDB(t *testing.T) (*repository, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	mock.ExpectPrepare("SELECT (.+) FROM tokens WHERE user_id")      // ListTokensByUser
	mock.ExpectPrepare("SELECT COUNT(.+) FROM tokens WHERE user_id") // CountTokensByUser
	mock.ExpectPrepare("SELECT (.+) FROM tokens WHERE id")           // GetTokenByUser
	mock.ExpectPrepare("INSERT INTO tokens")                         // CreateToken
	mock.ExpectPrepare("DELETE FROM tokens")                         // DeleteToken

	listTokens, err := sqlxDB.Preparex("SELECT id, user_id, token, name, expires_at, created_at, updated_at FROM tokens WHERE user_id = ? ORDER BY id LIMIT ? OFFSET ?")
	require.NoError(t, err)

	countTokens, err := sqlxDB.Preparex("SELECT COUNT(1) FROM tokens WHERE user_id = ?")
	require.NoError(t, err)

	getToken, err := sqlxDB.Preparex("SELECT id, user_id, token, name, expires_at, created_at, updated_at FROM tokens WHERE id = ? AND user_id = ?")
	require.NoError(t, err)

	createToken, err := sqlxDB.Preparex("INSERT INTO tokens (user_id, token, name, expires_at) VALUES (?, ?, ?, ?)")
	require.NoError(t, err)

	deleteToken, err := sqlxDB.Preparex("DELETE FROM tokens WHERE id = ?")
	require.NoError(t, err)

	queries := &Queries{
		ListTokensByUser:  listTokens,
		CountTokensByUser: countTokens,
		GetTokenByUser:    getToken,
		CreateToken:       createToken,
		DeleteToken:       deleteToken,
	}

	repo := &repository{
		db:      sqlxDB,
		queries: queries,
	}

	return repo, mock
}

func TestRepository_CreateToken(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		token   models.Token
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful creation",
			token: models.Token{
				UserID:    1,
				Token:     "test-token",
				Name:      "Test Token",
				ExpiresAt: null.TimeFrom(expiresAt),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO tokens").
					WithArgs(1, "test-token", "Test Token", expiresAt).
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id", "user_id", "token", "name",
							"expires_at", "created_at", "updated_at",
						}).AddRow(1, 1, "test-token", "Test Token", expiresAt, now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "database error",
			token: models.Token{
				UserID:    1,
				Token:     "test-token",
				Name:      "Test Token",
				ExpiresAt: null.TimeFrom(expiresAt),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO tokens").
					WithArgs(1, "test-token", "Test Token", expiresAt).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupTokenTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.CreateToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, got.ID)
			assert.NotZero(t, got.CreatedAt)
			assert.NotZero(t, got.UpdatedAt)
			assert.Equal(t, tt.token.Name, got.Name)
			assert.Equal(t, tt.token.Token, got.Token)
			assert.Equal(t, tt.token.UserID, got.UserID)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetTokenByUser(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		tokenID int
		userID  int
		mockFn  func(sqlmock.Sqlmock)
		want    *models.Token
		wantErr bool
		errType error
	}{
		{
			name:    "existing token",
			tokenID: 1,
			userID:  1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "token", "name",
					"expires_at", "created_at", "updated_at",
				}).AddRow(1, 1, "test-token", "Test Token", expiresAt, now, now)

				mock.ExpectQuery("SELECT (.+) FROM tokens").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			want: &models.Token{
				Base: models.Base{
					ID:        1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				UserID:    1,
				Token:     "test-token",
				Name:      "Test Token",
				ExpiresAt: null.TimeFrom(expiresAt),
			},
			wantErr: false,
		},
		{
			name:    "non-existent token",
			tokenID: 999,
			userID:  1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM tokens").
					WithArgs(999, 1).
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
			errType: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupTokenTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.GetTokenByUser(tt.tokenID, tt.userID)
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

func TestRepository_DeleteToken(t *testing.T) {
	tests := []struct {
		name    string
		tokenID int
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:    "successful deletion",
			tokenID: 1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM tokens").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:    "non-existent token",
			tokenID: 999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM tokens").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupTokenTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.DeleteToken(tt.tokenID)
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

func TestRepository_ListTokensByUser(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		userID  int
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []models.Token
		total   int
		wantErr bool
	}{
		{
			name:   "successful list",
			userID: 1,
			limit:  10,
			offset: 0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(1).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"id", "user_id", "token", "name",
					"expires_at", "created_at", "updated_at",
				}).
					AddRow(1, 1, "token1", "Token 1", expiresAt, now, now).
					AddRow(2, 1, "token2", "Token 2", expiresAt, now, now)

				mock.ExpectQuery("SELECT (.+) FROM tokens").
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			want: []models.Token{
				{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					UserID:    1,
					Token:     "token1",
					Name:      "Token 1",
					ExpiresAt: null.TimeFrom(expiresAt),
				},
				{
					Base: models.Base{
						ID:        2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					UserID:    1,
					Token:     "token2",
					Name:      "Token 2",
					ExpiresAt: null.TimeFrom(expiresAt),
				},
			},
			total:   2,
			wantErr: false,
		},
		{
			name:   "empty list",
			userID: 2,
			limit:  10,
			offset: 0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(2).
					WillReturnRows(countRows)
			},
			want:    []models.Token{},
			total:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupTokenTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListTokensByUser(tt.userID, tt.limit, tt.offset)
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
