package api

import (
	"inbox451/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v9"
)

// GET /users/:userId/tokens/
func (s *Server) ListTokensByUser(c echo.Context) error {
	ctx := c.Request().Context()
	userId := c.Param("userId")

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

	response, err := s.core.TokenService.ListByUser(ctx, userId, query.Limit, query.Offset)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

// GET /users/:userId/tokens/:tokenId
func (s *Server) GetTokenByUser(c echo.Context) error {
	userID := c.Param("userId")
	tokenID := c.Param("tokenId")

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

	response, err := s.core.TokenService.GetByUser(c.Request().Context(), tokenID, userID)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, response)
}

// POST /users/:userId/tokens/
func (s *Server) CreateTokenForUser(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user ID from URL
	userID := c.Param("userId")

	// Token Input DTO
	type TokenInput struct {
		Name      string    `json:"name" validate:"required"`
		ExpiresAt null.Time `json:"expires_at"`
	}

	input := new(TokenInput)
	if err := c.Bind(input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(input); err != nil {
		return s.core.HandleError(err, http.StatusBadRequest)
	}

	token := &models.Token{
		UserID:    userID,
		Name:      input.Name,
		ExpiresAt: input.ExpiresAt,
	}

	newToken, err := s.core.TokenService.CreateForUser(ctx, userID, token)
	if err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, newToken)
}

func (s *Server) DeleteTokenByUser(c echo.Context) error {
	userID := c.Param("userId")
	tokenID := c.Param("tokenId")
	if err := s.core.TokenService.DeleteByUser(c.Request().Context(), userID, tokenID); err != nil {
		return s.core.HandleError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}
