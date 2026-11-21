package rulesengine

import (
	"testing"
)

func TestCrossBatchHistoryFlags(t *testing.T) {
	oldA := ClaimSnapshot{
		TenantID: "tenant-a", ClaimID: "CLM-OLD-1", BatchID: "B-OLD",
		PolicyID: "POL-OLD-1", ClaimantID: "CUST-X", ProviderID: "PRV-X", InvoiceNumber: "INV-REUSED",
		LossType: "MOTOR", Status: "SETTLED",
		ClaimDate: mustDate(t, "2026-09-10"), IncidentDate: mustDate(t, "2026-09-08"),
		ClaimedAmount: 12_000_000, Currency: "IDR",
		Invoice: testInvoice(t, "INV-REUSED", "2026-09-09", "PRV-X", 12_000_000),
		Policy:  testPolicy(t, "POL-OLD-1", "2025-01-01", "2027-01-01", 100_000_000),
	}
	oldB := oldA
	oldB.ClaimID, oldB.PolicyID = "CLM-OLD-2", "POL-OLD-2"
	oldB.ClaimDate, oldB.IncidentDate = mustDate(t, "2026-09-20"), mustDate(t, "2026-09-18")
	oldB.PriorConfirmed = true
	oldB.Invoice = testInvoice(t, "INV-OLD-2", "2026-09-19", "PRV-X", 12_000_000)
	oldB.InvoiceNumber = "INV-OLD-2"

	current := ClaimSnapshot{
		TenantID: "tenant-a", ClaimID: "CLM-NEW", BatchID: "B-NEW",
		PolicyID: "POL-NEW", ClaimantID: "CUST-X", ProviderID: "PRV-X", InvoiceNumber: "INV-REUSED",
		LossType: "MOTOR", Status: "SUBMITTED",
		ClaimDate: mustDate(t, "2026-10-01"), IncidentDate: mustDate(t, "2026-09-29"),
		ClaimedAmount: 11_000_000, Currency: "IDR",
		Invoice: testInvoice(t, "INV-REUSED", "2026-09-30", "PRV-X", 11_000_000),
		Policy:  testPolicy(t, "POL-NEW", "2025-01-01", "2027-01-01", 100_000_000),
	}

	br, err := Evaluate(EvalContext{
		TenantID: "tenant-a",
		AsOf:     mustDate(t, "2026-11-20"),
		History:  []ClaimSnapshot{oldA, oldB},
		Current:  []ClaimSnapshot{current},
	}, CatalogVersion)
	if err != nil {
		t.Fatal(err)
	}
	if len(br.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(br.Results))
	}
	flags := flagsByRule(br.Results[0])
	for _, ruleID := range []string{"A1", "C4", "G1"} {
		if _, ok := flags[ruleID]; !ok {
			t.Errorf("expected rule %s to flag the cross-batch claim; got %v", ruleID, flagIDs(br.Results[0]))
		}
	}
}

func flagIDs(cr ClaimResult) []string {
	out := make([]string, 0, len(cr.Flags))
	for _, f := range cr.Flags {
		out = append(out, f.RuleID)
	}
	return out
}
