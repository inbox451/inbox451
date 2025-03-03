package storage

import (
	"inbox451/internal/models"

	_ "github.com/lib/pq"
)

func (r *repository) ListTokensByUser(userId int, limit, offset int) ([]models.Token, int, error) {
	var total int
	var tokens []models.Token
	err := r.queries.CountTokensByUser.Get(&total, userId)
	if err != nil {
		return nil, 0, err
	}

	err = r.queries.ListTokensByUser.Select(&tokens, userId, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return tokens, total, nil
}

func (r *repository) GetTokenByUser(tokenId int, userId int) (models.Token, error) {
	var token models.Token
	err := r.queries.GetTokenByUser.Get(&token, tokenId, userId)
	return token, handleDBError(err)
}

func (r *repository) CreateToken(token models.Token) (models.Token, error) {
	var tokenId int
	err := r.queries.CreateToken.QueryRow(
		token.UserID,
		token.Token,
		token.Name,
		token.ExpiresAt,
	).Scan(&tokenId)
	if err != nil {
		return models.Token{}, handleDBError(err)
	}
	return r.GetTokenByUser(tokenId, token.UserID)
}

func (r *repository) DeleteToken(tokenId int) error {
	result, err := r.queries.DeleteToken.Exec(tokenId)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}
