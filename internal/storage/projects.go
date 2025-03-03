package storage

import (
	"inbox451/internal/models"
)

func (r *repository) ListProjects(limit int, offset int) ([]models.Project, int, error) {
	var projects []models.Project
	var total int

	err := r.queries.CountProjects.Get(&total)
	if err != nil {
		return nil, 0, err
	}

	err = r.queries.ListProjects.Select(&projects, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *repository) ListProjectsByUser(userId int, limit int, offset int) ([]models.Project, int, error) {
	var total int
	var projects []models.Project

	err := r.queries.CountProjectsByUser.Get(&total, userId)
	if err != nil {
		return nil, 0, err
	}

	err = r.queries.ListProjectsByUser.Select(&projects, userId, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *repository) GetProject(id int) (models.Project, error) {
	var project models.Project
	err := r.queries.GetProject.Get(&project, id)
	return project, handleDBError(err)
}

func (r *repository) CreateProject(project models.Project) (models.Project, error) {
	var projectId int
	err := r.queries.CreateProject.QueryRow(project.Name).Scan(&projectId)
	if err != nil {
		return models.Project{}, handleDBError(err)
	}
	return r.GetProject(projectId)
}

func (r *repository) UpdateProject(project models.Project) (models.Project, error) {
	res, err := r.queries.UpdateProject.Exec(project.Name, project.ID)
	if err != nil {
		return models.Project{}, handleDBError(err)
	}
	if err := handleRowsAffected(res); err != nil {
		return models.Project{}, err
	}
	return r.GetProject(project.ID)
}

func (r *repository) GetProjectUser(projectId int, userId int) (models.ProjectUser, error) {
	var projectUser models.ProjectUser
	err := r.queries.GetProjectUser.Get(&projectUser, projectId, userId)
	return projectUser, handleDBError(err)
}

func (r *repository) ProjectAddUser(projectUser models.ProjectUser) (models.ProjectUser, error) {
	res, err := r.queries.AddUserToProject.Exec(
		projectUser.ProjectID,
		projectUser.UserID,
		projectUser.Role)

	if err != nil {
		return models.ProjectUser{}, handleDBError(err)
	}

	if err := handleRowsAffected(res); err != nil {
		return models.ProjectUser{}, err
	}

	return r.GetProjectUser(projectUser.ProjectID, projectUser.UserID)
}

func (r *repository) DeleteProject(id int) error {
	result, err := r.queries.DeleteProject.Exec(id)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}

func (r *repository) ProjectRemoveUser(projectId int, userId int) error {
	result, err := r.queries.RemoveUserFromProject.Exec(projectId, userId)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}
