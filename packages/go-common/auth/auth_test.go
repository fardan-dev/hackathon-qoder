package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeVerifier struct{}

func (fakeVerifier) Verify(_ context.Context, token string) (Principal, error) {
	return Principal{UserID: "u-" + token, TenantID: "018f0000-0000-7000-8000-000000000001",
		Email: token + "@demo.local", Role: token}, nil
}

func TestSuppressionApprovalRoleGate(t *testing.T) {
	handler := Middleware(fakeVerifier{})(RequireRoles(RoleSupervisor)(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })))

	reviewer := httptest.NewRequest(http.MethodPost, "/v1/suppressions", nil)
	reviewer.Header.Set("Authorization", "Bearer reviewer")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reviewer)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("reviewer status = %d, want 403", rec.Code)
	}

	supervisor := httptest.NewRequest(http.MethodPost, "/v1/suppressions", nil)
	supervisor.Header.Set("Authorization", "Bearer supervisor")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, supervisor)
	if rec.Code != http.StatusOK {
		t.Fatalf("supervisor status = %d, want 200", rec.Code)
	}
}

func TestUnauthenticatedIsRejected(t *testing.T) {
	handler := Middleware(fakeVerifier{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/me", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
