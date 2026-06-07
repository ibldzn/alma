package interfaces

import (
	"context"
	"time"

	"github.com/ibldzn/alma/internal/models"
)

type IAuthRepository interface {
	FindByUsername(ctx context.Context, username string) (models.User, error)
	UpdateLastLogin(ctx context.Context, userID uint64, at time.Time) error
}

type IAuthService interface {
	Authenticate(ctx context.Context, username, password string) (models.User, error)
}
