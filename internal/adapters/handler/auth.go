package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/ibldzn/alma/internal/types"
)

const invalidLoginMessage = "Username atau password salah."

type LoginPageData struct {
	Next     string
	Username string
	Error    string
}

func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	next := safeRedirectPath(r.URL.Query().Get("next"))
	if _, err := h.SessionManager.UserFromRequest(r); err == nil {
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}

	h.renderLogin(w, http.StatusOK, LoginPageData{Next: next})
}

func (h *Handler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderLogin(w, http.StatusBadRequest, LoginPageData{
			Next:  "/",
			Error: "Unable to read login request.",
		})
		return
	}

	next := safeRedirectPath(r.Form.Get("next"))
	username := strings.TrimSpace(r.Form.Get("username"))
	password := r.Form.Get("password")

	user, err := h.AuthService.Authenticate(r.Context(), username, password)
	if errors.Is(err, types.ErrInvalidCredentials) {
		h.renderLogin(w, http.StatusUnauthorized, LoginPageData{
			Next:     next,
			Username: username,
			Error:    invalidLoginMessage,
		})
		return
	}
	if err != nil {
		h.renderLogin(w, http.StatusInternalServerError, LoginPageData{
			Next:     next,
			Username: username,
			Error:    "Unable to sign in. Please try again.",
		})
		return
	}

	cookie, err := h.SessionManager.NewCookie(user)
	if err != nil {
		h.renderLogin(w, http.StatusInternalServerError, LoginPageData{
			Next:     next,
			Username: username,
			Error:    "Unable to create session.",
		})
		return
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, h.SessionManager.ClearCookie())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := h.SessionManager.UserFromRequest(r)
		if err != nil {
			http.Redirect(w, r, "/login?next="+url.QueryEscape(r.URL.RequestURI()), http.StatusFound)
			return
		}

		next.ServeHTTP(w, r.WithContext(withSessionUser(r.Context(), user)))
	})
}

func (h *Handler) renderLogin(w http.ResponseWriter, status int, data LoginPageData) {
	h.renderTemplate(w, status, "login.html", data)
}

func safeRedirectPath(next string) string {
	next = strings.TrimSpace(next)
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		return "/"
	}

	parsed, err := url.Parse(next)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return "/"
	}

	return next
}
