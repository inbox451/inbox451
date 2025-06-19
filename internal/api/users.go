package api

import (
	"inbox451/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) createUser(c echo.Context) error {
	ctx := c.Request().Context()

	var input models.User
	if err := c.Bind(&input); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := c.Validate(&input); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := s.core.UserService.Create(ctx, &input); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, input)
}

func (s *Server) getUsers(c echo.Context) error {
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

	response, err := s.core.UserService.List(ctx, query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getUser(c echo.Context) error {
	userID := c.Param("userId")
	user, err := s.core.UserService.Get(c.Request().Context(), userID)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	if user == nil {
		return s.core.HandleError(nil, http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, user)
}

func (s *Server) updateUser(c echo.Context) error {
	userID := c.Param("userId")
	var user models.User
	if err := c.Bind(&user); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}
	user.ID = userID

	if err := c.Validate(&user); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := s.core.UserService.Update(c.Request().Context(), &user); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) deleteUser(c echo.Context) error {
	userID := c.Param("userId")
	if err := s.core.UserService.Delete(c.Request().Context(), userID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}
