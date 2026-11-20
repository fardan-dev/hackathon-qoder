package rulesengine

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse date %q: %v", s, err)
	}
	return d
}

func datePtr(t *testing.T, s string) *time.Time {
	t.Helper()
	d := mustDate(t, s)
	return &d
}

func testPolicy(t *testing.T, id, inception, expiry string, sum int64) *PolicySnapshot {
	t.Helper()
	return &PolicySnapshot{
		PolicyID: id, Status: "ACTIVE",
		Inception: mustDate(t, inception), Expiry: mustDate(t, expiry),
		SumInsured: sum,
	}
}

func testInvoice(t *testing.T, number, date, provider string, amount int64) *InvoiceSnapshot {
	t.Helper()
	return &InvoiceSnapshot{Number: number, Date: mustDate(t, date), ProviderID: provider, Amount: amount}
}

func testContext(t *testing.T) EvalContext {
	t.Helper()
	histConfirmed := ClaimSnapshot{
		TenantID: "tenant-a", ClaimID: "CLM-1001", BatchID: "B-HIST",
		PolicyID: "POL-1", ClaimantID: "CUST-1", ProviderID: "PRV-1", InvoiceNumber: "INV-H1",
		LossType: "MOTOR", Status: "SETTLED",
		ClaimDate: mustDate(t, "2026-02-01"), IncidentDate: mustDate(t, "2026-01-01"),
		ClaimedAmount: 10_000_000, Currency: "IDR",
		Invoice:        testInvoice(t, "INV-H1", "2026-01-02", "PRV-1", 10_000_000),
		Policy:         testPolicy(t, "POL-1", "2025-01-01", "2027-01-01", 100_000_000),
		PriorConfirmed: true,
	}
	histCleared := ClaimSnapshot{
		TenantID: "tenant-a", ClaimID: "CLM-1002", BatchID: "B-HIST",
		PolicyID: "POL-7", ClaimantID: "CUST-2", ProviderID: "PRV-2", InvoiceNumber: "INV-H2",
		LossType: "MOTOR", Status: "CLEARED",
		ClaimDate: mustDate(t, "2026-10-20"), IncidentDate: mustDate(t, "2026-10-15"),
		ClaimedAmount: 20_000_000, Currency: "IDR",
		Invoice:            testInvoice(t, "INV-H2", "2026-10-16", "PRV-2", 20_000_000),
		Policy:             testPolicy(t, "POL-7", "2025-01-01", "2027-01-01", 100_000_000),
		PriorClearedReason: "FP_NON_INJURY",
	}

	dupA := ClaimSnapshot{
		TenantID: "tenant-a", ClaimID: "CLM-2001", BatchID: "B-1",
		PolicyID: "POL-2", ClaimantID: "CUST-2", ProviderID: "PRV-2", InvoiceNumber: "INV-900",
		LossType: "MOTOR", Status: "SUBMITTED",
		ClaimDate: mustDate(t, "2026-11-01"), IncidentDate: mustDate(t, "2026-10-15"),
		ClaimedAmount: 20_000_000, Currency: "IDR",
		Invoice:    testInvoice(t, "INV-900", "2026-10-16", "PRV-2", 20_000_000),
		Policy:     testPolicy(t, "POL-2", "2026-10-01", "2027-10-01", 50_000_000),
		Identities: map[string]string{"BANK_ACCOUNT": "hash-bank-shared", "PHONE": "hash-phone-1"},
		DocSHA256:  "doc-shared",
	}
	dupB := ClaimSnapshot{
		TenantID: "tenant-a", ClaimID: "CLM-2002", BatchID: "B-1",
		PolicyID: "POL-3", ClaimantID: "CUST-3", ProviderID: "PRV-2", InvoiceNumber: "INV-900",
		LossType: "MOTOR", Status: "SUBMITTED",
		ClaimDate: mustDate(t, "2026-11-01"), IncidentDate: mustDate(t, "2026-10-20"),
		ClaimedAmount: 6_000_000, Currency: "IDR",
		Invoice:    testInvoice(t, "INV-900", "2026-10-21", "PRV-2", 7_000_000),
		Policy:     testPolicy(t, "POL-3", "2025-01-01", "2027-01-01", 50_000_000),
		Identities: map[string]string{"BANK_ACCOUNT": "hash-bank-shared"},
		DocSHA256:  "doc-shared",
	}
	high := ClaimSnapshot{
		TenantID: "tenant-a", ClaimID: "CLM-2003", BatchID: "B-1",
		PolicyID: "POL-4", ClaimantID: "CUST-4", ProviderID: "PRV-3", InvoiceNumber: "INV-901",
		LossType: "PROPERTY", Status: "SUBMITTED",
		ClaimDate: mustDate(t, "2026-11-05"), IncidentDate: mustDate(t, "2026-10-01"),
		ClaimedAmount: 120_000_000, Currency: "IDR",
		Invoice: testInvoice(t, "INV-901", "2026-10-02", "PRV-3", 120_000_000),
		Policy:  testPolicy(t, "POL-4", "2024-01-01", "2028-01-01", 100_000_000),
	}
	clean := ClaimSnapshot{
		TenantID: "tenant-a", ClaimID: "CLM-2004", BatchID: "B-1",
		PolicyID: "POL-5", ClaimantID: "CUST-5", ProviderID: "PRV-4", InvoiceNumber: "INV-902",
		LossType: "MOTOR", Status: "SUBMITTED",
		ClaimDate: mustDate(t, "2026-11-06"), IncidentDate: mustDate(t, "2026-10-25"),
		ClaimedAmount: 3_000_000, Currency: "IDR",
		Invoice: testInvoice(t, "INV-902", "2026-10-26", "PRV-4", 3_000_000),
		Policy:  testPolicy(t, "POL-5", "2025-01-01", "2027-01-01", 50_000_000),
	}
	repeat := ClaimSnapshot{
		TenantID: "tenant-a", ClaimID: "CLM-2005", BatchID: "B-1",
		PolicyID: "POL-6", ClaimantID: "CUST-1", ProviderID: "PRV-5", InvoiceNumber: "INV-903",
		LossType: "MOTOR", Status: "SUBMITTED",
		ClaimDate: mustDate(t, "2026-11-07"), IncidentDate: mustDate(t, "2026-09-01"),
		ClaimedAmount: 8_000_000, Currency: "IDR",
		Invoice: testInvoice(t, "INV-903", "2026-09-02", "PRV-5", 8_000_000),
		Policy:  testPolicy(t, "POL-6", "2025-01-01", "2027-01-01", 50_000_000),
	}

	return EvalContext{
		TenantID: "tenant-a",
		AsOf:     mustDate(t, "2026-11-20"),
		History:  []ClaimSnapshot{histConfirmed, histCleared},
		Current:  []ClaimSnapshot{dupA, dupB, high, clean, repeat},
	}
}

