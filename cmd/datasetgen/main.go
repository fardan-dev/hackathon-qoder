package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/pii"
)

const baseDate = "2026-09-01"

type policyOverride struct {
	product     string
	scope       string
	sumInsured  int64
	premium     int64
	inception   string
	expiry      string
	status      string
	coverageAdd string
	reinstated  string
}

type claimOptions struct {
	claimant      string
	provider      string
	incident      string
	claimDate     string
	amount        int64
	status        string
	policy        policyOverride
	invoiceAmount int64
	docHash       string
	invoiceNumber string
	identities    []identityRow
	repID         string
}

type claimRow struct {
	claimID       string
	policyID      string
	claimant      string
	provider      string
	incident      string
	claimDate     string
	amount        int64
	status        string
	invoiceNumber string
	repID         string
}

type policyRow struct {
	policyID    string
	product     string
	scope       string
	inception   string
	expiry      string
	status      string
	sumInsured  int64
	premium     int64
	coverageAdd string
	reinstated  string
	assetID     string
	agentID     string
}

type invoiceRow struct {
	invoiceNumber string
	invoiceDate   string
	provider      string
	amount        int64
	claimID       string
	docHash       string
}

type identityRow struct {
	claimant string
	kind     string
	hash     string
}

type groundTruthRow struct {
	claimID  string
	typology string
	label    int
	expected string
	note     string
}

type generator struct {
	claims     []claimRow
	policies   []policyRow
	invoices   []invoiceRow
	identities []identityRow
	truth      []groundTruthRow
	claimN     int
	policyN    int
	identityN  int
	salt       string
}

func main() {
	out := flag.String("out", "datasets", "output directory")
	flag.Parse()
	salt := os.Getenv("TENANT_SALT")
	if salt == "" {
		salt = "change-me-tenant-salt"
	}
	g := &generator{salt: salt}
	g.build()
	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}
	g.save(*out)
	fmt.Printf("dataset written to %s: %d claims, %d invoices, %d policies, %d identities, %d labels\n",
		*out, len(g.claims), len(g.invoices), len(g.policies), len(g.identities), len(g.truth))
}

func (g *generator) save(out string) {
	g.write(filepath.Join(out, "policies.csv"), func(w *csv.Writer) {
		_ = w.Write([]string{"policy_id", "product", "loss_type_scope", "inception_date", "expiry_date",
			"status", "sum_insured", "premium", "coverage_added_date", "reinstatement_date", "asset_id", "agent_id"})
		for _, p := range g.policies {
			_ = w.Write([]string{p.policyID, p.product, p.scope, p.inception, p.expiry, p.status,
				itoa(p.sumInsured), itoa(p.premium), p.coverageAdd, p.reinstated, p.assetID, p.agentID})
		}
	})
	g.write(filepath.Join(out, "providers.csv"), func(w *csv.Writer) {
		_ = w.Write([]string{"provider_id", "name", "type", "region"})
		for i := 1; i <= 18; i++ {
			_ = w.Write([]string{fmt.Sprintf("PRV-%02d", i), fmt.Sprintf("Provider Demo %02d", i),
				providerType(i), providerRegion(i)})
		}
	})
	g.write(filepath.Join(out, "claims.csv"), func(w *csv.Writer) {
		_ = w.Write([]string{"claim_id", "policy_id", "claimant_id", "claim_date", "incident_date",
			"loss_type", "claimed_amount", "currency", "provider_id", "status", "invoice_number", "rep_id"})
		for _, c := range g.claims {
			_ = w.Write([]string{c.claimID, c.policyID, c.claimant, c.claimDate, c.incident,
				g.scopeOf(c.policyID), itoa(c.amount), "IDR", c.provider, c.status, c.invoiceNumber, c.repID})
		}
	})
	g.write(filepath.Join(out, "invoices.csv"), func(w *csv.Writer) {
		_ = w.Write([]string{"invoice_number", "invoice_date", "provider_id", "invoice_amount",
			"currency", "claim_id", "doc_sha256"})
		for _, in := range g.invoices {
			_ = w.Write([]string{in.invoiceNumber, in.invoiceDate, in.provider, itoa(in.amount),
				"IDR", in.claimID, in.docHash})
		}
	})
	g.write(filepath.Join(out, "identities.csv"), func(w *csv.Writer) {
		_ = w.Write([]string{"claimant_id", "attribute_type", "attribute_hash"})
		for _, id := range g.identities {
			_ = w.Write([]string{id.claimant, id.kind, id.hash})
		}
	})
	g.write(filepath.Join(out, "ground_truth.csv"), func(w *csv.Writer) {
		_ = w.Write([]string{"claim_id", "typology", "label", "expected_rule_ids", "note"})
		for _, t := range g.truth {
			_ = w.Write([]string{t.claimID, t.typology, itoa(int64(t.label)), t.expected, t.note})
		}
	})
}

