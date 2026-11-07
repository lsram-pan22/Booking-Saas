package repositories

import (
	"context"
	"errors"
	"strings"

	"booking-saas/controllers/super_admin/auth/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) FindByUsername(
	ctx context.Context,
	username string,
) (*models.SuperAdminAccount, error) {

	username = strings.TrimSpace(username)

	var user models.SuperAdminAccount

	query := `
        SELECT
            id,
			username,
            email,
            password_hash,
            full_name,
            system_role,
            status
        FROM users
        WHERE LOWER(username) = LOWER($1)
        LIMIT 1
    `

	err := r.DB.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.SystemRole,
		&user.Status,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}
