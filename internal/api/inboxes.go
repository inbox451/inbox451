package api

import (
	"errors"
	"inbox451/internal/storage"
	"net/http"
	"strconv"

	"inbox451/internal/models"

	"github.com/labstack/echo/v4"
)

func (s *Server) createInbox(c echo.Context) error {
	projectID, _ := strconv.Atoi(c.Param("projectId"))
	var inbox models.Inbox
	if err := c.Bind(&inbox); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}
	inbox.ProjectID = projectID

	err := c.Validate(inbox)
	if err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	newInbox, err := s.core.InboxService.Create(inbox)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, newInbox)
}

func (s *Server) getInboxes(c echo.Context) error {
	projectID, _ := strconv.Atoi(c.Param("projectId"))

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

	response, err := s.core.InboxService.ListByProject(projectID, query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getInbox(c echo.Context) error {
	inboxID, _ := strconv.Atoi(c.Param("inboxId"))
	inbox, err := s.core.InboxService.Get(inboxID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return s.core.HandleError(nil, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, inbox)
}

func (s *Server) updateInbox(c echo.Context) error {
	inboxID, _ := strconv.Atoi(c.Param("inboxId"))
	projectID, _ := strconv.Atoi(c.Param("projectId"))

	var inbox models.Inbox
	if err := c.Bind(&inbox); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	// Set both ID and AccountID before validation
	inbox.ID = inboxID
	inbox.ProjectID = projectID

	if err := c.Validate(&inbox); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	newInbox, err := s.core.InboxService.Update(inbox)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, newInbox)
}

func (s *Server) deleteInbox(c echo.Context) error {
	inboxID, _ := strconv.Atoi(c.Param("inboxId"))
	if err := s.core.InboxService.Delete(inboxID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}