func (g *generator) build() {
	g.pair("invoice_reuse", "A1;A3;A5", "two claims reuse the same invoice number across providers; shared doc hash",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.provider = []string{"PRV-03", "PRV-04"}[i]
			c.amount = []int64{20_000_000, 22_000_000}[i]
			c.invoiceNumber = "INV-REUSE-0001"
			c.docHash = g.docHash(1)
			return c
		})
	g.pair("duplicate_claim", "A2", "same claimant, incident date, amount and provider in one batch",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.claimant = "CLMNT-DUP"
			c.provider = "PRV-05"
			c.incident = "2026-09-12"
			c.claimDate = "2026-09-14"
			c.amount = 15_000_000
			return c
		})
	g.pair("amount_inflation", "A4;B1", "invoice amount materially below the claimed amount",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.provider = "PRV-01"
			c.amount = []int64{30_000_000, 25_000_000}[i]
			c.invoiceAmount = c.amount * 60 / 100
			return c
		})
	g.pair("document_amount_edit", "H1;H6 (gated off)", "document forensics typology; structured invoice only reflects part of the amount",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.provider = "PRV-06"
			c.policy = propertyPolicy()
			c.amount = []int64{60_000_000, 54_000_000}[i]
			c.invoiceAmount = c.amount * 75 / 100
			c.docHash = g.docHash(10 + i)
			return c
		})
	g.pair("document_forgery", "H2;H3;H4;H7 (gated off)", "document forensics typology; metadata analyzer not in thin slice",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.provider = "PRV-07"
			c.policy = propertyPolicy()
			c.amount = []int64{40_000_000, 45_000_000}[i]
			c.docHash = g.docHash(20 + i)
			return c
		})
	g.pair("document_rescan", "H5 (gated off); A3 fallback", "identical document hash reused under a different invoice number",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.provider = "PRV-08"
			c.policy = propertyPolicy()
			c.amount = []int64{35_000_000, 38_000_000}[i]
			c.docHash = g.docHash(30)
			return c
		})
	g.pair("structuring", "B3;B4", "amounts sit in the 85-100% band of the MOTOR high-amount threshold and are round",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.provider = "PRV-01"
			c.amount = []int64{45_000_000, 49_000_000}[i]
			return c
		})
	g.pair("cover_breach", "B5;E2", "first claim exceeds sum insured; second policy expired before incident",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.provider = "PRV-02"
			if i == 0 {
				c.policy.sumInsured = 20_000_000
				c.amount = 35_000_000
			} else {
				c.policy.expiry = "2026-08-31"
				c.policy.status = "EXPIRED"
				c.incident = "2026-10-15"
				c.claimDate = "2026-10-18"
				c.amount = 28_000_000
			}
			return c
		})
	g.pair("new_policy_reinstatement", "C1;C2;E1", "one claim right after inception, one after reinstatement",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.provider = "PRV-03"
			if i == 0 {
				c.policy.inception = "2026-10-01"
				c.policy.coverageAdd = "2026-10-01"
				c.incident = "2026-10-10"
				c.claimDate = "2026-10-12"
				c.amount = 12_000_000
			} else {
				c.policy.reinstated = "2026-09-20"
				c.incident = "2026-10-05"
				c.claimDate = "2026-10-07"
				c.amount = 14_000_000
			}
			return c
		})
	g.pair("repeat_claimant", "C4;F1;G3", "same claimant files twice inside the 90-day window",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.claimant = "CLMNT-REP"
			c.provider = "PRV-04"
			c.incident = []string{"2026-09-05", "2026-09-25"}[i]
			c.claimDate = []string{"2026-09-07", "2026-09-27"}[i]
			c.amount = []int64{8_250_000, 9_750_000}[i]
			return c
		})
	g.pair("provider_overbilling", "D4", "provider PRV-01 carries a disproportionate share of claims",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.provider = "PRV-01"
			c.amount = []int64{31_750_000, 33_250_000}[i]
			return c
		})
	for k := 0; k < 4; k++ {
		c := g.defaults(k)
		c.provider = "PRV-01"
		c.amount = []int64{28_250_000, 30_750_000, 33_500_000, 36_250_000}[k]
		c.incident = date(baseDate, 45+k*3)
		c.claimDate = date(c.incident, 2)
		g.add("provider_overbilling", 1, "D4", "provider PRV-01 carries a disproportionate share of claims", c)
	}
	g.pair("collusion_network", "D1;D2;D6", "two claimants share phone and bank account and use the same provider",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.claimant = []string{"CLMNT-COL-A", "CLMNT-COL-B"}[i]
			c.provider = "PRV-02"
			c.incident = []string{"2026-09-18", "2026-09-22"}[i]
			c.claimDate = []string{"2026-09-20", "2026-09-24"}[i]
			c.amount = []int64{18_400_000, 19_600_000}[i]
			c.identities = []identityRow{{
				claimant: c.claimant, kind: "BANK_ACCOUNT",
				hash: pii.Hash(g.salt, "BANK-SHARED-0001"),
			}}
			return c
		})
	offender := func(i int) claimOptions {
		c := g.defaults(i)
		c.claimant = "CLMNT-OFF"
		c.provider = "PRV-03"
		c.incident = date(baseDate, 40+i*4)
		c.claimDate = date(baseDate, 42+i*4)
		c.amount = []int64{22_500_000, 24_500_000, 26_500_000}[i]
		return c
	}
	g.add("repeat_offender", 1, "G1;G3", "claimant has cross-policy claims inside 12 months", offender(0))
	g.add("repeat_offender", 1, "G1;G3", "claimant has cross-policy claims inside 12 months", offender(1))
	g.add("repeat_offender", 1, "G1;G3", "third policy for the same claimant", offender(2))
	g.pair("high_quality_forgery", "H8;H9;H10 (gated off)", "document forensics typology; ELA/template checks gated off",
		func(i int) claimOptions {
			c := g.defaults(i)
			c.provider = "PRV-09"
			c.policy = propertyPolicy()
			c.policy.sumInsured = 200_000_000
			c.amount = []int64{55_000_000, 58_000_000}[i]
			c.docHash = g.docHash(40 + i)
			return c
		})

	controlAmounts := []int64{8_250_000, 11_750_000, 13_400_000, 7_950_000, 16_300_000,
		9_850_000, 12_250_000, 14_750_000, 6_950_000, 17_450_000}
	for i, amount := range controlAmounts {
		c := g.defaults(i)
		c.claimant = fmt.Sprintf("CLMNT-CLEAN-%02d", i+1)
		c.provider = fmt.Sprintf("PRV-%02d", 5+i)
		c.amount = amount
		if i%2 == 1 {
			c.policy = propertyPolicy()
		}
		g.add("control", 0, "", "clean control claim", c)
	}
	// Extra clean controls lift each line of business to the 10-provider
	// population that provider-level rules require, without touching thresholds.
	extras := []struct {
		amount   int64
		provider string
		property bool
	}{
		{9_250_000, "PRV-15", false},
		{10_450_000, "PRV-16", true},
		{11_650_000, "PRV-17", true},
		{12_850_000, "PRV-18", true},
	}
	for j, extra := range extras {
		c := g.defaults(10 + j)
		c.claimant = fmt.Sprintf("CLMNT-CLEAN-%02d", 11+j)
		c.provider = extra.provider
		c.amount = extra.amount
		if extra.property {
			c.policy = propertyPolicy()
		}
		g.add("control", 0, "", "clean control claim", c)
	}
	for i := range g.claims {
		g.claims[i].claimID = fmt.Sprintf("CLM-%04d", 1001+i)
	}
	for i := range g.invoices {
		g.invoices[i].claimID = g.claims[i].claimID
	}
	for i := range g.truth {
		g.truth[i].claimID = g.claims[i].claimID
	}
	for i := range g.policies {
		g.policies[i].assetID = fmt.Sprintf("AST-%04d", 5001+i)
	}
}

