package rulesengine

const (
	TypologyInvoiceReuse        = "invoice_reuse"
	TypologyDuplicateClaim      = "duplicate_claim"
	TypologyAmountInflation     = "amount_inflation"
	TypologyDocumentAmountEdit  = "document_amount_edit"
	TypologyDocumentForgery     = "document_forgery"
	TypologyDocumentRescan      = "document_rescan"
	TypologyStructuring         = "structuring"
	TypologyCoverBreach         = "cover_breach"
	TypologyNewPolicy           = "new_policy_reinstatement"
	TypologyRepeatClaimant      = "repeat_claimant"
	TypologyProviderOverbilling = "provider_overbilling"
	TypologyCollusion           = "collusion_network"
	TypologyRepeatOffender      = "repeat_offender"
	TypologyHighQualityForgery  = "high_quality_forgery"
)

func m(id string, family, name string, sev Severity, params map[string]any, status RuleStatus, owner string, typ []string, reason string) RuleMeta {
	return RuleMeta{
		ID: id, Version: 1, Family: family, Name: name, Severity: sev,
		Params: params, ParamsVersion: 1, Status: status, Owner: owner,
		Typologies: typ, ReasonCode: reason,
	}
}

func catalogRules() []rule {
	owner := "fraud-analytics"
	docOwner := "document-forensics"
	active := StatusActive
	gated := StatusGatedOff
	return []rule{
		{m("A1", "A", "Reused Invoice", SeverityHigh, map[string]any{}, active, owner, []string{TypologyInvoiceReuse}, "CF_INVOICE_REUSE"), evalA1},
		{m("A2", "A", "Duplicate Claim Fingerprint", SeverityHigh, map[string]any{}, active, owner, []string{TypologyDuplicateClaim}, "CF_DUPLICATE_CLAIM"), evalA2},
		{m("A3", "A", "Reused Document Hash", SeverityHigh, map[string]any{}, active, owner, []string{TypologyInvoiceReuse}, "CF_INVOICE_REUSE"), evalA3},
		{m("A4", "A", "Invoice-Claim Amount Mismatch", SeverityMedium, map[string]any{"tolerance": 0}, active, owner, []string{TypologyAmountInflation}, "CF_AMOUNT_INFLATION"), evalA4},
		{m("A5", "A", "Invoice Date Anomaly", SeverityLow, map[string]any{"max_days_after_claim": 30}, active, owner, []string{TypologyInvoiceReuse}, "CF_INVOICE_REUSE"), evalA5},
		{m("A6", "A", "Invoice Identity Conflict", SeverityHigh, map[string]any{}, active, owner, []string{TypologyInvoiceReuse}, "CF_INVOICE_REUSE"), evalA6},

		{m("B1", "B", "High Amount", SeverityMedium, map[string]any{"MOTOR_limit": 50_000_000, "PROPERTY_limit": 100_000_000}, active, owner, []string{TypologyAmountInflation}, "CF_AMOUNT_INFLATION"), evalB1},
		{m("B2", "B", "Batch Outlier", SeverityMedium, map[string]any{"multiplier": 3, "min_population": 30}, active, owner, []string{TypologyAmountInflation}, "CF_AMOUNT_INFLATION"), evalB2},
		{m("B3", "B", "Structured Below-Threshold", SeverityMedium, map[string]any{"lower_pct": 85, "upper_pct": 100}, active, owner, []string{TypologyStructuring}, "CF_AMOUNT_INFLATION"), evalB3},
		{m("B4", "B", "Round Number", SeverityLow, map[string]any{"min_amount": 5_000_000, "multiple": 1_000_000}, active, owner, []string{TypologyStructuring}, "CF_AMOUNT_INFLATION"), evalB4},
		{m("B5", "B", "Coverage Limit Breach", SeverityHigh, map[string]any{}, active, owner, []string{TypologyCoverBreach}, "CF_COVER_BREACH"), evalB5},
		{m("B6", "B", "Aggregate Near Limit", SeverityMedium, map[string]any{"pct_limit": 80, "months": 12}, active, owner, []string{TypologyCoverBreach}, "CF_COVER_BREACH"), evalB6},

		{m("C1", "C", "Early Claim", SeverityMedium, map[string]any{"max_days": 30}, active, owner, []string{TypologyNewPolicy}, "CF_COVER_BREACH"), evalC1},
		{m("C2", "C", "Post-Reinstatement Claim", SeverityHigh, map[string]any{"max_days": 30}, active, owner, []string{TypologyNewPolicy}, "CF_COVER_BREACH"), evalC2},
		{m("C3", "C", "Late Reporting", SeverityLow, map[string]any{"max_days": 30}, active, owner, []string{TypologyDuplicateClaim}, "CF_DUPLICATE_CLAIM"), evalC3},
		{m("C4", "C", "Repeat Claims Window", SeverityMedium, map[string]any{"window_days": 90, "min_claims": 2}, active, owner, []string{TypologyRepeatClaimant}, "CF_DUPLICATE_CLAIM"), evalC4},
		{m("C5", "C", "Resubmission After Rejection", SeverityHigh, map[string]any{"max_days": 180}, active, owner, []string{TypologyDuplicateClaim}, "CF_DUPLICATE_CLAIM"), evalC5},

		{m("D1", "D", "Shared Bank Account", SeverityHigh, map[string]any{}, active, owner, []string{TypologyCollusion}, "CF_NETWORK_COLLUSION"), evalD1},
		{m("D2", "D", "Shared Contact Identity", SeverityMedium, map[string]any{}, active, owner, []string{TypologyCollusion}, "CF_NETWORK_COLLUSION"), evalD2},
		{m("D3", "D", "Claimant-Provider Concentration", SeverityMedium, map[string]any{"min_claims": 3, "window_days": 180}, active, owner, []string{TypologyCollusion}, "CF_NETWORK_COLLUSION"), evalD3},
		{m("D4", "D", "Provider Volume Anomaly", SeverityMedium, map[string]any{"multiplier": 3, "min_providers": 10}, active, owner, []string{TypologyProviderOverbilling}, "CF_PROVIDER_OVERBILLING"), evalD4},
		{m("D5", "D", "Provider Billing Outlier", SeverityMedium, map[string]any{"multiplier": 3, "min_providers": 10}, active, owner, []string{TypologyProviderOverbilling}, "CF_PROVIDER_OVERBILLING"), evalD5},
		{m("D6", "D", "Collusion Ring", SeverityHigh, map[string]any{"min_claims": 3, "window_days": 180}, active, owner, []string{TypologyCollusion}, "CF_NETWORK_COLLUSION"), evalD6},
		{m("D7", "D", "Same Asset Multi-Policy", SeverityHigh, map[string]any{}, active, owner, []string{TypologyCollusion}, "CF_NETWORK_COLLUSION"), evalD7},

		{m("E1", "E", "Claim on Recently Added Coverage", SeverityMedium, map[string]any{"max_days": 30}, active, owner, []string{TypologyNewPolicy}, "CF_COVER_BREACH"), evalE1},
		{m("E2", "E", "Policy Not Active at Incident", SeverityHigh, map[string]any{}, active, owner, []string{TypologyCoverBreach}, "CF_COVER_BREACH"), evalE2},

		{m("F1", "F", "Submission Velocity", SeverityMedium, map[string]any{"min_claims": 2, "window_days": 90}, active, owner, []string{TypologyRepeatClaimant}, "CF_DUPLICATE_CLAIM"), evalF1},
		{m("F2", "F", "Withdraw-Resubmit Cycle", SeverityMedium, map[string]any{"max_days": 90}, active, owner, []string{TypologyDuplicateClaim}, "CF_DUPLICATE_CLAIM"), evalF2},
		{m("F3", "F", "Repeated NEEDS_INFO Stall", SeverityLow, map[string]any{"min_cycles": 2}, active, owner, []string{TypologyRepeatClaimant}, "CF_DUPLICATE_CLAIM"), evalF3},
		{m("F4", "F", "Representative Concentration", SeverityMedium, map[string]any{"min_claimants": 5, "window_days": 180}, active, owner, []string{TypologyCollusion}, "CF_NETWORK_COLLUSION"), evalF4},

		{m("G1", "G", "Prior Confirmed Flag", SeverityHigh, map[string]any{"months": 24}, active, owner, []string{TypologyRepeatOffender}, "CF_REPEAT_OFFENDER"), evalG1},
		{m("G2", "G", "Prior Cleared Suppression", SeverityLow, map[string]any{"months": 12}, active, owner, []string{TypologyRepeatOffender}, "CF_REPEAT_OFFENDER"), evalG2},
		{m("G3", "G", "Cross-Policy Claim History", SeverityMedium, map[string]any{"min_policies": 3, "months": 12}, active, owner, []string{TypologyRepeatClaimant}, "CF_REPEAT_OFFENDER"), evalG3},

		{m("H1", "H", "OCR-Field Mismatch", SeverityHigh, map[string]any{"amount_tolerance": 0, "string_similarity": 0.90}, gated, docOwner, []string{TypologyDocumentAmountEdit}, "CF_AMOUNT_INFLATION"), evalH},
		{m("H2", "H", "Metadata Tampering Trace", SeverityHigh, map[string]any{"software_list": []string{"photoshop", "gimp", "canva"}}, gated, docOwner, []string{TypologyDocumentForgery}, "CF_DOCUMENT_EDIT"), evalH},
		{m("H3", "H", "Metadata Timeline Contradiction", SeverityHigh, map[string]any{}, gated, docOwner, []string{TypologyDocumentForgery}, "CF_DOCUMENT_EDIT"), evalH},
		{m("H4", "H", "Re-save / Double Compression", SeverityMedium, map[string]any{}, gated, docOwner, []string{TypologyDocumentForgery}, "CF_DOCUMENT_EDIT"), evalH},
		{m("H5", "H", "Near-Duplicate Document", SeverityHigh, map[string]any{"hamming_max": 16, "similarity": 0.98}, gated, docOwner, []string{TypologyDocumentRescan, TypologyDocumentForgery}, "CF_DOCUMENT_EDIT"), evalH},
		{m("H6", "H", "OCR Arithmetic Mismatch", SeverityMedium, map[string]any{"tolerance_pct": 1}, gated, docOwner, []string{TypologyDocumentAmountEdit}, "CF_AMOUNT_INFLATION"), evalH},
		{m("H7", "H", "OCR Quality Failure (routing)", SeverityMedium, map[string]any{"confidence": 0.60}, gated, docOwner, []string{TypologyDocumentForgery}, "CF_DOCUMENT_EDIT"), evalH},
		{m("H8", "H", "Rasterized / Overlaid Text", SeverityMedium, map[string]any{}, gated, docOwner, []string{TypologyHighQualityForgery}, "CF_DOCUMENT_EDIT"), evalH},
		{m("H9", "H", "ELA Local Anomaly", SeverityLow, map[string]any{}, gated, docOwner, []string{TypologyHighQualityForgery}, "CF_DOCUMENT_EDIT"), evalH},
		{m("H10", "H", "Template Deviation", SeverityLow, map[string]any{"similarity": 0.80}, gated, docOwner, []string{TypologyHighQualityForgery}, "CF_DOCUMENT_EDIT"), evalH},
	}
}

// Catalog returns a copy of the 43 rule metadata entries.
func Catalog() []RuleMeta {
	rules := catalogRules()
	out := make([]RuleMeta, 0, len(rules))
	for _, r := range rules {
		out = append(out, r.meta)
	}
	return out
}

func CatalogByFamily(family string) []RuleMeta {
	var out []RuleMeta
	for _, m := range Catalog() {
		if m.Family == family {
			out = append(out, m)
		}
	}
	return out
}
