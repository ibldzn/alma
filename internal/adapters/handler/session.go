package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ibldzn/alma/internal/models"
)

const sessionCookieName = "alma_session"

var (
	errSessionInvalid = errors.New("invalid session")
	errSessionExpired = errors.New("expired session")
)

type SessionConfig struct {
	Secret []byte
	TTL    time.Duration
	Secure bool
}

type SessionManager struct {
	secret []byte
	ttl    time.Duration
	secure bool
	now    func() time.Time
}

type SessionUser struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type sessionPayload struct {
	User      SessionUser `json:"user"`
	ExpiresAt int64       `json:"exp"`
}

type sessionContextKey struct{}

func NewSessionManager(config SessionConfig) (*SessionManager, error) {
	if len(config.Secret) == 0 {
		return nil, errors.New("session secret is required")
	}

	ttl := config.TTL
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}

	return &SessionManager{
		secret: append([]byte(nil), config.Secret...),
		ttl:    ttl,
		secure: config.Secure,
		now:    time.Now,
	}, nil
}

func (m *SessionManager) NewCookie(user models.User) (*http.Cookie, error) {
	value, err := m.sign(sessionPayload{
		User: SessionUser{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
		},
		ExpiresAt: m.now().Add(m.ttl).Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(m.ttl.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.secure,
	}, nil
}

func (m *SessionManager) ClearCookie() *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   m.secure,
	}
}

func (m *SessionManager) UserFromRequest(r *http.Request) (SessionUser, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return SessionUser{}, errSessionInvalid
	}

	payload, err := m.verify(cookie.Value)
	if err != nil {
		return SessionUser{}, err
	}

	return payload.User, nil
}

func (m *SessionManager) sign(payload sessionPayload) (string, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signature := m.signature(payloadPart)
	signaturePart := base64.RawURLEncoding.EncodeToString(signature)
	return payloadPart + "." + signaturePart, nil
}

func (m *SessionManager) verify(value string) (sessionPayload, error) {
	payloadPart, signaturePart, ok := strings.Cut(value, ".")
	if !ok || payloadPart == "" || signaturePart == "" {
		return sessionPayload{}, errSessionInvalid
	}

	wantSignature := m.signature(payloadPart)
	gotSignature, err := base64.RawURLEncoding.DecodeString(signaturePart)
	if err != nil {
		return sessionPayload{}, errSessionInvalid
	}
	if !hmac.Equal(gotSignature, wantSignature) {
		return sessionPayload{}, errSessionInvalid
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return sessionPayload{}, errSessionInvalid
	}

	var payload sessionPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return sessionPayload{}, errSessionInvalid
	}
	if payload.User.ID == 0 || payload.User.Username == "" {
		return sessionPayload{}, errSessionInvalid
	}
	if payload.ExpiresAt <= m.now().Unix() {
		return sessionPayload{}, errSessionExpired
	}

	return payload, nil
}

func (m *SessionManager) signature(payloadPart string) []byte {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payloadPart))
	return mac.Sum(nil)
}

func (u SessionUser) DisplayName() string {
	if strings.TrimSpace(u.Name) != "" {
		return u.Name
	}
	return u.Username
}

func withSessionUser(ctx context.Context, user SessionUser) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, user)
}

func sessionUserFromContext(ctx context.Context) (SessionUser, bool) {
	user, ok := ctx.Value(sessionContextKey{}).(SessionUser)
	if !ok || user.ID == 0 {
		return SessionUser{}, false
	}
	return user, true
}
