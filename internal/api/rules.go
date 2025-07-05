package api

import (
	"net/http"

	"inbox451/internal/models"

	"github.com/labstack/echo/v4"
)

func (s *Server) createRule(c echo.Context) error {
	inboxID := c.Param("inboxId")
	var rule models.ForwardRule
	if err := c.Bind(&rule); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}
	rule.InboxID = inboxID

	if err := c.Validate(&rule); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := s.core.RuleService.Create(c.Request().Context(), &rule); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, rule)
}

func (s *Server) getRules(c echo.Context) error {
	inboxID := c.Param("inboxId")

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

	response, err := s.core.RuleService.ListByInbox(c.Request().Context(), inboxID, query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getRule(c echo.Context) error {
	ruleID := c.Param("ruleId")
	rule, err := s.core.RuleService.Get(c.Request().Context(), ruleID)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	if rule == nil {
		return s.core.HandleError(nil, http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, rule)
}

func (s *Server) updateRule(c echo.Context) error {
	ruleID := c.Param("ruleId")
	inboxID := c.Param("inboxId")

	var rule models.ForwardRule
	if err := c.Bind(&rule); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	// Set both ID and InboxID before validation
	rule.ID = ruleID
	rule.InboxID = inboxID

	if err := c.Validate(&rule); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	if err := s.core.RuleService.Update(c.Request().Context(), &rule); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) deleteRule(c echo.Context) error {
	ruleID := c.Param("ruleId")
	if err := s.core.RuleService.Delete(c.Request().Context(), ruleID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}
