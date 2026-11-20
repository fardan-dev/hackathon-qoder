package main

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/contracts"
	"github.com/fardanwardhana/qoder-hackathon/packages/rulesengine"
)

const (
	gateTenant = "018f0000-0000-7000-8000-000000000001"
	gateBatch  = "B-DEMO-CONTROL-GATE"
	gateAsOf   = "2026-11-01T00:00:00Z"
)

var gateInputFiles = []string{"claims.csv", "invoices.csv", "policies.csv", "providers.csv", "identities.csv"}

// Rule priority follows BRD §7 / FR-09..FR-11. H9 and H10 are the only
// experimental rules with an explicit FP budget; H1..H8 are gated off.
var rulePriority = map[string]string{
	"A1": "Must", "A2": "Must", "B1": "Must", "B5": "Must", "C4": "Must", "E2": "Must",
	"A3": "Should", "A4": "Should", "A6": "Should", "B2": "Should", "B3": "Should",
	"B6": "Should", "C1": "Should", "C2": "Should", "C5": "Should", "D1": "Should",
	"D2": "Should", "D7": "Should", "F1": "Should", "G1": "Should", "G2": "Should",
	"A5": "Could", "B4": "Could", "C3": "Could", "D3": "Could", "D4": "Could",
	"D5": "Could", "D6": "Could", "E1": "Could", "F2": "Could", "F3": "Could",
	"F4": "Could", "G3": "Could",
	"H1": "GatedOff", "H2": "GatedOff", "H3": "GatedOff", "H4": "GatedOff",
	"H5": "GatedOff", "H6": "GatedOff", "H7": "GatedOff", "H8": "GatedOff",
	"H9": "H9-H10", "H10": "H9-H10",
}

type truthRow struct {
	label    int
	typology string
	expected []string
}

type ruleGate struct {
	RuleID           string `json:"rule_id"`
	Priority         string `json:"priority"`
	ControlFlags     int    `json:"control_flags"`
	ControlClaims    int    `json:"control_claims"`
	FalsePositiveBPS int    `json:"false_positive_bps"`
	ThresholdBPS     int    `json:"threshold_bps"`
	Passed           bool   `json:"passed"`
}

type gateReport struct {
	Version         string     `json:"version"`
	DatasetSHA256   string     `json:"dataset_sha256"`
	GroundTruthHash string     `json:"ground_truth_sha256"`
	RulePackVersion string     `json:"rule_pack_version"`
	EngineVersion   string     `json:"engine_version"`
	AsOf            string     `json:"as_of"`
	OutputHash      string     `json:"output_hash"`
	Claims          int        `json:"claims"`
	PositiveClaims  int        `json:"positive_claims"`
	PositiveFlagged int        `json:"positive_claims_flagged"`
	CaseRecallBPS   int        `json:"case_recall_bps"`
	ControlClaims   int        `json:"control_claims"`
	GatedOffRules   int        `json:"gated_off_rules"`
	Rules           []ruleGate `json:"rules"`
	Violations      []string   `json:"violations"`
	Passed          bool       `json:"passed"`
}

