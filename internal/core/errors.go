package core

import (
	"errors"
	"fmt"
	"net/http"

	"inbox451/internal/storage"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type APIError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

type ValidationError struct {
	Field  string `json:"field"`
	Rule   string `json:"rule"`
	Value  string `json:"value"`
	Reason string `json:"reason"`
}

var (
	ErrNotFound = &APIError{
		Code:    http.StatusNotFound,
		Message: "resource not found",
	}

	ErrBadRequest = &APIError{
		Code:    http.StatusBadRequest,
		Message: "bad request",
	}

	ErrAuthFailed = errors.New("invalid credentials")

	ErrAccountInactive = errors.New("user account is inactive")

	ErrPasswordLoginDisabled = errors.New("password login is disabled for this user")
)

func (c *Core) HandleError(err error, code int) error {
	if err == nil {
		// If err is nil but we're handling a not found case
		if code == http.StatusNotFound {
			return echo.NewHTTPError(http.StatusNotFound, ErrNotFound)
		}
		// For other nil error cases, just return nil
		return nil
	}

	// Handle specific storage errors
	switch {
	case errors.Is(err, storage.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, ErrNotFound)
	case errors.Is(err, storage.ErrNoRowsAffected):
		return echo.NewHTTPError(http.StatusNotFound, ErrNotFound)
	}

	if code >= 500 {
		c.Logger.ErrorWithStack(err)
	} else {
		c.Logger.Error("HTTP %d: %v", code, err)
	}

	// Handle validation errors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		details := make([]ValidationError, 0)
		for _, err := range validationErrors {
			details = append(details, ValidationError{
				Field:  err.Field(),
				Rule:   err.Tag(),
				Value:  fmt.Sprintf("%v", err.Value()),
				Reason: err.Error(),
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, APIError{
			Code:    http.StatusBadRequest,
			Message: "Validation failed",
			Details: details,
		})
	}

	// Handle API errors
	if apiErr, ok := err.(*APIError); ok {
		return echo.NewHTTPError(apiErr.Code, apiErr)
	}

	// Handle generic errors
	return echo.NewHTTPError(code, APIError{
		Code:    code,
		Message: err.Error(),
	})
}
