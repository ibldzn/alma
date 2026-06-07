package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/ibldzn/alma/internal/constants"
	"github.com/ibldzn/alma/internal/models"
	"github.com/jmoiron/sqlx"
)

type AuthRepository struct {
	AppDB *sqlx.DB
}

func NewAuthRepository(appDB *sqlx.DB) *AuthRepository {
	return &AuthRepository{
		AppDB: appDB,
	}
}

func (r *AuthRepository) FindByUsername(ctx context.Context, username string) (models.User, error) {
	query := fmt.Sprintf(`
		SELECT id, name, username, password, last_login, created_at
		FROM %s
		WHERE username = ?
		LIMIT 1
	`, constants.UsersTable)

	var user models.User
	if err := r.AppDB.GetContext(ctx, &user, query, username); err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *AuthRepository) UpdateLastLogin(ctx context.Context, userID uint64, at time.Time) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET last_login = ?
		WHERE id = ?
	`, constants.UsersTable)

	_, err := r.AppDB.ExecContext(ctx, query, at, userID)
	return err
}
