package api

import (
	"inbox451/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) createProject(c echo.Context) error {
	ctx := c.Request().Context()

	var project models.Project
	if err := c.Bind(&project); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := c.Validate(&project); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := s.core.ProjectService.Create(ctx, &project); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, project)
}

func (s *Server) getProjects(c echo.Context) error {
	ctx := c.Request().Context()

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

	response, err := s.core.ProjectService.List(ctx, query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getProjectsByUser(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("userId")

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

	response, err := s.core.ProjectService.ListByUser(ctx, userID, query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getProject(c echo.Context) error {
	projectID := c.Param("projectId")
	project, err := s.core.ProjectService.Get(c.Request().Context(), projectID)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	if project == nil {
		return s.core.HandleError(nil, http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, project)
}

func (s *Server) updateProject(c echo.Context) error {
	projectID := c.Param("projectId")
	var project models.Project
	if err := c.Bind(&project); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}
	project.ID = projectID

	if err := c.Validate(&project); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := s.core.ProjectService.Update(c.Request().Context(), &project); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) projectAddUser(c echo.Context) error {
	projectID := c.Param("projectId")

	s.core.Logger.Info("hello")

	var projectUser models.ProjectUser
	if err := c.Bind(&projectUser); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}
	projectUser.ProjectID = projectID

	if err := c.Validate(&projectUser); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := s.core.ProjectService.AddUser(c.Request().Context(), &projectUser); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) deleteProject(c echo.Context) error {
	projectID := c.Param("projectId")
	if err := s.core.ProjectService.Delete(c.Request().Context(), projectID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) projectRemoveUser(c echo.Context) error {
	projectID := c.Param("projectId")
	userID := c.Param("userId")
	if err := s.core.ProjectService.RemoveUser(c.Request().Context(), projectID, userID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}
