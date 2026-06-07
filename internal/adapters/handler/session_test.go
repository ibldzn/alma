package handler

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ibldzn/alma/internal/models"
)

func TestSessionManagerRoundTrip(t *testing.T) {
	now := time.Date(2026, time.June, 7, 10, 0, 0, 0, time.UTC)
	manager := testSessionManager(t, now)
	cookie, err := manager.NewCookie(models.User{
		ID:       9,
		Name:     "Haytsam",
		Username: "haytsam",
	})
	if err != nil {
		t.Fatalf("NewCookie returned error: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)
	user, err := manager.UserFromRequest(req)
	if err != nil {
		t.Fatalf("UserFromRequest returned error: %v", err)
	}

	if user.ID != 9 || user.Name != "Haytsam" || user.Username != "haytsam" {
		t.Fatalf("session user = %+v", user)
	}
	if !cookie.HttpOnly || cookie.SameSite == 0 || cookie.Path != "/" {
		t.Fatalf("cookie security attrs = %+v", cookie)
	}
}

func TestSessionManagerRejectsTamperedCookie(t *testing.T) {
	manager := testSessionManager(t, time.Date(2026, time.June, 7, 10, 0, 0, 0, time.UTC))
	cookie, err := manager.NewCookie(models.User{
		ID:       9,
		Username: "haytsam",
	})
	if err != nil {
		t.Fatalf("NewCookie returned error: %v", err)
	}
	cookie.Value += "tampered"

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)
	_, err = manager.UserFromRequest(req)
	if !errors.Is(err, errSessionInvalid) {
		t.Fatalf("UserFromRequest error = %v, want errSessionInvalid", err)
	}
}

func TestSessionManagerRejectsExpiredCookie(t *testing.T) {
	now := time.Date(2026, time.June, 7, 10, 0, 0, 0, time.UTC)
	manager := testSessionManager(t, now)
	cookie, err := manager.NewCookie(models.User{
		ID:       9,
		Username: "haytsam",
	})
	if err != nil {
		t.Fatalf("NewCookie returned error: %v", err)
	}

	manager.now = func() time.Time { return now.Add(2 * time.Hour) }
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)
	_, err = manager.UserFromRequest(req)
	if !errors.Is(err, errSessionExpired) {
		t.Fatalf("UserFromRequest error = %v, want errSessionExpired", err)
	}
}

func testSessionManager(t *testing.T, now time.Time) *SessionManager {
	t.Helper()

	manager, err := NewSessionManager(SessionConfig{
		Secret: []byte("test-session-secret"),
		TTL:    time.Hour,
	})
	if err != nil {
		t.Fatalf("NewSessionManager returned error: %v", err)
	}
	manager.now = func() time.Time { return now }
	return manager
}
