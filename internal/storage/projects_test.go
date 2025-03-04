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
	mock.ExpectPrepare("SELECT (.+) FROM project_users WHERE project_id")

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

	getProjectUser, err := sqlxDB.Preparex("SELECT project_id, user_id, role, created_at, updated_at FROM project_users WHERE project_id = ? AND user_id = ?")
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
		GetProjectUser:        getProjectUser,
	}

	repo := &repository{
		db:      sqlxDB,
		queries: queries,
	}

	return repo, mock
}

func TestRepository_CreateProject(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		project models.Project
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful creation",
			project: models.Project{
				Name: "Test Project",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				// The implementation expects only an ID to be returned from INSERT
				mock.ExpectQuery("INSERT INTO projects").
					WithArgs("Test Project").
					WillReturnRows(
						sqlmock.NewRows([]string{"id"}).
							AddRow(1),
					)

				// Then the implementation calls GetProject to fetch the complete project
				rows := sqlmock.NewRows([]string{
					"id", "name", "created_at", "updated_at",
				}).AddRow(1, "Test Project", now, now)

				mock.ExpectQuery("SELECT (.+) FROM projects WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "database error",
			project: models.Project{
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

			got, err := repo.CreateProject(tt.project)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, got.ID)
			assert.NotZero(t, got.CreatedAt)
			assert.NotZero(t, got.UpdatedAt)
			assert.Equal(t, tt.project.Name, got.Name)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_GetProject(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		mockFn  func(sqlmock.Sqlmock)
		want    models.Project
		wantErr bool
		errType error
	}{
		{
			name: "existing project",
			id:   1,
			mockFn: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "created_at", "updated_at",
				}).AddRow(1, "Test Project", now, now)

				mock.ExpectQuery("SELECT (.+) FROM projects").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: models.Project{
				Base: models.Base{
					ID:        1,
					CreatedAt: null.TimeFrom(now),
					UpdatedAt: null.TimeFrom(now),
				},
				Name: "Test Project",
			},
			wantErr: false,
		},
		{
			name: "non-existent project",
			id:   999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM projects").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			want:    models.Project{},
			wantErr: true,
			errType: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, err := repo.GetProject(tt.id)
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

	tests := []struct {
		name    string
		project models.Project
		mockFn  func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful update",
			project: models.Project{
				Base: models.Base{ID: 1},
				Name: "Updated Project",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				// The implementation uses Exec, not Query
				mock.ExpectExec("UPDATE projects").
					WithArgs("Updated Project", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// After update, the implementation calls GetProject to get the updated project
				rows := sqlmock.NewRows([]string{
					"id", "name", "created_at", "updated_at",
				}).AddRow(1, "Updated Project", now, now)

				mock.ExpectQuery("SELECT (.+) FROM projects WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "non-existent project",
			project: models.Project{
				Base: models.Base{ID: 999},
				Name: "Updated Project",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("UPDATE projects").
					WithArgs("Updated Project", 999).
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

			got, err := repo.UpdateProject(tt.project)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, got.UpdatedAt)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_DeleteProject(t *testing.T) {
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
				mock.ExpectExec("DELETE FROM projects").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "non-existent project",
			id:   999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM projects").
					WithArgs(999).
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

			err := repo.DeleteProject(tt.id)
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

	tests := []struct {
		name    string
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []models.Project
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
					AddRow(1, "Project 1", now, now).
					AddRow(2, "Project 2", now, now)

				mock.ExpectQuery("SELECT (.+) FROM projects").
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			want: []models.Project{
				{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					Name: "Project 1",
				},
				{
					Base: models.Base{
						ID:        2,
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

				// Even with zero count, the implementation still calls Select
				// So we need to mock an empty result set
				rows := sqlmock.NewRows([]string{
					"id", "name", "created_at", "updated_at",
				})

				mock.ExpectQuery("SELECT (.+) FROM projects").
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			want:    nil,
			total:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListProjects(tt.limit, tt.offset)
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

	tests := []struct {
		name    string
		userID  int
		limit   int
		offset  int
		mockFn  func(sqlmock.Sqlmock)
		want    []models.Project
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
					"id", "name", "created_at", "updated_at",
				}).
					AddRow(1, "Project 1", now, now).
					AddRow(2, "Project 2", now, now)

				mock.ExpectQuery("SELECT (.+) FROM projects").
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			want: []models.Project{
				{
					Base: models.Base{
						ID:        1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					Name: "Project 1",
				},
				{
					Base: models.Base{
						ID:        2,
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
			userID: 2,
			limit:  10,
			offset: 0,
			mockFn: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(2).
					WillReturnRows(countRows)

				// Even with zero count, the implementation still calls Select
				// So we need to mock an empty result set
				rows := sqlmock.NewRows([]string{
					"id", "name", "created_at", "updated_at",
				})

				mock.ExpectQuery("SELECT (.+) FROM projects").
					WithArgs(2, 10, 0). // Fixed the arguments here - need 3 args: userID, limit, offset
					WillReturnRows(rows)
			},
			want:    nil,
			total:   0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupProjectTestDB(t)
			defer repo.db.Close()

			tt.mockFn(mock)

			got, total, err := repo.ListProjectsByUser(tt.userID, tt.limit, tt.offset)
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

	tests := []struct {
		name        string
		projectUser models.ProjectUser
		mockFn      func(sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name: "successful add",
			projectUser: models.ProjectUser{
				ProjectID: 1,
				UserID:    1,
				Role:      "member",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO project_users").
					WithArgs(1, 1, "member").
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Then mock the GetProjectUser call that happens afterward
				mock.ExpectQuery("SELECT (.+) FROM project_users").
					WithArgs(1, 1).
					WillReturnRows(
						sqlmock.NewRows([]string{"project_id", "user_id", "role", "created_at", "updated_at"}).
							AddRow(1, 1, "member", now, now),
					)
			},
			wantErr: false,
		},
		{
			name: "duplicate user in project",
			projectUser: models.ProjectUser{
				ProjectID: 1,
				UserID:    1,
				Role:      "member",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO project_users").
					WithArgs(1, 1, "member").
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

			got, err := repo.ProjectAddUser(tt.projectUser)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotZero(t, got.CreatedAt)
			assert.NotZero(t, got.UpdatedAt)
			assert.Equal(t, tt.projectUser.ProjectID, got.ProjectID)
			assert.Equal(t, tt.projectUser.UserID, got.UserID)
			assert.Equal(t, tt.projectUser.Role, got.Role)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRepository_ProjectRemoveUser(t *testing.T) {
	tests := []struct {
		name      string
		projectID int
		userID    int
		mockFn    func(sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "successful removal",
			projectID: 1,
			userID:    1,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM project_users").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "non-existent relationship",
			projectID: 999,
			userID:    999,
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM project_users").
					WithArgs(999, 999).
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

			err := repo.ProjectRemoveUser(tt.projectID, tt.userID)
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
