package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/app"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/auth"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/httpx"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/pii"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/server"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const sessionTTL = 12 * time.Hour

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userRow struct {
	ID           string
	Email        string
	DisplayName  string
	PasswordHash string
	Role         string
}

type limiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func newLimiter() *limiter { return &limiter{attempts: map[string][]time.Time{}} }

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-time.Minute)
	recent := l.attempts[key][:0]
	for _, t := range l.attempts[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= 5 {
		l.attempts[key] = recent
		return false
	}
	l.attempts[key] = append(recent, now)
	return true
}

func (l *limiter) reset(key string) {
	l.mu.Lock()
	delete(l.attempts, key)
	l.mu.Unlock()
}

func main() {
	ctx, cancel := server.SignalContext()
	defer cancel()
	a, err := app.New(ctx, "identity")
	if err != nil {
		slog.Error("startup_failed", "error", err.Error())
		return
	}
	defer a.Close()

	lim := newLimiter()
	verifier := a.Verifier()
	a.Router.Route("/v1", func(r chi.Router) {
		r.Post("/auth/login", login(a, lim))
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(verifier))
			r.Get("/me", me)
			r.Post("/auth/logout", logout(a))
		})
	})
	a.Router.With(httpx.InternalOnly(a.Cfg.InternalToken)).Post("/internal/auth/verify", verify(a))
	if err := a.Run(ctx); err != nil {
		a.Log.Error("server_failed", "error", err.Error())
	}
}

func login(a *app.App, lim *limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := httpx.DecodeJSON(w, r, &req); err != nil {
			httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_JSON", "invalid login payload")
			return
		}
		email := strings.ToLower(strings.TrimSpace(req.Email))
		if email == "" || req.Password == "" {
			httpx.WriteProblem(w, r, http.StatusBadRequest, "REQUIRED_FIELD", "email and password are required")
			return
		}
		key := clientIP(r) + "|" + email
		if !lim.allow(key) {
			httpx.WriteProblem(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "too many login attempts")
			return
		}

		var u userRow
		err := a.Pool.QueryRow(r.Context(), `SELECT id::text, email, display_name, password_hash, role
			FROM identity.users WHERE tenant_id=$1 AND email=$2`, a.Cfg.TenantID, email).
			Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Role)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httpx.WriteProblem(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is incorrect")
				return
			}
			a.Log.Error("login_lookup_failed", "error", err.Error())
			httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "login failed")
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
			httpx.WriteProblem(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is incorrect")
			return
		}
		lim.reset(key)

		token, err := newToken()
		if err != nil {
			httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not create session")
			return
		}
		expires := time.Now().UTC().Add(sessionTTL)
		_, err = a.Pool.Exec(r.Context(), `INSERT INTO identity.sessions (token_hash, user_id, tenant_id, expires_at)
			VALUES ($1, $2, $3, $4)`, pii.TokenHash(token), u.ID, a.Cfg.TenantID, expires)
		if err != nil {
			a.Log.Error("session_insert_failed", "error", err.Error())
			httpx.WriteProblem(w, r, http.StatusInternalServerError, "INTERNAL", "could not create session")
			return
		}
		_, _ = a.Pool.Exec(r.Context(), `DELETE FROM identity.sessions WHERE expires_at < now()`)
		http.SetCookie(w, &http.Cookie{
			Name: auth.SessionCookie, Value: token, Path: "/", HttpOnly: true,
			SameSite: http.SameSiteLaxMode, Secure: a.Cfg.CookieSecure, MaxAge: int(sessionTTL.Seconds()),
		})
		a.Log.Info("login_succeeded", "user_id", u.ID, "role", u.Role)
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"token": token,
			"user": auth.Principal{
				UserID: u.ID, TenantID: a.Cfg.TenantID, Email: u.Email, DisplayName: u.DisplayName, Role: u.Role,
			},
			"expires_at": expires,
		})
	}
}

func me(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.From(r.Context())
	httpx.WriteJSON(w, http.StatusOK, p)
}

func logout(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := auth.TokenFrom(r)
		if token != "" {
			_, _ = a.Pool.Exec(r.Context(), `DELETE FROM identity.sessions WHERE token_hash=$1`, pii.TokenHash(token))
		}
		http.SetCookie(w, &http.Cookie{
			Name: auth.SessionCookie, Value: "", Path: "/", HttpOnly: true,
			SameSite: http.SameSiteLaxMode, Secure: a.Cfg.CookieSecure, MaxAge: -1,
		})
		w.WriteHeader(http.StatusNoContent)
	}
}

func verify(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Token string `json:"token"`
		}
		if err := httpx.DecodeJSON(w, r, &req); err != nil || req.Token == "" {
			httpx.WriteProblem(w, r, http.StatusBadRequest, "INVALID_REQUEST", "token is required")
			return
		}
		var p auth.Principal
		err := a.Pool.QueryRow(r.Context(), `SELECT u.id::text, u.tenant_id::text, u.email, u.display_name, u.role
			FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id
			WHERE s.token_hash=$1 AND s.expires_at > now()`, pii.TokenHash(req.Token)).
			Scan(&p.UserID, &p.TenantID, &p.Email, &p.DisplayName, &p.Role)
		if err != nil {
			httpx.WriteProblem(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "session invalid or expired")
			return
		}
		_, _ = a.Pool.Exec(r.Context(), `UPDATE identity.sessions SET last_seen_at=now() WHERE token_hash=$1`, pii.TokenHash(req.Token))
		httpx.WriteJSON(w, http.StatusOK, p)
	}
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func clientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); v != "" {
		return strings.Split(v, ",")[0]
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
