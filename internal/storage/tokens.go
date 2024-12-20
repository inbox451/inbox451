package storage

import (
	"context"

	"inbox451/internal/models"

	_ "github.com/lib/pq"
)

func (r *repository) ListTokensByUser(ctx context.Context, user_id int, limit, offset int) ([]*models.Token, int, error) {
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

func (r *repository) GetTokenByUser(ctx context.Context, token_id int, user_id int) (*models.Token, error) {
	var token models.Token
	err := r.queries.GetTokenByUser.GetContext(ctx, &token, token_id, user_id)
	return &token, handleDBError(err)
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

func (r *repository) DeleteToken(ctx context.Context, tokenID int) error {
	result, err := r.queries.DeleteToken.ExecContext(ctx, tokenID)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}