func flagsByRule(cr ClaimResult) map[string]Flag {
	out := make(map[string]Flag, len(cr.Flags))
	for _, f := range cr.Flags {
		out[f.RuleID] = f
	}
	return out
}

func resultByClaim(t *testing.T, br BatchResult, claimID string) ClaimResult {
	t.Helper()
	for _, r := range br.Results {
		if r.ClaimID == claimID {
			return r
		}
	}
	t.Fatalf("claim %s missing from results", claimID)
	return ClaimResult{}
}

func TestCatalogShape(t *testing.T) {
	cat := Catalog()
	if len(cat) != 43 {
		t.Fatalf("catalog size = %d, want 43", len(cat))
	}
	var active, gated int
	seen := map[string]bool{}
	for _, r := range cat {
		if seen[r.ID] {
			t.Fatalf("duplicate rule id %s", r.ID)
		}
		seen[r.ID] = true
		if r.Version != 1 || r.ParamsVersion != 1 {
			t.Errorf("rule %s version/params version = %d/%d", r.ID, r.Version, r.ParamsVersion)
		}
		if len(r.Typologies) == 0 {
			t.Errorf("rule %s has no typology mapping", r.ID)
		}
		switch r.Status {
		case StatusActive:
			active++
		case StatusGatedOff:
			gated++
		default:
			t.Errorf("rule %s has unknown status %q", r.ID, r.Status)
		}
	}
	if active != 33 || gated != 10 {
		t.Fatalf("active/gated = %d/%d, want 33/10", active, gated)
	}
	if len(CatalogByFamily("H")) != 10 {
		t.Fatalf("family H size != 10")
	}
}

