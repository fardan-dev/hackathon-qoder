package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/httpx"
)

const (
	RoleAdmin      = "admin"
	RoleReviewer   = "reviewer"
	RoleSupervisor = "supervisor"

	SessionCookie = "claimshield_session"
)

type Principal struct {
	UserID      string `json:"user_id"`
	TenantID    string `json:"tenant_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type Verifier interface {
	Verify(ctx context.Context, token string) (Principal, error)
}

// HTTPVerifier asks the identity service to validate an opaque session token.
// No service other than identity reads the sessions table.
type HTTPVerifier struct {
	baseURL       string
	internalToken string
	client        *http.Client
}

func NewHTTPVerifier(baseURL, internalToken string) *HTTPVerifier {
	return &HTTPVerifier{
		baseURL:       strings.TrimRight(baseURL, "/"),
		internalToken: internalToken,
		client:        &http.Client{Timeout: 2 * time.Second},
	}
}

func (v *HTTPVerifier) Verify(ctx context.Context, token string) (Principal, error) {
	body, _ := json.Marshal(map[string]string{"token": token})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.baseURL+"/internal/auth/verify", bytes.NewReader(body))
	if err != nil {
		return Principal{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", v.internalToken)
	resp, err := v.client.Do(req)
	if err != nil {
		return Principal{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Principal{}, fmt.Errorf("identity rejected session: status %d", resp.StatusCode)
	}
	var p Principal
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return Principal{}, err
	}
	if p.UserID == "" || p.TenantID == "" || p.Role == "" {
		return Principal{}, fmt.Errorf("identity returned incomplete principal")
	}
	return p, nil
}

type ctxKey string

const principalKey ctxKey = "principal"

func Middleware(v Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := tokenFrom(r)
			if token == "" {
				httpx.WriteProblem(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "session required")
				return
			}
			p, err := v.Verify(r.Context(), token)
			if err != nil {
				httpx.WriteProblem(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "session invalid or expired")
				return
			}
			ctx := context.WithValue(r.Context(), principalKey, p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := From(r.Context())
			if !ok {
				httpx.WriteProblem(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "session required")
				return
			}
			if _, ok := allowed[p.Role]; !ok {
				httpx.WriteProblem(w, r, http.StatusForbidden, "FORBIDDEN", "role is not allowed to perform this action")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func From(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

// TokenFrom returns the bearer token or session cookie value.
func TokenFrom(r *http.Request) string {
	return tokenFrom(r)
}

func tokenFrom(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	if c, err := r.Cookie(SessionCookie); err == nil {
		return c.Value
	}
	return ""
}
