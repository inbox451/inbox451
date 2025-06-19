package api

import (
	"inbox451/internal/models"
	"inbox451/internal/storage"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) getMessages(c echo.Context) error {
	inboxID := c.Param("inboxId")

	var query models.MessageQuery
	if err := c.Bind(&query); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if query.Limit == 0 {
		query.Limit = 10
	}

	if err := c.Validate(&query); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	response, err := s.core.MessageService.ListByInbox(c.Request().Context(), inboxID, query.Limit, query.Offset, query.IsRead)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getMessage(c echo.Context) error {
	messageID := c.Param("messageId")

	message, err := s.core.MessageService.Get(c.Request().Context(), messageID)
	if err != nil {
		if err == storage.ErrNotFound {
			return s.core.HandleError(err, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	if message == nil {
		return s.core.HandleError(storage.ErrNotFound, http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, message)
}

func (s *Server) markMessageRead(c echo.Context) error {
	messageID := c.Param("messageId")

	err := s.core.MessageService.MarkAsRead(c.Request().Context(), messageID)
	if err != nil {
		if err == storage.ErrNotFound {
			return s.core.HandleError(err, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (s *Server) markMessageUnread(c echo.Context) error {
	messageID := c.Param("messageId")

	err := s.core.MessageService.MarkAsUnread(c.Request().Context(), messageID)
	if err != nil {
		if err == storage.ErrNotFound {
			return s.core.HandleError(err, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (s *Server) deleteMessage(c echo.Context) error {
	messageID := c.Param("messageId")

	err := s.core.MessageService.Delete(c.Request().Context(), messageID)
	if err != nil {
		if err == storage.ErrNotFound {
			return s.core.HandleError(err, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}