func (g *generator) pair(typology, expected, note string, opts func(i int) claimOptions) {
	for i := 0; i < 2; i++ {
		g.add(typology, 1, expected, note, opts(i))
	}
}

func (g *generator) defaults(i int) claimOptions {
	incident := date(baseDate, 20+i*3)
	return claimOptions{
		claimant:  fmt.Sprintf("CLMNT-%04d", 1+g.claimN),
		provider:  fmt.Sprintf("PRV-%02d", (i%4)+1),
		incident:  incident,
		claimDate: date(incident, 2),
		amount:    10_000_000,
		status:    "OPEN",
		policy: policyOverride{
			product: "MOTOR-COMPREHENSIVE", scope: "MOTOR", sumInsured: 100_000_000, premium: 3_000_000,
			inception: "2025-01-01", expiry: "2026-12-31", status: "ACTIVE",
		},
	}
}

func (g *generator) add(typology string, label int, expected, note string, c claimOptions) {
	g.claimN++
	g.policyN++
	claimID := fmt.Sprintf("CLM-%04d", 1000+g.claimN)
	policyID := fmt.Sprintf("POL-%04d", 5000+g.policyN)
	invoiceNumber := c.invoiceNumber
	if invoiceNumber == "" {
		invoiceNumber = fmt.Sprintf("INV-%04d", 7000+g.claimN)
	}
	invoiceAmount := c.invoiceAmount
	if invoiceAmount == 0 {
		invoiceAmount = c.amount
	}
	p := c.policy
	if p.product == "" {
		p = policyOverride{product: "MOTOR-COMPREHENSIVE", scope: "MOTOR", sumInsured: 100_000_000,
			premium: 3_000_000, inception: "2025-01-01", expiry: "2026-12-31", status: "ACTIVE"}
	}
	if p.premium == 0 {
		p.premium = 3_000_000
	}
	g.policies = append(g.policies, policyRow{
		policyID: policyID, product: p.product, scope: p.scope, inception: p.inception,
		expiry: p.expiry, status: p.status, sumInsured: p.sumInsured, premium: p.premium,
		coverageAdd: p.coverageAdd, reinstated: p.reinstated, agentID: fmt.Sprintf("AGT-%03d", 1+g.claimN%20),
	})
	g.claims = append(g.claims, claimRow{
		claimID: claimID, policyID: policyID, claimant: c.claimant, provider: c.provider,
		incident: c.incident, claimDate: c.claimDate, amount: c.amount, status: c.status,
		invoiceNumber: invoiceNumber, repID: c.repID,
	})
	g.invoices = append(g.invoices, invoiceRow{
		invoiceNumber: invoiceNumber, invoiceDate: c.claimDate, provider: c.provider,
		amount: invoiceAmount, claimID: claimID, docHash: c.docHash,
	})
	g.identityN++
	if !g.hasIdentity(c.claimant, "PHONE") {
		g.identities = append(g.identities, identityRow{
			claimant: c.claimant, kind: "PHONE",
			hash: pii.Hash(g.salt, fmt.Sprintf("0812%08d", 10000000+g.identityN)),
		})
	}
	for _, extra := range c.identities {
		if !g.hasIdentity(extra.claimant, extra.kind) {
			g.identities = append(g.identities, extra)
		}
	}
	g.truth = append(g.truth, groundTruthRow{
		claimID: claimID, typology: typology, label: label, expected: expected, note: note,
	})
}

