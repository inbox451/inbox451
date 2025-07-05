package core

import (
	"context"
	"errors"
	"inbox451/internal/test"
	"io"
	"testing"
	"time"

	"inbox451/internal/test"

	"inbox451/internal/logger"
	"inbox451/internal/mocks"
	"inbox451/internal/models"
	"inbox451/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	null "github.com/volatiletech/null/v9"
)

func setupProjectTestCore(t *testing.T) (*Core, *mocks.Repository) {
	mockRepo := mocks.NewRepository(t)
	logger := logger.New(io.Discard, logger.DEBUG)

	core := &Core{
		Logger:     logger,
		Repository: mockRepo,
	}
	core.ProjectService = NewProjectService(core)

	return core, mockRepo
}

func TestProjectService_Create(t *testing.T) {
	tests := []struct {
		name    string
		project *models.Project
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful creation",
			project: &models.Project{
				Name: "Test Project",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateProject", mock.Anything, mock.AnythingOfType("*models.Project")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			project: &models.Project{
				Name: "Test Project",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("CreateProject", mock.Anything, mock.AnythingOfType("*models.Project")).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupProjectTestCore(t)
			tt.mockFn(mockRepo)

			err := core.ProjectService.Create(context.Background(), tt.project)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_Get(t *testing.T) {
	testProjectID1 := test.RandomTestUUID()
	testNonExistingProjectID := test.RandomTestUUID()
	now := time.Now()
	tests := []struct {
		name    string
		id      string
		mockFn  func(*mocks.Repository)
		want    *models.Project
		wantErr bool
		errType error
	}{
		{
			name: "existing project",
			id:   testProjectID1,
			mockFn: func(m *mocks.Repository) {
				m.On("GetProject", mock.Anything, testProjectID1).Return(&models.Project{
					Base: models.Base{
						ID:        testProjectID1,
						CreatedAt: null.TimeFrom(now),
						UpdatedAt: null.TimeFrom(now),
					},
					Name: "Test Project",
				}, nil)
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
			id:   testNonExistingProjectID,
			mockFn: func(m *mocks.Repository) {
				m.On("GetProject", mock.Anything, testNonExistingProjectID).Return(nil, storage.ErrNotFound)
			},
			want:    nil,
			wantErr: true,
			errType: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupProjectTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.ProjectService.Get(context.Background(), tt.id)
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

func TestProjectService_List(t *testing.T) {
	testProjectID1 := test.RandomTestUUID()
	testProjectID2 := test.RandomTestUUID()
	tests := []struct {
		name    string
		limit   int
		offset  int
		mockFn  func(*mocks.Repository)
		want    *models.PaginatedResponse
		wantErr bool
	}{
		{
			name:   "successful list",
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				projects := []*models.Project{
					{Base: models.Base{ID: testProjectID1}, Name: "Project 1"},
					{Base: models.Base{ID: testProjectID2}, Name: "Project 2"},
				}
				m.On("ListProjects", mock.Anything, 10, 0).Return(projects, 2, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.Project{
					{Base: models.Base{ID: testProjectID1}, Name: "Project 1"},
					{Base: models.Base{ID: testProjectID2}, Name: "Project 2"},
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
			name:   "repository error",
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListProjects", mock.Anything, 10, 0).Return([]*models.Project(nil), 0, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupProjectTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.ProjectService.List(context.Background(), tt.limit, tt.offset)
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

func TestProjectService_Update(t *testing.T) {
	testProjectID1 := test.RandomTestUUID()
	testNonExistingProjectID := test.RandomTestUUID()
	tests := []struct {
		name    string
		project *models.Project
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful update",
			project: &models.Project{
				Base: models.Base{ID: testProjectID1},
				Name: "Updated Project",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateProject", mock.Anything, mock.AnythingOfType("*models.Project")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "update non-existent project",
			project: &models.Project{
				Base: models.Base{ID: testNonExistingProjectID},
				Name: "Updated Project",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("UpdateProject", mock.Anything, mock.AnythingOfType("*models.Project")).
					Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupProjectTestCore(t)
			tt.mockFn(mockRepo)

			err := core.ProjectService.Update(context.Background(), tt.project)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_Delete(t *testing.T) {
	testProjectID1 := test.RandomTestUUID()
	testNonExistingProjectID := test.RandomTestUUID()
	tests := []struct {
		name    string
		id      string
		mockFn  func(*mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   testProjectID1,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteProject", mock.Anything, testProjectID1).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "delete non-existent project",
			id:   testNonExistingProjectID,
			mockFn: func(m *mocks.Repository) {
				m.On("DeleteProject", mock.Anything, testNonExistingProjectID).Return(storage.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupProjectTestCore(t)
			tt.mockFn(mockRepo)

			err := core.ProjectService.Delete(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_AddUser(t *testing.T) {
	testProjectID1 := test.RandomTestUUID()
	testUserID1 := test.RandomTestUUID()
	tests := []struct {
		name        string
		projectUser *models.ProjectUser
		mockFn      func(*mocks.Repository)
		wantErr     bool
	}{
		{
			name: "successful add user",
			projectUser: &models.ProjectUser{
				ProjectID: testProjectID1,
				UserID:    testUserID1,
				Role:      "member",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("ProjectAddUser", mock.Anything, mock.MatchedBy(func(pu *models.ProjectUser) bool {
					return pu.ProjectID == testProjectID1 && pu.UserID == testUserID1 && pu.Role == "member"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			projectUser: &models.ProjectUser{
				ProjectID: testProjectID1,
				UserID:    testUserID1,
				Role:      "member",
			},
			mockFn: func(m *mocks.Repository) {
				m.On("ProjectAddUser", mock.Anything, mock.Anything).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupProjectTestCore(t)
			tt.mockFn(mockRepo)

			err := core.ProjectService.AddUser(context.Background(), tt.projectUser)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_RemoveUser(t *testing.T) {
	testProjectID1 := test.RandomTestUUID()
	testUserID1 := test.RandomTestUUID()
	tests := []struct {
		name      string
		projectID string
		userID    string
		mockFn    func(*mocks.Repository)
		wantErr   bool
	}{
		{
			name:      "successful remove user",
			projectID: testProjectID1,
			userID:    testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("ProjectRemoveUser", mock.Anything, testProjectID1, testUserID1).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "repository error",
			projectID: testProjectID1,
			userID:    testUserID1,
			mockFn: func(m *mocks.Repository) {
				m.On("ProjectRemoveUser", mock.Anything, testProjectID1, testUserID1).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupProjectTestCore(t)
			tt.mockFn(mockRepo)

			err := core.ProjectService.RemoveUser(context.Background(), tt.projectID, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_ListByUser(t *testing.T) {
	testProjectID1 := test.RandomTestUUID()
	testProjectID2 := test.RandomTestUUID()
	testUserID1 := test.RandomTestUUID()
	now := time.Now()
	tests := []struct {
		name    string
		userID  string
		limit   int
		offset  int
		mockFn  func(*mocks.Repository)
		want    *models.PaginatedResponse
		wantErr bool
	}{
		{
			name:   "successful list",
			userID: testUserID1,
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				projects := []*models.Project{
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
				}
				m.On("ListProjectsByUser", mock.Anything, testUserID1, 10, 0).Return(projects, 2, nil)
			},
			want: &models.PaginatedResponse{
				Data: []*models.Project{
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
				Pagination: models.Pagination{
					Total:  2,
					Limit:  10,
					Offset: 0,
				},
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: testUserID1,
			limit:  10,
			offset: 0,
			mockFn: func(m *mocks.Repository) {
				m.On("ListProjectsByUser", mock.Anything, testUserID1, 10, 0).
					Return([]*models.Project(nil), 0, errors.New("database error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, mockRepo := setupProjectTestCore(t)
			tt.mockFn(mockRepo)

			got, err := core.ProjectService.ListByUser(context.Background(), tt.userID, tt.limit, tt.offset)
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
