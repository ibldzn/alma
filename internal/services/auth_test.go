package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/types"
	"golang.org/x/crypto/bcrypt"
)

type fakeAuthRepository struct {
	user              models.User
	err               error
	updateErr         error
	lastLoginUserID   uint64
	lastLoginAt       time.Time
	findByUsernameArg string
}

func (f *fakeAuthRepository) FindByUsername(ctx context.Context, username string) (models.User, error) {
	f.findByUsernameArg = username
	if f.err != nil {
		return models.User{}, f.err
	}
	return f.user, nil
}

func (f *fakeAuthRepository) UpdateLastLogin(ctx context.Context, userID uint64, at time.Time) error {
	f.lastLoginUserID = userID
	f.lastLoginAt = at
	return f.updateErr
}

func TestAuthServiceAuthenticateValidCredentials(t *testing.T) {
	hash := bcryptHash(t, "secret")
	now := time.Date(2026, time.June, 7, 10, 0, 0, 0, time.UTC)
	repo := &fakeAuthRepository{
		user: models.User{
			ID:       7,
			Name:     "Haytsam",
			Username: "haytsam",
			Password: hash,
		},
	}
	service := NewAuthService(repo)
	service.Now = func() time.Time { return now }

	user, err := service.Authenticate(context.Background(), " haytsam ", "secret")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}

	if repo.findByUsernameArg != "haytsam" {
		t.Fatalf("username arg = %q, want haytsam", repo.findByUsernameArg)
	}
	if repo.lastLoginUserID != 7 || !repo.lastLoginAt.Equal(now) {
		t.Fatalf("last login update = id %d at %s, want id 7 at %s", repo.lastLoginUserID, repo.lastLoginAt, now)
	}
	if user.Password != "" {
		t.Fatalf("user.Password = %q, want empty", user.Password)
	}
	if user.LastLogin == nil || !user.LastLogin.Equal(now) {
		t.Fatalf("user.LastLogin = %v, want %s", user.LastLogin, now)
	}
}

func TestAuthServiceAuthenticateRejectsWrongPassword(t *testing.T) {
	repo := &fakeAuthRepository{
		user: models.User{
			ID:       7,
			Username: "haytsam",
			Password: bcryptHash(t, "secret"),
		},
	}
	service := NewAuthService(repo)

	_, err := service.Authenticate(context.Background(), "haytsam", "wrong")
	if !errors.Is(err, types.ErrInvalidCredentials) {
		t.Fatalf("Authenticate error = %v, want ErrInvalidCredentials", err)
	}
	if repo.lastLoginUserID != 0 {
		t.Fatalf("last login updated for invalid password")
	}
}

func TestAuthServiceAuthenticateRejectsMissingUser(t *testing.T) {
	repo := &fakeAuthRepository{err: sql.ErrNoRows}
	service := NewAuthService(repo)

	_, err := service.Authenticate(context.Background(), "missing", "secret")
	if !errors.Is(err, types.ErrInvalidCredentials) {
		t.Fatalf("Authenticate error = %v, want ErrInvalidCredentials", err)
	}
}

func bcryptHash(t *testing.T, password string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword returned error: %v", err)
	}
	return string(hash)
}
