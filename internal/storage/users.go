package storage

import (
	"database/sql"
	"errors"

	"inbox451/internal/models"

	_ "github.com/lib/pq"
)

func (r *repository) ListUsers(limit int, offset int) ([]models.User, int, error) {
	var total int
	var users []models.User

	err := r.queries.CountUsers.Get(&total)
	if err != nil {
		return nil, 0, err
	}

	err = r.queries.ListUsers.Select(&users, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *repository) GetUser(userId int) (models.User, error) {
	var user models.User
	err := r.queries.GetUser.Get(&user, userId)
	return user, handleDBError(err)
}

func (r *repository) GetUserByUsername(username string) (models.User, error) {
	var user models.User
	err := r.queries.GetUserByUsername.Get(&user, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, nil
		}
		return user, err
	}
	return user, nil
}

func (r *repository) CreateUser(user models.User) (models.User, error) {
	var userId int
	err := r.queries.CreateUser.QueryRow(
		user.Name,
		user.Username,
		user.Password,
		user.Email,
		user.Status,
		user.Role,
		user.PasswordLogin).
		Scan(&userId)
	if err != nil {
		return models.User{}, handleDBError(err)
	}
	return r.GetUser(userId)
}

func (r *repository) UpdateUser(user models.User) (models.User, error) {
	res, err := r.queries.UpdateUser.Exec(
		user.Name,
		user.Username,
		user.Password,
		user.Email,
		user.Status,
		user.Role,
		user.PasswordLogin,
		user.ID)
	if err != nil {
		return models.User{}, handleDBError(err)
	}

	if err := handleRowsAffected(res); err != nil {
		return models.User{}, err
	}

	return r.GetUser(user.ID)
}

func (r *repository) DeleteUser(id int) error {
	result, err := r.queries.DeleteUser.Exec(id)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}
