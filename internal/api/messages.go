package api

import (
	"errors"
	"net/http"
	"strconv"

	"inbox451/internal/models"
	"inbox451/internal/storage"

	"github.com/labstack/echo/v4"
)

func (s *Server) getMessages(c echo.Context) error {
	inboxID, _ := strconv.Atoi(c.Param("inboxId"))

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

	response, err := s.core.MessageService.ListByInbox(inboxID, query.Limit, query.Offset, query.IsRead)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getMessage(c echo.Context) error {
	messageID, _ := strconv.Atoi(c.Param("messageId"))

	message, err := s.core.MessageService.Get(messageID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {

			return s.core.HandleError(err, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, message)
}

func (s *Server) markMessageRead(c echo.Context) error {
	messageID, _ := strconv.Atoi(c.Param("messageId"))

	_, err := s.core.MessageService.MarkAsRead(messageID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return s.core.HandleError(err, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (s *Server) markMessageUnread(c echo.Context) error {
	messageID, _ := strconv.Atoi(c.Param("messageId"))

	_, err := s.core.MessageService.MarkAsUnread(messageID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return s.core.HandleError(err, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (s *Server) deleteMessage(c echo.Context) error {
	messageID, _ := strconv.Atoi(c.Param("messageId"))

	err := s.core.MessageService.Delete(messageID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return s.core.HandleError(err, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}