func TestEngineFlagsAndFourElements(t *testing.T) {
	br, err := Evaluate(testContext(t), CatalogVersion)
	if err != nil {
		t.Fatal(err)
	}
	if len(br.Results) != 5 {
		t.Fatalf("results = %d, want 5", len(br.Results))
	}

	dupA := flagsByRule(resultByClaim(t, br, "CLM-2001"))
	for _, id := range []string{"A1", "A2", "A3", "A6", "B4", "C1", "C4", "D1", "D6", "G2"} {
		if _, ok := dupA[id]; !ok {
			t.Errorf("CLM-2001 missing rule %s; got %v", id, keys(dupA))
		}
	}
	dupB := flagsByRule(resultByClaim(t, br, "CLM-2002"))
	for _, id := range []string{"A1", "A3", "A4", "A6", "B4", "D1"} {
		if _, ok := dupB[id]; !ok {
			t.Errorf("CLM-2002 missing rule %s; got %v", id, keys(dupB))
		}
	}
	high := flagsByRule(resultByClaim(t, br, "CLM-2003"))
	for _, id := range []string{"B1", "B5", "C3"} {
		if _, ok := high[id]; !ok {
			t.Errorf("CLM-2003 missing rule %s; got %v", id, keys(high))
		}
	}
	if clean := resultByClaim(t, br, "CLM-2004"); len(clean.Flags) != 0 {
		t.Errorf("CLM-2004 expected no flags, got %v", keys(flagsByRule(clean)))
	}
	repeat := flagsByRule(resultByClaim(t, br, "CLM-2005"))
	for _, id := range []string{"C3", "G1"} {
		if _, ok := repeat[id]; !ok {
			t.Errorf("CLM-2005 missing rule %s; got %v", id, keys(repeat))
		}
	}

	for _, cr := range br.Results {
		var gated int
		for _, ne := range cr.NotEvaluated {
			if ne.Reason == "GATED_OFF" {
				gated++
				if !strings.HasPrefix(ne.RuleID, "H") {
					t.Errorf("claim %s non-H rule %s marked GATED_OFF", cr.ClaimID, ne.RuleID)
				}
			}
		}
		if gated != 10 {
			t.Errorf("claim %s gated rules = %d, want 10", cr.ClaimID, gated)
		}
		for _, f := range cr.Flags {
			if f.RuleID == "" || f.RuleVersion == 0 || f.ParamsVersion == 0 {
				t.Errorf("flag missing rule identity: %+v", f)
			}
			if f.TriggerValue == "" {
				t.Errorf("flag %s missing trigger value", f.RuleID)
			}
			if len(f.EvidenceRefs) == 0 {
				t.Errorf("flag %s missing evidence refs", f.RuleID)
			}
			if strings.TrimSpace(f.Explanation) == "" {
				t.Errorf("flag %s missing explanation", f.RuleID)
			}
			if f.FlagID == "" || !strings.HasPrefix(f.FlagID, "sha256:") {
				t.Errorf("flag %s has invalid flag id %q", f.RuleID, f.FlagID)
			}
		}
	}
}

