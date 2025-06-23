package storage

import (
	"context"
	"database/sql"
	"inbox451/internal/test"
	"testing"
	"time"

	"inbox451/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	null "github.com/volatiletech/null/v9"
)

func setupProjectTestDB(t *testing.T) (*repository, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	mock.ExpectPrepare("SELECT (.+) FROM projects")                               // ListProjects
	mock.ExpectPrepare("SELECT COUNT(.+) FROM projects")                          // CountProjects
	mock.ExpectPrepare("SELECT (.+) FROM projects WHERE id")                      // GetProject
	mock.ExpectPrepare("INSERT INTO projects")                                    // CreateProject
	mock.ExpectPrepare("UPDATE projects")                                         // UpdateProject
	mock.ExpectPrepare("DELETE FROM projects")                                    // DeleteProject
	mock.ExpectPrepare("INSERT INTO project_users")                               // AddUserToProject
	mock.ExpectPrepare("DELETE FROM project_users")                               // RemoveUserFromProject
	mock.ExpectPrepare("SELECT (.+) FROM projects INNER JOIN project_users")      // ListProjectsByUser
	mock.ExpectPrepare("SELECT COUNT(.+) FROM projects INNER JOIN project_users") // CountProjectsByUser

	listProjects, err := sqlxDB.Preparex("SELECT id, name, created_at, updated_at FROM projects ORDER BY id LIMIT ? OFFSET ?")
	require.NoError(t, err)

	countProjects, err := sqlxDB.Preparex("SELECT COUNT(*) FROM projects")
	require.NoError(t, err)

	getProject, err := sqlxDB.Preparex("SELECT id, name, created_at, updated_at FROM projects WHERE id = ?")
	require.NoError(t, err)

	createProject, err := sqlxDB.Preparex("INSERT INTO projects (name) VALUES (?)")
	require.NoError(t, err)

	updateProject, err := sqlxDB.Preparex("UPDATE projects SET name = ? WHERE id = ?")
	require.NoError(t, err)

	deleteProject, err := sqlxDB.Preparex("DELETE FROM projects WHERE id = ?")
	require.NoError(t, err)

	addUserToProject, err := sqlxDB.Preparex("INSERT INTO project_users (user_id, project_id, role) VALUES (?, ?, ?)")
	require.NoError(t, err)

	removeUserFromProject, err := sqlxDB.Preparex("DELETE FROM project_users WHERE user_id = ? AND project_id = ?")
	require.NoError(t, err)

	listProjectsByUser, err := sqlxDB.Preparex("SELECT projects.id, projects.name, projects.created_at, projects.updated_at FROM projects INNER JOIN project_users ON projects.id = project_users.project_id WHERE project_users.user_id = ? ORDER BY projects.id LIMIT ? OFFSET ?")
	require.NoError(t, err)

	countProjectsByUser, err := sqlxDB.Preparex("SELECT COUNT(DISTINCT(projects.id)) FROM projects INNER JOIN project_users ON projects.id = project_users.project_id WHERE project_users.user_id = ?")
	require.NoError(t, err)

	queries := &Queries{
		ListProjects:          listProjects,
		CountProjects:         countProjects,
		GetProject:            getProject,
		CreateProject:         createProject,
		UpdateProject:         updateProject,
		DeleteProject:         deleteProject,
		AddUserToProject:      addUserToProject,
		RemoveUserFromProject: removeUserFromProject,
		ListProjectsByUser:    listProjectsByUser,
		CountProjectsByUser:   countProjectsByUser,
	}

	repo := &repository{
		db:      sqlxDB,
		queries: queries,
	}

	return repo, mock
}

