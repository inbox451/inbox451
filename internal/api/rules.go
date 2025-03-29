package api

import (
	"errors"
	"inbox451/internal/storage"
	"net/http"
	"strconv"

	"inbox451/internal/models"

	"github.com/labstack/echo/v4"
)

func (s *Server) createRule(c echo.Context) error {
	inboxID, _ := strconv.Atoi(c.Param("inboxId"))
	var rule models.ForwardRule
	if err := c.Bind(&rule); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}
	rule.InboxID = inboxID

	if err := c.Validate(&rule); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	newRule, err := s.core.RuleService.Create(rule)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, newRule)
}

func (s *Server) getRules(c echo.Context) error {
	inboxID, _ := strconv.Atoi(c.Param("inboxId"))

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

	response, err := s.core.RuleService.ListByInbox(inboxID, query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getRule(c echo.Context) error {
	ruleID, _ := strconv.Atoi(c.Param("ruleId"))
	rule, err := s.core.RuleService.Get(ruleID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return s.core.HandleError(nil, http.StatusNotFound)
		}
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, rule)
}

func (s *Server) updateRule(c echo.Context) error {
	ruleID, _ := strconv.Atoi(c.Param("ruleId"))
	inboxID, _ := strconv.Atoi(c.Param("inboxId"))

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

	updatedRule, err := s.core.RuleService.Update(rule)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, updatedRule)
}

func (s *Server) deleteRule(c echo.Context) error {
	ruleID, _ := strconv.Atoi(c.Param("ruleId"))
	if err := s.core.RuleService.Delete(ruleID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}
