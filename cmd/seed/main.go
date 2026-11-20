package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var namespace = uuid.MustParse("2f6a7c8d-9e10-4b3c-8d5e-6f7a8b9c0d1e")

type seedUser struct {
	Email    string
	Name     string
	Role     string
	Password string
}

type coverage struct {
	Key      string
	Label    string
	Covered  bool
	Rules    []string
	Residual string
	Sort     int
}

var coverageRows = []coverage{
	{"invoice_reuse", "Invoice reuse / double billing", true, []string{"A1", "A3", "A5", "A6"}, "", 1},
	{"duplicate_claim", "Duplicate claim submission", true, []string{"A2", "C3", "C5", "F2"}, "", 2},
	{"amount_inflation", "Amount inflation vs invoice", true, []string{"A4", "B1", "B2", "B3", "B4"}, "", 3},
	{"structuring", "Structuring below approval threshold", true, []string{"B3", "B4"}, "", 4},
	{"cover_breach", "Claim outside cover / limit breach", true, []string{"B5", "B6", "C1", "C2", "E1", "E2"}, "", 5},
	{"new_policy_reinstatement", "New policy or reinstatement claim", true, []string{"C1", "C2", "E1"}, "", 6},
	{"repeat_claimant", "Repeat claimant within window", true, []string{"C4", "F1", "F3", "G3"}, "", 7},
	{"provider_overbilling", "Provider overbilling / volume anomaly", true, []string{"D4", "D5"}, "", 8},
	{"collusion_network", "Collusion network (shared identity/provider)", true, []string{"D1", "D2", "D3", "D6", "D7", "F4"}, "", 9},
	{"repeat_offender", "Repeat offender with prior confirmed flag", true, []string{"G1", "G2", "G3"}, "", 10},
	{"document_amount_edit", "Document amount edit (OCR mismatch)", false, []string{"H1", "H6"}, "Document forensics analyzer is gated off in Phase 0/1; relies on structured amount mismatch only.", 11},
	{"document_forgery", "Document metadata forgery", false, []string{"H2", "H3", "H4", "H5", "H7"}, "No document forensics analyzer in the thin slice; residual risk accepted until Phase 2.", 12},
	{"document_rescan", "Document re-scan / near-duplicate", false, []string{"H5"}, "Perceptual hashing is gated off; only exact doc_sha256 reuse is covered by A3.", 13},
	{"high_quality_forgery", "High-quality forgery (ELA/template)", false, []string{"H8", "H9", "H10"}, "ELA and template-deviation analysis are gated off; residual risk accepted.", 14},
}

func main() {
	for _, path := range []string{".env", "../.env"} {
		loadEnvFile(path)
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://claims:claims@localhost:5432/claims?sslmode=disable"
	}
	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		tenantID = config.DefaultTenantID
	}
	if _, err := uuid.Parse(tenantID); err != nil {
		log.Fatalf("TENANT_ID must be a uuid: %v", err)
	}
	users := []seedUser{
		{Email: "admin@claimshield.local", Name: "Admin Demo", Role: "admin", Password: requireEnv("SEED_ADMIN_PASSWORD")},
		{Email: "reviewer@claimshield.local", Name: "Reviewer Demo", Role: "reviewer", Password: requireEnv("SEED_REVIEWER_PASSWORD")},
		{Email: "supervisor@claimshield.local", Name: "Supervisor Demo", Role: "supervisor", Password: requireEnv("SEED_SUPERVISOR_PASSWORD")},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("hash %s: %v", u.Email, err)
		}
		id := uuid.NewSHA1(namespace, []byte(tenantID+"|"+u.Email)).String()
		if _, err := pool.Exec(ctx, `INSERT INTO identity.users
			(id, tenant_id, email, display_name, password_hash, role)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (tenant_id, email) DO UPDATE SET
				display_name=EXCLUDED.display_name, password_hash=EXCLUDED.password_hash, role=EXCLUDED.role`,
			id, tenantID, u.Email, u.Name, string(hash), u.Role); err != nil {
			log.Fatalf("seed user %s: %v", u.Email, err)
		}
		fmt.Printf("user %-28s role=%s\n", u.Email, u.Role)
	}
	for _, c := range coverageRows {
		if _, err := pool.Exec(ctx, `INSERT INTO reporting.coverage_typology
			(typology_key, label, in_scope, phase01_covered, rule_ids, residual_risk, sort_order)
			VALUES ($1,$2,true,$3,$4,$5,$6)
			ON CONFLICT (typology_key) DO UPDATE SET
				label=EXCLUDED.label, in_scope=EXCLUDED.in_scope,
				phase01_covered=EXCLUDED.phase01_covered, rule_ids=EXCLUDED.rule_ids,
				residual_risk=EXCLUDED.residual_risk, sort_order=EXCLUDED.sort_order`,
			c.Key, c.Label, c.Covered, c.Rules, nullIfEmpty(c.Residual), c.Sort); err != nil {
			log.Fatalf("seed coverage %s: %v", c.Key, err)
		}
	}
	fmt.Printf("coverage matrix: %d typologies (%d covered)\n", len(coverageRows), coveredCount())
}

func coveredCount() int {
	n := 0
	for _, c := range coverageRows {
		if c.Covered {
			n++
		}
	}
	return n
}

func requireEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("%s is required (see .env.example); demo credentials must not be hardcoded", key)
	}
	return v
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func loadEnvFile(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}
