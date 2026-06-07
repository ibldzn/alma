package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/ibldzn/alma/internal/interfaces"
	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/types"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	AuthRepo interfaces.IAuthRepository
	Now      func() time.Time
}

func NewAuthService(repo interfaces.IAuthRepository) *AuthService {
	return &AuthService{
		AuthRepo: repo,
		Now:      time.Now,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, username, password string) (models.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return models.User{}, types.ErrInvalidCredentials
	}

	user, err := s.AuthRepo.FindByUsername(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, types.ErrInvalidCredentials
	}
	if err != nil {
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return models.User{}, types.ErrInvalidCredentials
	}

	now := s.now()
	if err := s.AuthRepo.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return models.User{}, err
	}

	user.Password = ""
	user.LastLogin = &now
	return user, nil
}

func (s *AuthService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}
