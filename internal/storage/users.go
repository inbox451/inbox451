package storage

import (
	"context"
	"database/sql"
	"errors"

	"inbox451/internal/models"

	_ "github.com/lib/pq"
)

func (r *repository) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	var total int
	err := r.queries.CountUsers.GetContext(ctx, &total)
	if err != nil {
		return nil, 0, err
	}

	var users []*models.User
	err = r.queries.ListUsers.SelectContext(ctx, &users, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *repository) GetUser(ctx context.Context, userID int) (*models.User, error) {
	var user models.User
	err := r.queries.GetUser.GetContext(ctx, &user, userID)
	return &user, handleDBError(err)
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.queries.GetUserByUsername.GetContext(ctx, &user, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	// Assuming you add a query named 'get-user-by-email' in queries.sql
	err := r.queries.GetUserByEmail.GetContext(ctx, &user, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil, nil when not found
		}
		return nil, handleDBError(err)
	}
	return &user, nil
}

func (r *repository) CreateUser(ctx context.Context, user *models.User) error {
	return r.queries.CreateUser.QueryRowContext(ctx,
		user.Name,
		user.Username,
		user.Password,
		user.Email,
		user.Status,
		user.Role,
		user.PasswordLogin).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *repository) UpdateUser(ctx context.Context, user *models.User) error {
	// Only include password if it's not empty after potential hashing
	// If the password field in the incoming user model is empty, we keep the existing one in DB.
	// The CASE statement in the SQL query handles this.
	passwordToUpdate := ""
	if user.Password.Valid {
		passwordToUpdate = user.Password.String
	}

	return r.queries.UpdateUser.QueryRowContext(ctx,
		user.Name,
		user.Username,
		passwordToUpdate,
		user.Email,
		user.Status,
		user.Role,
		user.PasswordLogin,
		user.ID).
		Scan(&user.UpdatedAt)
}

func (r *repository) DeleteUser(ctx context.Context, id int) error {
	result, err := r.queries.DeleteUser.ExecContext(ctx, id)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}
