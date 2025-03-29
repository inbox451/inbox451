package api

import (
	"errors"
	"inbox451/internal/storage"
	"net/http"
	"strconv"

	"inbox451/internal/models"

	"github.com/labstack/echo/v4"
)

func (s *Server) createProject(c echo.Context) error {
	var project models.Project
	if err := c.Bind(&project); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := c.Validate(&project); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	newProject, err := s.core.ProjectService.Create(project)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, newProject)
}

func (s *Server) getProjects(c echo.Context) error {
	var query models.PaginationQuery
	if err := c.Bind(&query); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if query.Limit == 0 {
		query.Limit = 10
	}

	if err := c.Validate(&query); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	response, err := s.core.ProjectService.List(query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getProjectsByUser(c echo.Context) error {
	userID, _ := strconv.Atoi(c.Param("userId"))

	var query models.PaginationQuery
	if err := c.Bind(&query); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if query.Limit == 0 {
		query.Limit = 10
	}

	if err := c.Validate(&query); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	response, err := s.core.ProjectService.ListByUser(userID, query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getProject(c echo.Context) error {
	projectID, _ := strconv.Atoi(c.Param("projectId"))
	project, err := s.core.ProjectService.Get(projectID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return s.core.HandleError(nil, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, project)
}

func (s *Server) updateProject(c echo.Context) error {
	projectID, _ := strconv.Atoi(c.Param("projectId"))
	var project models.Project
	if err := c.Bind(&project); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}
	project.ID = projectID

	if err := c.Validate(&project); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	updatedProject, err := s.core.ProjectService.Update(project)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, updatedProject)
}

func (s *Server) projectAddUser(c echo.Context) error {
	projectID, _ := strconv.Atoi(c.Param("projectId"))

	s.core.Logger.Info("hello")

	var projectUser models.ProjectUser
	if err := c.Bind(&projectUser); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}
	projectUser.ProjectID = projectID

	if err := c.Validate(&projectUser); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	newProjectUser, err := s.core.ProjectService.AddUser(projectUser)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, newProjectUser)
}

func (s *Server) deleteProject(c echo.Context) error {
	projectID, _ := strconv.Atoi(c.Param("projectId"))
	if err := s.core.ProjectService.Delete(projectID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) projectRemoveUser(c echo.Context) error {
	projectID, _ := strconv.Atoi(c.Param("projectId"))
	userID, _ := strconv.Atoi(c.Param("userId"))
	if err := s.core.ProjectService.RemoveUser(projectID, userID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}