func TestControlClaimFalsePositiveGate(t *testing.T) {
	dir := writeDataset(t)
	assertNoGroundTruthColumns(t, dir)
	batch := loadGateBatch(t, dir)
	truth := readGroundTruth(t, dir)

	br, err := rulesengine.Evaluate(rulesengine.EvalContext{
		TenantID:            gateTenant,
		AsOf:                mustParseTime(t, gateAsOf),
		Current:             buildSnapshots(t, batch),
		History:             nil,
		DocForensicsEnabled: true,
	}, rulesengine.CatalogVersion)
	if err != nil {
		t.Fatal(err)
	}

	controls := map[string]bool{}
	positives := map[string]bool{}
	for claimID, row := range truth {
		if row.label == 0 {
			controls[claimID] = true
		} else {
			positives[claimID] = true
		}
	}
	if len(controls) == 0 || len(positives) == 0 {
		t.Fatal("dataset needs both control and planted claims")
	}

	flagsPerRule := map[string]int{}
	flaggedPositives := 0
	for _, res := range br.Results {
		if positives[res.ClaimID] && hasActiveFlag(res.Flags) {
			flaggedPositives++
		}
		if !controls[res.ClaimID] {
			continue
		}
		seen := map[string]bool{}
		for _, f := range res.Flags {
			if seen[f.RuleID] {
				continue
			}
			seen[f.RuleID] = true
			flagsPerRule[f.RuleID]++
		}
	}

	catalog := rulesengine.Catalog()
	report := gateReport{
		Version:         "FP-GATE-v1",
		DatasetSHA256:   hashFiles(t, dir, gateInputFiles),
		GroundTruthHash: hashFiles(t, dir, []string{"ground_truth.csv"}),
		RulePackVersion: br.RulePackVersion,
		EngineVersion:   br.EngineVersion,
		AsOf:            br.AsOf.Format(time.RFC3339),
		OutputHash:      br.OutputHash,
		Claims:          len(batch.Claims),
		PositiveClaims:  len(positives),
		PositiveFlagged: flaggedPositives,
		CaseRecallBPS:   flaggedPositives * 10000 / len(positives),
		ControlClaims:   len(controls),
		Rules:           []ruleGate{},
		Violations:      []string{},
	}
	if len(catalog) != len(rulePriority) {
		t.Fatalf("priority map has %d rules, catalog has %d", len(rulePriority), len(catalog))
	}
	for _, meta := range catalog {
		class, ok := rulePriority[meta.ID]
		if !ok {
			t.Fatalf("rule %s has no FP gate priority", meta.ID)
		}
		if class == "GatedOff" {
			report.GatedOffRules++
			continue
		}
		flags := flagsPerRule[meta.ID]
		rate := flags * 10000 / len(controls)
		threshold := thresholdBPS(class)
		passed := passesGate(class, flags, len(controls))
		report.Rules = append(report.Rules, ruleGate{
			RuleID: meta.ID, Priority: class, ControlFlags: flags, ControlClaims: len(controls),
			FalsePositiveBPS: rate, ThresholdBPS: threshold, Passed: passed,
		})
		if !passed {
			report.Violations = append(report.Violations,
				fmt.Sprintf("rule %s (%s): %d/%d control claims flagged (%d bps, budget %d bps)",
					meta.ID, class, flags, len(controls), rate, threshold))
		}
	}
	report.Passed = len(report.Violations) == 0
	writeGateArtifact(t, report)
	if !report.Passed {
		t.Fatalf("control-claim false-positive gate failed:\n%s", strings.Join(report.Violations, "\n"))
	}
}

// TestEngineOutputIndependentOfGroundTruth proves the pipeline reads only the
// production-shaped input files. Removing ground truth must not change output.
func TestEngineOutputIndependentOfGroundTruth(t *testing.T) {
	dir := writeDataset(t)
	first, err := evaluateDir(t, dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "ground_truth.csv")); err != nil {
		t.Fatal(err)
	}
	second, err := evaluateDir(t, dir)
	if err != nil {
		t.Fatalf("evaluation must not require ground truth: %v", err)
	}
	if first.OutputHash != second.OutputHash {
		t.Fatalf("ground truth affected engine output: %s != %s", first.OutputHash, second.OutputHash)
	}
}

func evaluateDir(t *testing.T, dir string) (rulesengine.BatchResult, error) {
	t.Helper()
	batch := loadGateBatch(t, dir)
	return rulesengine.Evaluate(rulesengine.EvalContext{
		TenantID:            gateTenant,
		AsOf:                mustParseTime(t, gateAsOf),
		Current:             buildSnapshots(t, batch),
		DocForensicsEnabled: true,
	}, rulesengine.CatalogVersion)
}

func hasActiveFlag(flags []rulesengine.Flag) bool {
	for _, f := range flags {
		if f.Family != "H" {
			return true
		}
	}
	return false
}

func thresholdBPS(class string) int {
	switch class {
	case "Must":
		return 0
	case "Should", "Could":
		return 500
	case "H9-H10":
		return 800
	default:
		return 0
	}
}

func passesGate(class string, flags, controls int) bool {
	rate := flags * 10000 / controls
	switch class {
	case "Must":
		return flags == 0
	case "Should", "Could":
		return rate < 500
	case "H9-H10":
		return rate < 800
	default:
		return true
	}
}

func assertNoGroundTruthColumns(t *testing.T, dir string) {
	t.Helper()
	forbidden := map[string]bool{"label": true, "typology": true, "expected_rule_ids": true, "ground_truth": true}
	for _, name := range gateInputFiles {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		header, err := csv.NewReader(f).Read()
		_ = f.Close()
		if err != nil {
			t.Fatalf("read %s header: %v", name, err)
		}
		for _, col := range header {
			if forbidden[strings.ToLower(strings.TrimSpace(col))] {
				t.Fatalf("%s exposes ground-truth column %q to the engine", name, col)
			}
		}
	}
}

func loadGateBatch(t *testing.T, dir string) contracts.Batch {
	t.Helper()
	var batch contracts.Batch
	for _, name := range gateInputFiles {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		dataset, ok := contracts.DatasetForFile(name)
		if !ok {
			t.Fatalf("no dataset mapping for %s", name)
		}
		parsed, errs := contracts.ParseFile(dataset, name, raw)
		if len(errs) > 0 {
			t.Fatalf("%s parse errors: %+v", name, errs)
		}
		batch.Claims = append(batch.Claims, parsed.Claims...)
		batch.Invoices = append(batch.Invoices, parsed.Invoices...)
		batch.Policies = append(batch.Policies, parsed.Policies...)
		batch.Providers = append(batch.Providers, parsed.Providers...)
		batch.Identities = append(batch.Identities, parsed.Identities...)
	}
	return batch
}

