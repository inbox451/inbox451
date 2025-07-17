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
	testUserID1 := test.RandomTestUUID()
	testTokenID1 := test.RandomTestUUID()

	tests := []struct {
		name    string
		token   *models.Token
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful creation",
			token: &models.Token{
				UserID:    testUserID1,
				Token:     "test-token",
				Name:      "Test Token",
				ExpiresAt: null.TimeFrom(expiresAt),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO tokens").
					WithArgs(testUserID1, "test-token", "Test Token", expiresAt).
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id", "user_id", "token", "name",
							"expires_at", "created_at", "updated_at",
						}).AddRow(testTokenID1, testUserID1, "test-token", "Test Token", expiresAt, now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "database error",
			token: &models.Token{
				UserID:    testUserID1,
				Token:     "test-token",
				Name:      "Test Token",
				ExpiresAt: null.TimeFrom(expiresAt),
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO tokens").
					WithArgs(testUserID1, "test-token", "Test Token", expiresAt).
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

			err := repo.CreateToken(context.Background(), tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, tt.token.ID)
			assert.NotZero(t, tt.token.CreatedAt)
			assert.NotZero(t, tt.token.UpdatedAt)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetTokenByUser(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	testUserID1 := test.RandomTestUUID()
	testTokenID1 := test.RandomTestUUID()
	nonExistingTokenID := test.RandomTestUUID()

	tests := []struct {
		name    string
		tokenID string
		userID  string
		mockFn  func(sqlmock.Sqlmock)
		want    *models.Token
		wantErr bool
		errType error
	}{
		{
			name:    "existing token",
			tokenID: testTokenID1,
			userID:  testUserID1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "token", "name",
					"expires_at", "created_at", "updated_at",
				}).AddRow(testTokenID1, testUserID1, "test-token", "Test Token", expiresAt, now, now)

				mock.ExpectQuery("SELECT (.+) FROM tokens").
					WithArgs(testTokenID1, testUserID1).
					WillReturnRows(rows)
			},
			want: &models.Token{
				Base: models.Base{
					ID:        testTokenID1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				UserID:    testUserID1,
				Token:     "test-token",
				Name:      "Test Token",
				ExpiresAt: null.TimeFrom(expiresAt),
			},
			wantErr: false,
		},
		{
			name:    "non-existent token",
			tokenID: nonExistingTokenID,
			userID:  testUserID1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM tokens").
					WithArgs(nonExistingTokenID, testUserID1).
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

			got, err := repo.GetTokenByUser(context.Background(), tt.tokenID, tt.userID)
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
	testTokenID1 := test.RandomTestUUID()
	nonExistingTokenID := test.RandomTestUUID()
	tests := []struct {
		name    string
		tokenID string
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:    "successful deletion",
			tokenID: testTokenID1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM tokens").
					WithArgs(testTokenID1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:    "non-existent token",
			tokenID: nonExistingTokenID,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM tokens").
					WithArgs(nonExistingTokenID).
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

			err := repo.DeleteToken(context.Background(), tt.tokenID)
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
	testUserID1 := test.RandomTestUUID()
	testUserID2 := test.RandomTestUUID()
	testTokenID1 := test.RandomTestUUID()
	testTokenID2 := test.RandomTestUUID()

	tests := []struct {
		name    string
		userID  string
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []*models.Token
		total   int
		wantErr bool
	}{
		{
			name:   "successful list",
			userID: testUserID1,
			limit:  10,
			offset: 0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(testUserID1).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"id", "user_id", "token", "name",
					"expires_at", "created_at", "updated_at",
				}).
					AddRow(testTokenID1, testUserID1, "token1", "Token 1", expiresAt, now, now).
					AddRow(testTokenID2, testUserID1, "token2", "Token 2", expiresAt, now, now)

				mock.ExpectQuery("SELECT (.+) FROM tokens").
					WithArgs(testUserID1, 10, 0).
					WillReturnRows(rows)
			},
			want: []*models.Token{
				{
					Base: models.Base{
						ID:        testTokenID1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					UserID:    testUserID1,
					Token:     "token1",
					Name:      "Token 1",
					ExpiresAt: null.TimeFrom(expiresAt),
				},
				{
					Base: models.Base{
						ID:        testTokenID2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					UserID:    testUserID1,
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
			userID: testUserID2,
			limit:  10,
			offset: 0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(testUserID2).
					WillReturnRows(countRows)
			},
			want:    []*models.Token{},
			total:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupTokenTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListTokensByUser(context.Background(), tt.userID, tt.limit, tt.offset)
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
