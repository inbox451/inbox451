package core

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"inbox451/internal/models"
)

// TokenService handles operations related to API tokens
type TokenService struct {
	core *Core
}

// NewTokensService creates a new TokenService instance
func NewTokensService(core *Core) TokenService {
	return TokenService{core: core}
}

// ListByUser retrieves a paginated list of tokens for a given user
//
// Parameters:
//   - ctx: Context for the request
//   - userId: ID of the user to list tokens for
//   - limit: Maximum number of tokens to return
//   - offset: Number of tokens to skip
//
// Returns:
//   - *models.PaginatedResponse containing the tokens and pagination info
//   - error if the operation fails
func (s *TokenService) ListByUser(ctx context.Context, userId int, limit, offset int) (*models.PaginatedResponse, error) {
	s.core.Logger.Info("Listing tokens for userId %d with limit: %d and offset: %d", userId, limit, offset)

	tokens, total, err := s.core.Repository.ListTokensByUser(ctx, userId, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list tokens for userId %d: %v", userId, err)
		return nil, err
	}

	response := &models.PaginatedResponse{
		Data: tokens,
		Pagination: models.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	s.core.Logger.Info("Successfully retrieved %d tokens for userID %d (total: %d)", len(tokens), userId, total)
	return response, nil
}

// GetByUser retrieves a specific token for a user
//
// Parameters:
//   - ctx: Context for the request
//   - tokenID: ID of the token to retrieve
//   - userID: ID of the user who owns the token
//
// Returns:
//   - *models.Token if found
//   - ErrNotFound if token doesn't exist
//   - error if the operation fails
func (s *TokenService) GetByUser(ctx context.Context, tokenID int, userID int) (*models.Token, error) {
	s.core.Logger.Debug("Fetching token with ID: %d for userID: %d ", tokenID, userID)

	token, err := s.core.Repository.GetTokenByUser(ctx, tokenID, userID)
	if err != nil {
		s.core.Logger.Error("Failed to fetch token: %v", err)
		return nil, err
	}

	if token == nil {
		s.core.Logger.Info("Token not found with ID: %d for userID %d", tokenID, userID)
		return nil, ErrNotFound
	}

	return token, nil
}

// CreateForUser creates a new API token for a user
//
// Parameters:
//   - ctx: Context for the request
//   - userID: ID of the user to create token for
//   - tokenData: Optional token configuration including name and expiration
//
// Returns:
//   - *models.Token containing the newly created token
//   - error if the operation fails
func (s *TokenService) CreateForUser(ctx context.Context, userID int, tokenData *models.Token) (*models.Token, error) {
	s.core.Logger.Debug("Creating token for userId: %d", userID)

	newToken := models.Token{}
	newToken.UserID = userID

	// Use the provided name
	if tokenData != nil && tokenData.Name != "" {
		newToken.Name = tokenData.Name
	} else {
		newToken.Name = "API Token"
	}

	tokenStr, err := GenerateSecureTokenBase64()
	if err != nil {
		s.core.Logger.Error("Failed to generate secure token: %v", err)
		return nil, err
	}
	newToken.Token = tokenStr

	// Set expires_at if provided
	if tokenData != nil {
		newToken.ExpiresAt = tokenData.ExpiresAt
	}

	err = s.core.Repository.CreateToken(ctx, &newToken)
	if err != nil {
		return nil, err
	}

	return &newToken, nil
}

// DeleteByUser deletes a specific token belonging to a user
//
// Parameters:
//   - ctx: Context for the request
//   - userID: ID of the user who owns the token
//   - tokenID: ID of the token to delete
//
// Returns:
//   - error if the operation fails or token doesn't exist
func (s *TokenService) DeleteByUser(ctx context.Context, userID int, tokenID int) error {
	s.core.Logger.Debug("Deleting token with ID: %d for userID %d", tokenID, userID)

	// Check if token exists for this user
	_, err := s.GetByUser(ctx, userID, tokenID)
	if err != nil {
		return err
	}

	if err := s.core.Repository.DeleteToken(ctx, tokenID); err != nil {
		s.core.Logger.Error("Failed to delete token: %v", err)
		return err
	}

	s.core.Logger.Debug("Successfully deleted token with ID: %d for userId %d", tokenID, userID)
	return nil
}

func (s *TokenService) GetByValue(ctx context.Context, tokenValue string) (*models.Token, error) {
	tokenMask := tokenValue[0:3] + "..." + tokenValue[len(tokenValue)-3:]
	s.core.Logger.Debug("Fetching token by value: %s", tokenMask)

	token, err := s.core.Repository.GetTokenByValue(ctx, tokenValue)
	if err != nil {
		s.core.Logger.Error("Failed to fetch token by value: %v", err)
		return nil, err
	}

	if token == nil {
		s.core.Logger.Info("Token not found with value: %s", tokenMask)
		return nil, ErrNotFound
	}

	return token, nil
}

// GenerateSecureTokenBase64 generates a cryptographically secure random token
// encoded in URL-safe base64. It returns a string of approximately 43 characters
// (for 32 bytes of entropy) that is safe for use in URLs and file names.
//
// The generated token uses the following format:
//   - 32 bytes of random data from crypto/rand
//   - URL-safe base64 encoding
//   - No padding characters
//
// Example output: "xJ_dwq8k-rLp5xGhq2d4mNvKzHjY3bWl1fTnM9iR0oE"
//
// If the random number generator fails, it returns an empty string and an error.
func GenerateSecureTokenBase64() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
