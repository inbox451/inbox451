package api

import (
	"errors"
	"inbox451/internal/storage"
	"net/http"
	"strconv"

	"inbox451/internal/models"

	"github.com/labstack/echo/v4"
)

func (s *Server) createUser(c echo.Context) error {
	var input models.User
	if err := c.Bind(&input); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := c.Validate(&input); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	newUser, err := s.core.UserService.Create(input)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, newUser)
}

func (s *Server) getUsers(c echo.Context) error {
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

	response, err := s.core.UserService.List(query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getUser(c echo.Context) error {
	userID, _ := strconv.Atoi(c.Param("userId"))
	user, err := s.core.UserService.Get(userID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return s.core.HandleError(nil, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, user)
}

func (s *Server) updateUser(c echo.Context) error {
	userID, _ := strconv.Atoi(c.Param("userId"))
	var user models.User
	if err := c.Bind(&user); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}
	user.ID = userID

	if err := c.Validate(&user); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	updatedUser, err := s.core.UserService.Update(user)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return s.core.HandleError(nil, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, updatedUser)
}

func (s *Server) deleteUser(c echo.Context) error {
	userID, _ := strconv.Atoi(c.Param("userId"))
	if err := s.core.UserService.Delete(userID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}
