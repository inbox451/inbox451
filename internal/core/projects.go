package core

import (
	"context"

	"inbox451/internal/models"
)

type ProjectService struct {
	core *Core
}

func NewProjectService(core *Core) ProjectService {
	return ProjectService{core: core}
}

func (s *ProjectService) List(ctx context.Context, limit, offset int) (*models.PaginatedResponse, error) {
	s.core.Logger.Debug("Listing projects with limit: %d and offset: %d", limit, offset)

	projects, total, err := s.core.Repository.ListProjects(ctx, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list projects: %v", err)
		return nil, err
	}

	response := &models.PaginatedResponse{
		Data: projects,
		Pagination: models.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	s.core.Logger.Debug("Successfully retrieved %d projects (total: %d)", len(projects), total)
	return response, nil
}

func (s *ProjectService) ListByUser(ctx context.Context, userID string, limit, offset int) (*models.PaginatedResponse, error) {
	s.core.Logger.Debug("Listing projects with limit: %d and offset: %d for user %d", limit, offset, userID)

	projects, total, err := s.core.Repository.ListProjectsByUser(ctx, userID, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list projects: %v", err)
		return nil, err
	}

	response := &models.PaginatedResponse{
		Data: projects,
		Pagination: models.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	s.core.Logger.Debug("Successfully retrieved %d projects (total: %d) for user %d", len(projects), total, userID)
	return response, nil
}

func (s *ProjectService) Get(ctx context.Context, projectId string) (*models.Project, error) {
	s.core.Logger.Debug("Fetching project with ID: %d", projectId)

	project, err := s.core.Repository.GetProject(ctx, projectId)
	if err != nil {
		s.core.Logger.Error("Failed to fetch project: %v", err)
		return nil, err
	}

	if project == nil {
		s.core.Logger.Info("Project not found with ID: %d", projectId)
		return nil, ErrNotFound
	}

	return project, nil
}

func (s *ProjectService) Create(ctx context.Context, project *models.Project) error {
	s.core.Logger.Info("Creating new project: %s", project.Name)

	if err := s.core.Repository.CreateProject(ctx, project); err != nil {
		s.core.Logger.Error("Failed to create project: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully created project with ID: %d", project.ID)
	return nil
}

func (s *ProjectService) Update(ctx context.Context, project *models.Project) error {
	s.core.Logger.Info("Updating project with ID: %d", project.ID)

	if err := s.core.Repository.UpdateProject(ctx, project); err != nil {
		s.core.Logger.Error("Failed to update project: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully updated project with ID: %d", project.ID)
	return nil
}

func (s *ProjectService) AddUser(ctx context.Context, projectUser *models.ProjectUser) error {
	s.core.Logger.Debug("Adding user %d to project %d with role=%s", projectUser.UserID, projectUser.ProjectID, projectUser.Role)

	if err := s.core.Repository.ProjectAddUser(ctx, projectUser); err != nil {
		s.core.Logger.Error("Failed to add user to project: %v", err)
		return err
	}

	s.core.Logger.Debug("Successfully added user %d to project %d", projectUser.ProjectID, projectUser.UserID)
	return nil
}

func (s *ProjectService) RemoveUser(ctx context.Context, projectID string, userID string) error {
	s.core.Logger.Debug("Remove user %d to project %d", userID, projectID)

	if err := s.core.Repository.ProjectRemoveUser(ctx, projectID, userID); err != nil {
		s.core.Logger.Error("Failed to add user to project: %v", err)
		return err
	}

	s.core.Logger.Debug("Successfully removed user %d from project %d", projectID, userID)
	return nil
}

func (s *ProjectService) Delete(ctx context.Context, id string) error {
	s.core.Logger.Info("Deleting project with ID: %d", id)

	if err := s.core.Repository.DeleteProject(ctx, id); err != nil {
		s.core.Logger.Error("Failed to delete project: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully deleted project with ID: %d", id)
	return nil
}