func buildSnapshots(t *testing.T, batch contracts.Batch) []rulesengine.ClaimSnapshot {
	t.Helper()
	policies := map[string]contracts.Policy{}
	for _, p := range batch.Policies {
		policies[p.PolicyID] = p
	}
	providers := map[string]contracts.Provider{}
	for _, p := range batch.Providers {
		providers[p.ProviderID] = p
	}
	invoices := map[string]contracts.Invoice{}
	for _, in := range batch.Invoices {
		invoices[in.ClaimID] = in
	}
	identities := map[string]map[string]string{}
	for _, ia := range batch.Identities {
		if identities[ia.ClaimantID] == nil {
			identities[ia.ClaimantID] = map[string]string{}
		}
		identities[ia.ClaimantID][ia.AttributeType] = ia.AttributeHash
	}

	out := make([]rulesengine.ClaimSnapshot, 0, len(batch.Claims))
	for _, c := range batch.Claims {
		snap := rulesengine.ClaimSnapshot{
			TenantID: gateTenant, ClaimID: c.ClaimID, BatchID: gateBatch,
			PolicyID: c.PolicyID, ClaimantID: c.ClaimantID, ProviderID: c.ProviderID,
			InvoiceNumber: c.InvoiceNumber, LossType: c.LossType, Status: c.Status,
			RepID: c.RepID, ClaimDate: c.ClaimDate.Time, IncidentDate: c.IncidentDate.Time,
			ClaimedAmount: c.ClaimedAmount, Currency: c.Currency,
			Identities: identities[c.ClaimantID],
		}
		if p, ok := policies[c.PolicyID]; ok {
			snap.Policy = &rulesengine.PolicySnapshot{
				PolicyID: p.PolicyID, Product: p.Product, LossTypeScope: p.LossTypeScope,
				Status: p.Status, Inception: p.InceptionDate.Time,
				CoverageAdded: datePtr(p.CoverageAddedDate), Expiry: p.ExpiryDate.Time,
				Reinstatement: datePtr(p.ReinstatementDate), SumInsured: p.SumInsured,
				Premium: p.Premium, AssetID: p.AssetID,
			}
		}
		if in, ok := invoices[c.ClaimID]; ok {
			snap.Invoice = &rulesengine.InvoiceSnapshot{
				Number: in.InvoiceNumber, Date: in.InvoiceDate.Time, ProviderID: in.ProviderID,
				Amount: in.InvoiceAmount, ClaimID: in.ClaimID, DocSHA256: in.DocSHA256,
			}
			snap.DocSHA256 = in.DocSHA256
		}
		if pr, ok := providers[c.ProviderID]; ok {
			snap.Provider = &rulesengine.ProviderSnapshot{
				ProviderID: pr.ProviderID, Name: pr.Name, Type: pr.Type, Region: pr.Region,
			}
		}
		out = append(out, snap)
	}
	return out
}

func readGroundTruth(t *testing.T, dir string) map[string]truthRow {
	t.Helper()
	f, err := os.Open(filepath.Join(dir, "ground_truth.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]truthRow{}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			t.Fatalf("ground truth row %d has %d columns", i+1, len(row))
		}
		label, err := strconv.Atoi(strings.TrimSpace(row[2]))
		if err != nil {
			t.Fatalf("ground truth row %d label: %v", i+1, err)
		}
		out[row[0]] = truthRow{label: label, typology: row[1], expected: splitExpected(row[3])}
	}
	return out
}

func splitExpected(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if idx := strings.IndexByte(part, ' '); idx > 0 {
			part = part[:idx]
		}
		out = append(out, part)
	}
	sort.Strings(out)
	return out
}

func datePtr(d *contracts.Date) *time.Time {
	if d == nil || d.IsZero() {
		return nil
	}
	v := d.Time
	return &v
}

func mustParseTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func hashFiles(t *testing.T, dir string, names []string) string {
	t.Helper()
	h := sha256.New()
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(h, "%s|%x\n", name, sha256.Sum256(raw))
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

func writeGateArtifact(t *testing.T, report gateReport) {
	t.Helper()
	path := os.Getenv("FP_GATE_OUT")
	if path == "" {
		path = filepath.Join(t.TempDir(), "fp-gate-report.json")
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(moduleRoot(t), path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	t.Logf("backtest artifact written to %s (passed=%t)", path, report.Passed)
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}
