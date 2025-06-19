package storage

import (
	"context"

	"inbox451/internal/models"

	_ "github.com/lib/pq"
)

func (r *repository) ListTokensByUser(ctx context.Context, user_id string, limit, offset int) ([]*models.Token, int, error) {
	var total int
	err := r.queries.CountTokensByUser.GetContext(ctx, &total, user_id)
	if err != nil {
		return nil, 0, err
	}

	tokens := []*models.Token{}
	if total > 0 {
		err = r.queries.ListTokensByUser.SelectContext(ctx, &tokens, user_id, limit, offset)
		if err != nil {
			return nil, 0, err
		}
	}

	return tokens, total, nil
}

func (r *repository) GetTokenByUser(ctx context.Context, token_id string, user_id string) (*models.Token, error) {
	var token models.Token
	err := r.queries.GetTokenByUser.GetContext(ctx, &token, token_id, user_id)
	return &token, handleDBError(err)
}

// GetTokenByValue finds a token by its value (the actual token string)
func (r *repository) GetTokenByValue(ctx context.Context, tokenValue string) (*models.Token, error) {
	var token models.Token
	err := r.queries.GetTokenByValue.GetContext(ctx, &token, tokenValue)
	return &token, handleDBError(err)
}

// UpdateTokenLastUsed updates the last_used_at timestamp for a token.
func (r *repository) UpdateTokenLastUsed(ctx context.Context, tokenID string) error {
	_, err := r.queries.UpdateTokenLastUsed.ExecContext(ctx, tokenID)
	return handleDBError(err) // Doesn't need handleRowsAffected, it's okay if it doesn't update
}

// PruneExpiredTokens deletes tokens that have passed their expiration date.
func (r *repository) PruneExpiredTokens(ctx context.Context) (int64, error) {
	result, err := r.queries.PruneExpiredTokens.ExecContext(ctx)
	if err != nil {
		return 0, handleDBError(err)
	}
	return result.RowsAffected()
}

func (r *repository) CreateToken(ctx context.Context, token *models.Token) error {
	err := r.queries.CreateToken.QueryRowContext(
		ctx,
		token.UserID,
		token.Token,
		token.Name,
		token.ExpiresAt,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.Name,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.UpdatedAt,
	)
	return handleDBError(err)
}

func (r *repository) DeleteToken(ctx context.Context, tokenID string) error {
	result, err := r.queries.DeleteToken.ExecContext(ctx, tokenID)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}