func keys(m map[string]Flag) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestEngineDeterministic(t *testing.T) {
	first, err := Evaluate(testContext(t), CatalogVersion)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Evaluate(testContext(t), CatalogVersion)
	if err != nil {
		t.Fatal(err)
	}
	if first.OutputHash != second.OutputHash {
		t.Fatalf("identical input produced different hashes: %s vs %s", first.OutputHash, second.OutputHash)
	}

	reordered := testContext(t)
	reverseClaims(reordered.Current)
	reverseClaims(reordered.History)
	reordered.KillSwitches = []KillSwitch{
		{Scope: "rule", Target: "Z9", Enabled: true},
		{Scope: "family", Target: "B", Enabled: false},
	}
	third, err := Evaluate(reordered, CatalogVersion)
	if err != nil {
		t.Fatal(err)
	}
	firstWithSameSwitches := testContext(t)
	firstWithSameSwitches.KillSwitches = []KillSwitch{
		{Scope: "family", Target: "B", Enabled: false},
		{Scope: "rule", Target: "Z9", Enabled: true},
	}
	fourth, err := Evaluate(firstWithSameSwitches, CatalogVersion)
	if err != nil {
		t.Fatal(err)
	}
	if third.OutputHash != fourth.OutputHash {
		t.Fatalf("input ordering changed output hash: %s vs %s", third.OutputHash, fourth.OutputHash)
	}
}

func reverseClaims(in []ClaimSnapshot) {
	for i, j := 0, len(in)-1; i < j; i, j = i+1, j-1 {
		in[i], in[j] = in[j], in[i]
	}
}

func TestGatedFamilyOffAndOn(t *testing.T) {
	ec := testContext(t)
	ec.DocForensicsEnabled = true
	br, err := Evaluate(ec, CatalogVersion)
	if err != nil {
		t.Fatal(err)
	}
	for _, cr := range br.Results {
		for _, f := range cr.Flags {
			if f.Family == "H" {
				t.Fatalf("claim %s produced H flag %s with analyzer unavailable", cr.ClaimID, f.RuleID)
			}
		}
		var hNotEvaluated int
		for _, ne := range cr.NotEvaluated {
			if strings.HasPrefix(ne.RuleID, "H") {
				hNotEvaluated++
				if ne.Reason != "DATA_UNAVAILABLE" {
					t.Errorf("H rule %s reason = %s, want DATA_UNAVAILABLE", ne.RuleID, ne.Reason)
				}
			}
		}
		if hNotEvaluated != 10 {
			t.Errorf("claim %s H not_evaluated = %d, want 10", cr.ClaimID, hNotEvaluated)
		}
	}
}

func TestKillSwitchStopsFamily(t *testing.T) {
	ec := testContext(t)
	ec.KillSwitches = []KillSwitch{{Scope: "family", Target: "A", Enabled: true}}
	br, err := Evaluate(ec, CatalogVersion)
	if err != nil {
		t.Fatal(err)
	}
	for _, cr := range br.Results {
		for _, f := range cr.Flags {
			if f.Family == "A" {
				t.Fatalf("claim %s still has A flag %s after family kill switch", cr.ClaimID, f.RuleID)
			}
		}
		var killed int
		for _, ne := range cr.NotEvaluated {
			if ne.Reason == "KILL_SWITCH" {
				killed++
			}
		}
		if killed != 6 {
			t.Errorf("claim %s killed rules = %d, want 6", cr.ClaimID, killed)
		}
	}
}

func TestRulePackVersionPinned(t *testing.T) {
	if _, err := Evaluate(testContext(t), "2026.10.0"); err == nil {
		t.Fatal("expected error for unknown rule pack version")
	}
	ec := testContext(t)
	ec.AsOf = time.Time{}
	if _, err := Evaluate(ec, CatalogVersion); err == nil {
		t.Fatal("expected error for missing as_of")
	}
}

func TestLabelIsolationBoundary(t *testing.T) {
	types := []reflect.Type{
		reflect.TypeOf(ClaimSnapshot{}),
		reflect.TypeOf(InvoiceSnapshot{}),
		reflect.TypeOf(PolicySnapshot{}),
		reflect.TypeOf(EvalContext{}),
	}
	blocked := []string{"label", "ground_truth", "groundtruth", "typology", "is_fraud"}
	for _, typ := range types {
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			tag := strings.ToLower(f.Tag.Get("json"))
			name := strings.ToLower(f.Name)
			for _, word := range blocked {
				if strings.Contains(name, word) || strings.Contains(tag, word) {
					t.Errorf("%s.%s exposes ground-truth-like field (name=%s tag=%s)", typ.Name(), f.Name, name, tag)
				}
			}
		}
	}
}