func (g *generator) hasIdentity(claimant, kind string) bool {
	for _, id := range g.identities {
		if id.claimant == claimant && id.kind == kind {
			return true
		}
	}
	return false
}

func (g *generator) scopeOf(policyID string) string {
	for _, p := range g.policies {
		if p.policyID == policyID {
			return p.scope
		}
	}
	return "MOTOR"
}

func (g *generator) docHash(n int) string {
	return fmt.Sprintf("sha256:%064x", n)
}

func (g *generator) write(path string, fn func(*csv.Writer)) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	fn(w)
	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatalf("write %s: %v", path, err)
	}
}

func propertyPolicy() policyOverride {
	return policyOverride{product: "PROPERTY-ALLRISK", scope: "PROPERTY", sumInsured: 150_000_000,
		premium: 4_500_000, inception: "2025-01-01", expiry: "2026-12-31", status: "ACTIVE"}
}

func providerType(i int) string {
	types := []string{"WORKSHOP", "GARAGE", "SURVEYOR", "HOSPITAL"}
	return types[(i-1)%len(types)]
}

func providerRegion(i int) string {
	regions := []string{"DKI JAKARTA", "JAWA BARAT", "JAWA TIMUR", "SUMATERA UTARA"}
	return regions[(i-1)%len(regions)]
}

func date(base string, addDays int) string {
	t, err := time.Parse("2006-01-02", base)
	if err != nil {
		panic(err)
	}
	return t.AddDate(0, 0, addDays).Format("2006-01-02")
}

func itoa(v int64) string {
	return fmt.Sprintf("%d", v)
}