func TestRepository_CreateProject(t *testing.T) {
	now := time.Now()
	testProjectID1 := test.RandomTestUUID()

	tests := []struct {
		name    string
		project *models.Project
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful creation",
			project: &models.Project{
				Name: "Test Project",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO projects").
					WithArgs("Test Project").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
							AddRow(testProjectID1, now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "database error",
			project: &models.Project{
				Name: "Test Project",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO projects").
					WithArgs("Test Project").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.CreateProject(context.Background(), tt.project)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, tt.project.ID)
			assert.NotZero(t, tt.project.CreatedAt)
			assert.NotZero(t, tt.project.UpdatedAt)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetProject(t *testing.T) {
	now := time.Now()
	testProjectID1 := test.RandomTestUUID()
	nonExistingProjectID := test.RandomTestUUID()

	tests := []struct {
		name    string
		id      string
		mockFn  func(sqlmock.Sqlmock)
		want    *models.Project
		wantErr bool
		errType error
	}{
		{
			name: "existing project",
			id:   testProjectID1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "created_at", "updated_at",
				}).AddRow(testProjectID1, "Test Project", now, now)

				mock.ExpectQuery("SELECT (.+) FROM projects").
					WithArgs(testProjectID1).
					WillReturnRows(rows)
			},
			want: &models.Project{
				Base: models.Base{
					ID:        testProjectID1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				Name: "Test Project",
			},
			wantErr: false,
		},
		{
			name: "non-existent project",
			id:   nonExistingProjectID,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM projects").
					WithArgs(nonExistingProjectID).
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
			errType: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.GetProject(context.Background(), tt.id)
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

func TestRepository_UpdateProject(t *testing.T) {
	now := time.Now()
	testProjectID1 := test.RandomTestUUID()
	nonExistingProjectID := test.RandomTestUUID()

	tests := []struct {
		name    string
		project *models.Project
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful update",
			project: &models.Project{
				Base: models.Base{ID: testProjectID1},
				Name: "Updated Project",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("UPDATE projects").
					WithArgs("Updated Project", testProjectID1).
					WillReturnRows(
						sqlmock.NewRows([]string{"updated_at"}).
							AddRow(now),
					)
			},
			wantErr: false,
		},
		{
			name: "non-existent project",
			project: &models.Project{
				Base: models.Base{ID: nonExistingProjectID},
				Name: "Updated Project",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("UPDATE projects").
					WithArgs("Updated Project", nonExistingProjectID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.UpdateProject(context.Background(), tt.project)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, tt.project.UpdatedAt)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_DeleteProject(t *testing.T) {
	testProjectID1 := test.RandomTestUUID()
	nonExistingProjectID := test.RandomTestUUID()
	tests := []struct {
		name    string
		id      string
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   testProjectID1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM projects").
					WithArgs(testProjectID1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "non-existent project",
			id:   nonExistingProjectID,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM projects").
					WithArgs(nonExistingProjectID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.DeleteProject(context.Background(), tt.id)
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

func TestRepository_ListProjects(t *testing.T) {
	now := time.Now()
	testProjectID1 := test.RandomTestUUID()
	testProjectID2 := test.RandomTestUUID()

	tests := []struct {
		name    string
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []*models.Project
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
					"id", "name", "created_at", "updated_at",
				}).
					AddRow(testProjectID1, "Project 1", now, now).
					AddRow(testProjectID2, "Project 2", now, now)

				mock.ExpectQuery("SELECT (.+) FROM projects").
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			want: []*models.Project{
				{
					Base: models.Base{
						ID:        testProjectID1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					Name: "Project 1",
				},
				{
					Base: models.Base{
						ID:        testProjectID2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					Name: "Project 2",
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
			want:    []*models.Project{},
			total:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListProjects(context.Background(), tt.limit, tt.offset)
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

func TestRepository_ListProjectsByUser(t *testing.T) {
	now := time.Now()
	testProjectID1 := test.RandomTestUUID()
	testProjectID2 := test.RandomTestUUID()
	testUserID1 := test.RandomTestUUID()
	testUserID2 := test.RandomTestUUID()

	tests := []struct {
		name    string
		userID  string
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []*models.Project
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
					"id", "name", "created_at", "updated_at",
				}).
					AddRow(testProjectID1, "Project 1", now, now).
					AddRow(testProjectID2, "Project 2", now, now)

				mock.ExpectQuery("SELECT (.+) FROM projects").
					WithArgs(testUserID1, 10, 0).
					WillReturnRows(rows)
			},
			want: []*models.Project{
				{
					Base: models.Base{
						ID:        testProjectID1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					Name: "Project 1",
				},
				{
					Base: models.Base{
						ID:        testProjectID2,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					Name: "Project 2",
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
			want:    []*models.Project{},
			total:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListProjectsByUser(context.Background(), tt.userID, tt.limit, tt.offset)
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

func TestRepository_ProjectAddUser(t *testing.T) {
	now := time.Now()
	testProjectID1 := test.RandomTestUUID()
	testUserID1 := test.RandomTestUUID()

	tests := []struct {
		name        string
		projectUser *models.ProjectUser
		mockFn      func(sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name: "successful add",
			projectUser: &models.ProjectUser{
				ProjectID: testProjectID1,
				UserID:    testUserID1,
				Role:      "member",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO project_users").
					WithArgs(testProjectID1, testUserID1, "member").
					WillReturnRows(
						sqlmock.NewRows([]string{"created_at", "updated_at"}).
							AddRow(now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "duplicate user in project",
			projectUser: &models.ProjectUser{
				ProjectID: testProjectID1,
				UserID:    testUserID1,
				Role:      "member",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO project_users").
					WithArgs(testProjectID1, testUserID1, "member").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.ProjectAddUser(context.Background(), tt.projectUser)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, tt.projectUser.CreatedAt)
			assert.NotZero(t, tt.projectUser.UpdatedAt)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_ProjectRemoveUser(t *testing.T) {
	testProjectID1 := test.RandomTestUUID()
	testUserID1 := test.RandomTestUUID()
	testUserID2 := test.RandomTestUUID()
	nonExistingProjectID := test.RandomTestUUID()
	tests := []struct {
		name      string
		projectID string
		userID    string
		mockFn    func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "successful removal",
			projectID: testProjectID1,
			userID:    testUserID1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM project_users").
					WithArgs(testProjectID1, testUserID1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "non-existent relationship",
			projectID: nonExistingProjectID,
			userID:    testUserID2,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM project_users").
					WithArgs(nonExistingProjectID, testUserID2).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			err := repo.ProjectRemoveUser(context.Background(), tt.projectID, tt.userID)
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
