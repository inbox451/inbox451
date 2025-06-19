package api

import (
	"inbox451/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) createInbox(c echo.Context) error {
	projectID := c.Param("projectId")
	var inbox models.Inbox
	if err := c.Bind(&inbox); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}
	inbox.ProjectID = projectID

	if err := c.Validate(&inbox); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := s.core.InboxService.Create(c.Request().Context(), &inbox); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, inbox)
}

func (s *Server) getInboxes(c echo.Context) error {
	accountID := c.Param("projectId")

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

	response, err := s.core.InboxService.ListByProject(c.Request().Context(), accountID, query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getInbox(c echo.Context) error {
	inboxID := c.Param("inboxId")
	inbox, err := s.core.InboxService.Get(c.Request().Context(), inboxID)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	if inbox == nil {
		return s.core.HandleError(nil, http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, inbox)
}

func (s *Server) updateInbox(c echo.Context) error {
	inboxID := c.Param("inboxId")
	projectID := c.Param("projectId")

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

	if err := s.core.InboxService.Update(c.Request().Context(), &inbox); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) deleteInbox(c echo.Context) error {
	inboxID := c.Param("inboxId")
	if err := s.core.InboxService.Delete(c.Request().Context(), inboxID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}
