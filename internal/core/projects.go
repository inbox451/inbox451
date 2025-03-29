package core

import (
	"inbox451/internal/models"
)

type ProjectService struct {
	core *Core
}

func NewProjectService(core *Core) ProjectService {
	return ProjectService{core: core}
}

func (s *ProjectService) List(limit, offset int) (models.PaginatedResponse, error) {
	s.core.Logger.Debug("Listing projects with limit: %d and offset: %d", limit, offset)

	projects, total, err := s.core.Repository.ListProjects(limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list projects: %v", err)
		return models.PaginatedResponse{}, err
	}

	response := models.PaginatedResponse{
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

func (s *ProjectService) ListByUser(userID int, limit, offset int) (models.PaginatedResponse, error) {
	s.core.Logger.Debug("Listing projects with limit: %d and offset: %d for user %d", limit, offset, userID)

	projects, total, err := s.core.Repository.ListProjectsByUser(userID, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list projects: %v", err)
		return models.PaginatedResponse{}, err
	}

	response := models.PaginatedResponse{
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

func (s *ProjectService) Get(projectId int) (models.Project, error) {
	s.core.Logger.Debug("Fetching project with ID: %d", projectId)

	project, err := s.core.Repository.GetProject(projectId)
	if err != nil {
		s.core.Logger.Error("Failed to fetch project: %v", err)
		return project, err
	}

	return project, nil
}

func (s *ProjectService) Create(project models.Project) (models.Project, error) {
	s.core.Logger.Info("Creating new project: %s", project.Name)

	project, err := s.core.Repository.CreateProject(project)
	if err != nil {
		s.core.Logger.Error("Failed to create project: %v", err)
		return project, err
	}

	s.core.Logger.Info("Successfully created project with ID: %d", project.ID)
	return project, nil
}

func (s *ProjectService) Update(project models.Project) (models.Project, error) {
	s.core.Logger.Info("Updating project with ID: %d", project.ID)

	project, err := s.core.Repository.UpdateProject(project)
	if err != nil {
		s.core.Logger.Error("Failed to update project: %v", err)
		return project, err
	}

	s.core.Logger.Info("Successfully updated project with ID: %d", project.ID)
	return project, nil
}

func (s *ProjectService) AddUser(projectUser models.ProjectUser) (models.ProjectUser, error) {
	s.core.Logger.Debug("Adding user %d to project %d with role=%s", projectUser.UserID, projectUser.ProjectID, projectUser.Role)

	projectUser, err := s.core.Repository.ProjectAddUser(projectUser)
	if err != nil {
		s.core.Logger.Error("Failed to add user to project: %v", err)
		return projectUser, err
	}

	s.core.Logger.Debug("Successfully added user %d to project %d", projectUser.ProjectID, projectUser.UserID)
	return projectUser, nil
}

func (s *ProjectService) RemoveUser(projectID int, userID int) error {
	s.core.Logger.Debug("Remove user %d to project %d", userID, projectID)

	err := s.core.Repository.ProjectRemoveUser(projectID, userID)
	if err != nil {
		s.core.Logger.Error("Failed to remove user from project: %v", err)
		return err
	}

	s.core.Logger.Debug("Successfully removed user %d from project %d", projectID, userID)
	return nil
}

func (s *ProjectService) Delete(id int) error {
	s.core.Logger.Info("Deleting project with ID: %d", id)

	if err := s.core.Repository.DeleteProject(id); err != nil {
		s.core.Logger.Error("Failed to delete project: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully deleted project with ID: %d", id)
	return nil
}
